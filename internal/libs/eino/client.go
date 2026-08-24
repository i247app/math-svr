package eino

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/cloudwego/eino-ext/components/model/claude"
	"github.com/cloudwego/eino-ext/components/model/gemini"
	einoollama "github.com/cloudwego/eino-ext/components/model/ollama"
	einoopenai "github.com/cloudwego/eino-ext/components/model/openai"
	einomodel "github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/eino-contrib/jsonschema"
	"google.golang.org/genai"

	"math-ai.com/math-ai/internal/infrastructure/logger"
)

// Client is the eino-backed LLM client. It wraps the eino-ext ChatModel
// selected by Config.Backend and exposes the same narrow chat surface as
// libs/langchain.Client, minus embeddings (the eino provider reports
// embeddings as unsupported at the adapter layer).
//
// JSON mode: eino fixes the response format at model construction, not
// per call. NewClient therefore builds up to two ChatModel instances —
// a plain one and a JSON-mode one — and Generate/GenerateStream route on
// ChatRequest.JSONMode:
//
//	googleai  → single instance; JSON requested per call via
//	            gemini.WithResponseJSONSchema (permissive object schema).
//	openai    → second instance with ResponseFormat json_object.
//	ollama    → second instance with Format "json".
//	anthropic → no native JSON mode; the flag is silently ignored,
//	            matching the langchain provider's documented posture.
//
// The client is safe for concurrent use; the underlying vendor SDK HTTP
// clients pool connections internally. Nothing to Close — constructed
// once at boot and retained on the bot adapter for the process lifetime.
type Client struct {
	cfg Config

	// plain serves requests with JSONMode=false.
	plain einomodel.BaseChatModel

	// jsonModel serves requests with JSONMode=true. Never nil: equal to
	// plain when JSON mode is handled per-call (googleai) or unsupported
	// (anthropic).
	jsonModel einomodel.BaseChatModel

	// jsonCallOpts are extra per-call options appended when JSONMode=true.
	// Non-nil only for backends that request JSON per call (googleai).
	jsonCallOpts []einomodel.Option
}

// Backend returns the backend the client was constructed with.
func (c *Client) Backend() Backend { return c.cfg.Backend }

// Model returns the configured default chat model id.
func (c *Client) Model() string { return c.cfg.Model }

// NewClient validates cfg, builds the configured eino-ext backend, and
// (when cfg.RequireAtBoot is true) runs a tiny probe call to fail fast on
// bad credentials.
//
// Errors:
//   - cfg.Validate failure → wraps ErrInvalidConfig (adapter factory maps
//     to MathError(BOT_CONFIG_INVALID)).
//   - backend construction failure → wraps the underlying error (adapter
//     factory maps to MathError(BOT_CONNECT_FAILED)).
//   - probe failure under RequireAtBoot=true → wraps the underlying error
//     (adapter factory maps to MathError(BOT_CONNECT_FAILED)).
func NewClient(ctx context.Context, cfg Config) (*Client, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	cfg.applyDefaults()

	plain, jsonModel, jsonCallOpts, err := newModels(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("eino: build backend %q: %w", cfg.Backend, err)
	}

	c := &Client{cfg: cfg, plain: plain, jsonModel: jsonModel, jsonCallOpts: jsonCallOpts}

	if cfg.RequireAtBoot {
		if err := c.probe(ctx); err != nil {
			return nil, fmt.Errorf("eino: boot probe failed: %w", err)
		}
	} else {
		logger.From(ctx).Infof("eino: client ready (backend=%s, model=%s, probe-skipped)",
			cfg.Backend, cfg.Model)
	}

	return c, nil
}

// newModels selects and constructs the concrete eino-ext backend. It
// returns the plain-mode model, the JSON-mode model (== plain when JSON
// is per-call or unsupported) and any per-call JSON options.
//
// Adding a new backend = one new case here plus the matching import.
func newModels(ctx context.Context, cfg Config) (einomodel.BaseChatModel, einomodel.BaseChatModel, []einomodel.Option, error) {
	switch cfg.Backend {
	case BackendGoogleAI:
		clientCfg := &genai.ClientConfig{
			APIKey:  cfg.APIKey,
			Backend: genai.BackendGeminiAPI,
		}
		if cfg.BaseURL != "" {
			clientCfg.HTTPOptions.BaseURL = cfg.BaseURL
		}
		cli, err := genai.NewClient(ctx, clientCfg)
		if err != nil {
			return nil, nil, nil, err
		}
		m, err := gemini.NewChatModel(ctx, &gemini.Config{
			Client: cli,
			Model:  cfg.Model,
		})
		if err != nil {
			return nil, nil, nil, err
		}
		// JSON mode per call: a permissive object schema flips the Gemini
		// response MIME type to application/json without constraining the
		// payload shape (quiz/exercise prompts define the real contract).
		jsonOpts := []einomodel.Option{
			gemini.WithResponseJSONSchema(&jsonschema.Schema{Type: "object"}),
		}
		return m, m, jsonOpts, nil

	case BackendOpenAI:
		build := func(jsonMode bool) (einomodel.BaseChatModel, error) {
			mc := &einoopenai.ChatModelConfig{
				APIKey:      cfg.APIKey,
				BaseURL:     cfg.BaseURL,
				Model:       cfg.Model,
				Timeout:     cfg.Timeout,
				ExtraFields: openAIExtraFields(cfg),
			}
			if jsonMode {
				mc.ResponseFormat = &einoopenai.ChatCompletionResponseFormat{
					Type: einoopenai.ChatCompletionResponseFormatTypeJSONObject,
				}
			}
			return einoopenai.NewChatModel(ctx, mc)
		}
		plain, err := build(false)
		if err != nil {
			return nil, nil, nil, err
		}
		jsonModel, err := build(true)
		if err != nil {
			return nil, nil, nil, err
		}
		return plain, jsonModel, nil, nil

	case BackendAnthropic:
		maxTokens := cfg.MaxTokens
		if maxTokens <= 0 {
			maxTokens = defaultAnthropicMaxTokens
		}
		claudeCfg := &claude.Config{
			APIKey:    cfg.APIKey,
			Model:     cfg.Model,
			MaxTokens: maxTokens,
		}
		if cfg.BaseURL != "" {
			baseURL := cfg.BaseURL
			claudeCfg.BaseURL = &baseURL
		}
		m, err := claude.NewChatModel(ctx, claudeCfg)
		if err != nil {
			return nil, nil, nil, err
		}
		// Anthropic has no native JSON response format; JSONMode is
		// silently ignored (prompt-level contracts still apply).
		return m, m, nil, nil

	case BackendOllama:
		build := func(jsonMode bool) (einomodel.BaseChatModel, error) {
			mc := &einoollama.ChatModelConfig{
				BaseURL: cfg.BaseURL,
				Model:   cfg.Model,
				Timeout: cfg.Timeout,
			}
			if jsonMode {
				mc.Format = json.RawMessage(`"json"`)
			}
			return einoollama.NewChatModel(ctx, mc)
		}
		plain, err := build(false)
		if err != nil {
			return nil, nil, nil, err
		}
		jsonModel, err := build(true)
		if err != nil {
			return nil, nil, nil, err
		}
		return plain, jsonModel, nil, nil

	default:
		return nil, nil, nil, fmt.Errorf("%w: unknown backend %q", ErrInvalidConfig, cfg.Backend)
	}
}

// openAIExtraFields builds the passthrough map for chat-completions
// parameters that eino-ext's ChatModelConfig does not surface as typed
// fields. eino-ext merges the map into the request JSON verbatim (see
// go-openai internal/marshaller.go), which is the only route to `store`.
//
// Returns nil when there is nothing to add, so the request body stays
// byte-identical to before for every deploy that leaves Store off.
func openAIExtraFields(cfg Config) map[string]any {
	if !cfg.Store {
		return nil
	}
	return map[string]any{"store": true}
}

// probe makes a one-token Generate call to verify credentials and
// connectivity. Used only when RequireAtBoot=true.
func (c *Client) probe(ctx context.Context) error {
	_, err := c.Generate(ctx, ChatRequest{
		Model:     c.cfg.Model,
		Messages:  []Message{{Role: RoleUser, Content: "ping"}},
		MaxTokens: 1,
	})
	return err
}

// pick resolves the model instance and per-call options for req,
// honouring JSONMode routing.
func (c *Client) pick(req ChatRequest) (einomodel.BaseChatModel, []einomodel.Option) {
	opts := c.buildCallOptions(req)
	if req.JSONMode {
		return c.jsonModel, append(opts, c.jsonCallOpts...)
	}
	return c.plain, opts
}

// Generate runs a non-streamed chat completion under the configured
// timeout and retry budget. Returns ErrInvalidConfig for an empty
// message list; for other failures see translateError.
func (c *Client) Generate(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	log := logger.From(ctx)

	if len(req.Messages) == 0 {
		return nil, fmt.Errorf("%w: at least one message is required", ErrInvalidConfig)
	}

	callCtx, cancel := context.WithTimeout(ctx, c.cfg.Timeout)
	defer cancel()

	msgs := toEinoMessages(req.Messages)
	m, opts := c.pick(req)

	// Attach an httptrace so we can log whether this outbound call opened a
	// fresh connection (TLS handshake) or reused a pooled keep-alive one.
	traceCtx, ct := withConnTrace(callCtx)

	out, err := c.withRetry(traceCtx, func(retryCtx context.Context) (*schema.Message, error) {
		return m.Generate(retryCtx, msgs, opts...)
	})
	log.Infof("langchain.conn captured=%t reused=%t was_idle=%t idle_ms=%d tls_handshake_ms=%d remote=%s",
		ct.gotConn, ct.reused, ct.wasIdle, ct.idleMs, ct.tlsDoneMs, ct.remote)

	if err != nil {
		return nil, c.translateError(err)
	}
	if out == nil {
		return nil, ErrDecodeResponse
	}

	resp := &ChatResponse{
		Model:   firstNonEmpty(req.Model, c.cfg.Model),
		Content: out.Content,
	}
	if out.ResponseMeta != nil {
		resp.FinishReason = out.ResponseMeta.FinishReason
		resp.Usage = fromTokenUsage(out.ResponseMeta.Usage)
	}

	// Operator metadata only — never the prompt or the response body.
	logger.From(ctx).Infof("eino.generate backend=%s model=%s finish=%s prompt_tokens=%d completion_tokens=%d",
		c.cfg.Backend, resp.Model, resp.FinishReason,
		resp.Usage.PromptTokens, resp.Usage.CompletionTokens)
	return resp, nil
}

// GenerateStream runs a streaming chat completion. Each token chunk is
// pushed to onChunk; the final invocation receives Done=true plus
// optional Usage info. onChunk MUST NOT block — the caller's pipeline
// owns downstream buffering.
//
// Streaming is not retried automatically: a mid-stream failure would
// require replaying delivered tokens, which is the caller's policy
// decision — same contract as libs/langchain.
func (c *Client) GenerateStream(ctx context.Context, req ChatRequest, onChunk func(StreamChunk) error) (*ChatResponse, error) {
	if len(req.Messages) == 0 {
		return nil, fmt.Errorf("%w: at least one message is required", ErrInvalidConfig)
	}
	if onChunk == nil {
		return nil, fmt.Errorf("%w: onChunk is required for streaming", ErrInvalidConfig)
	}

	callCtx, cancel := context.WithTimeout(ctx, c.cfg.Timeout)
	defer cancel()

	msgs := toEinoMessages(req.Messages)
	m, opts := c.pick(req)

	sr, err := m.Stream(callCtx, msgs, opts...)
	if err != nil {
		return nil, c.translateError(err)
	}
	defer sr.Close()

	var chunks []*schema.Message
	for {
		chunk, recvErr := sr.Recv()
		if errors.Is(recvErr, io.EOF) {
			break
		}
		if recvErr != nil {
			return nil, c.translateError(recvErr)
		}
		chunks = append(chunks, chunk)
		if chunk.Content != "" {
			if cbErr := onChunk(StreamChunk{Delta: chunk.Content}); cbErr != nil {
				return nil, cbErr
			}
		}
	}

	out := &ChatResponse{Model: firstNonEmpty(req.Model, c.cfg.Model)}
	final := StreamChunk{Done: true}
	if len(chunks) > 0 {
		full, cErr := schema.ConcatMessages(chunks)
		if cErr != nil {
			return nil, fmt.Errorf("%w: %v", ErrDecodeResponse, cErr)
		}
		out.Content = full.Content
		if full.ResponseMeta != nil {
			out.FinishReason = full.ResponseMeta.FinishReason
			out.Usage = fromTokenUsage(full.ResponseMeta.Usage)
		}
		if out.Usage != (Usage{}) {
			usage := out.Usage
			final.Usage = &usage
		}
	}
	if cbErr := onChunk(final); cbErr != nil {
		return out, cbErr
	}
	return out, nil
}

// withRetry executes op under the client's MaxRetries / RetryDelay
// budget. Errors that classify as auth / config / rate-limited / context
// length are NOT retried — they will not recover within the same call.
// Transport-shaped errors and retryable HTTP 5xx are retried with a
// linear sleep. Same policy as libs/langchain.
func (c *Client) withRetry(ctx context.Context, op func(ctx context.Context) (*schema.Message, error)) (*schema.Message, error) {
	var lastErr error
	for attempt := 0; attempt <= c.cfg.MaxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(c.cfg.RetryDelay):
			}
		}

		out, err := op(ctx)
		if err == nil {
			return out, nil
		}

		lastErr = err
		if !c.shouldRetry(err) {
			return nil, err
		}
	}
	return nil, lastErr
}

// shouldRetry classifies an error from the underlying eino model call.
// Non-recoverable: auth, config, context-too-large, rate-limit, ctx
// cancellation. Everything else (transport, 5xx, unclassified) is
// considered transient. Vendor errors are lifted first so typed HTTP
// statuses participate in the decision.
func (c *Client) shouldRetry(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}
	if errors.Is(err, ErrInvalidConfig) || errors.Is(err, ErrContextTooLarge) ||
		errors.Is(err, ErrUnsupportedOp) || errors.Is(err, ErrRateLimited) {
		return false
	}

	classified := err
	if lifted := liftVendorError(c.cfg.Backend, err); lifted != nil {
		classified = lifted
	}
	if IsAuthError(classified) || IsConfigError(classified) || IsRateLimited(classified) {
		return false
	}
	var api *APIError
	if errors.As(classified, &api) {
		return IsRetryable(api.HTTPStatus)
	}
	return true
}

// buildCallOptions translates a ChatRequest plus the Client's defaults
// into the eino per-call option slice. Per-call values win when set;
// otherwise the client config defaults apply. Mirrors
// langchain.Client.buildCallOptions.
func (c *Client) buildCallOptions(req ChatRequest) []einomodel.Option {
	opts := []einomodel.Option{}

	model := firstNonEmpty(req.Model, c.cfg.Model)
	if model != "" {
		opts = append(opts, einomodel.WithModel(model))
	}

	temp := c.cfg.Temperature
	if req.Temperature >= 0 {
		temp = req.Temperature
	}
	if temp >= 0 {
		opts = append(opts, einomodel.WithTemperature(float32(temp)))
	}

	topP := c.cfg.TopP
	if req.TopP >= 0 {
		topP = req.TopP
	}
	if topP >= 0 {
		opts = append(opts, einomodel.WithTopP(float32(topP)))
	}

	maxTokens := c.cfg.MaxTokens
	if req.MaxTokens > 0 {
		maxTokens = req.MaxTokens
	}
	if maxTokens > 0 {
		if c.cfg.Backend == BackendOpenAI {
			// Newer OpenAI models (o-series, gpt-4.1/4o latest, gpt-5, ...)
			// reject the deprecated `max_tokens` and require
			// `max_completion_tokens`. eino-ext's common WithMaxTokens maps
			// to `max_tokens`, so route through the openai-specific option
			// instead. Safe for official OpenAI/Azure, where
			// `max_completion_tokens` is supported by every chat model.
			opts = append(opts, einoopenai.WithMaxCompletionTokens(maxTokens))
		} else {
			opts = append(opts, einomodel.WithMaxTokens(maxTokens))
		}
	}

	if len(req.Stop) > 0 {
		opts = append(opts, einomodel.WithStop(req.Stop))
	}
	return opts
}

// toEinoMessages converts the package-internal Message slice into eino's
// schema.Message shape.
func toEinoMessages(msgs []Message) []*schema.Message {
	out := make([]*schema.Message, 0, len(msgs))
	for _, m := range msgs {
		out = append(out, &schema.Message{
			Role:    toEinoRole(m.Role),
			Content: m.Content,
			Name:    m.Name,
		})
	}
	return out
}

func toEinoRole(r Role) schema.RoleType {
	switch r {
	case RoleSystem:
		return schema.System
	case RoleAssistant:
		return schema.Assistant
	case RoleTool:
		return schema.Tool
	default:
		return schema.User
	}
}

// fromTokenUsage converts eino's token usage into the package shape,
// synthesising TotalTokens when the vendor omits it.
func fromTokenUsage(u *schema.TokenUsage) Usage {
	if u == nil {
		return Usage{}
	}
	out := Usage{
		PromptTokens:     u.PromptTokens,
		CompletionTokens: u.CompletionTokens,
		TotalTokens:      u.TotalTokens,
	}
	if out.TotalTokens == 0 {
		out.TotalTokens = out.PromptTokens + out.CompletionTokens
	}
	return out
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

// translateError lifts a raw eino / vendor error into one of the libs
// sentinels when the shape is recognisable. The adapter layer maps these
// to MathError(BOT_*) codes; see adapter/bot/errors.go.
//
// Typed vendor errors (google genai, go-openai) are lifted into *APIError
// first; everything else is classified by the same substrings as
// libs/langchain and passed through unchanged when unrecognised.
func (c *Client) translateError(err error) error {
	if err == nil {
		return nil
	}

	// Preserve typed sentinels.
	if errors.Is(err, ErrInvalidConfig) || errors.Is(err, ErrDecodeResponse) ||
		errors.Is(err, ErrContextTooLarge) || errors.Is(err, ErrUnsupportedOp) ||
		errors.Is(err, ErrRateLimited) {
		return err
	}

	if lifted := liftVendorError(c.cfg.Backend, err); lifted != nil {
		if IsRateLimited(lifted) {
			return fmt.Errorf("%w: %v", ErrRateLimited, lifted)
		}
		return lifted
	}

	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "context length") ||
		strings.Contains(msg, "context window") ||
		strings.Contains(msg, "maximum context"):
		return fmt.Errorf("%w: %v", ErrContextTooLarge, err)
	case strings.Contains(msg, "rate limit") ||
		strings.Contains(msg, "too many requests") ||
		strings.Contains(msg, "resource_exhausted") ||
		strings.Contains(msg, "resource exhausted") ||
		strings.Contains(msg, "quota"):
		return fmt.Errorf("%w: %v", ErrRateLimited, err)
	}

	return err
}
