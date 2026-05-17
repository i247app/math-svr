package langchain

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/googleai"

	"math-ai.com/math-ai/internal/infrastructure/logger"
)

// Client is the langchaingo-backed LLM client. It wraps a single concrete
// llms.Model (chosen by Config.Backend) and exposes a narrow chat / embed
// surface to the adapter layer.
//
// The client is safe for concurrent use; langchaingo's HTTP / gRPC clients
// pool connections internally. Lifecycle: nothing to Close — the inner
// transports manage their own pools. The struct is constructed once at
// boot and retained on the bot adapter for the process lifetime.
type Client struct {
	cfg Config
	llm llms.Model

	// embedder is the optional embeddings.EmbedderClient view of llm. It
	// is set when the backend's llms.Model also implements that interface.
	// Backends that need a separate embedding model are wired here too —
	// see newLLM.
	embedder embeddings.EmbedderClient
}

// Backend returns the backend the client was constructed with.
func (c *Client) Backend() Backend { return c.cfg.Backend }

// Model returns the configured default chat model id.
func (c *Client) Model() string { return c.cfg.Model }

// EmbedModel returns the configured default embedding model id. May be
// empty when the deploy does not need embeddings.
func (c *Client) EmbedModel() string { return c.cfg.EmbedModel }

// NewClient validates cfg, builds the configured langchaingo backend, and
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

	llm, embedder, err := newLLM(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("langchain: build backend %q: %w", cfg.Backend, err)
	}

	c := &Client{cfg: cfg, llm: llm, embedder: embedder}

	if cfg.RequireAtBoot {
		if err := c.probe(ctx); err != nil {
			return nil, fmt.Errorf("langchain: boot probe failed: %w", err)
		}
	} else {
		log := logger.From(ctx)
		log.Infof("langchain: client ready (backend=%s, model=%s, embed_model=%s, probe-skipped)",
			cfg.Backend, cfg.Model, cfg.EmbedModel)
	}

	return c, nil
}

// newLLM selects and constructs the concrete langchaingo backend. The
// returned embedder is non-nil when the same handle also implements
// embeddings.EmbedderClient.
//
// Adding a new backend = one new case here plus the matching import and
// any vendor-specific option mapping.
func newLLM(ctx context.Context, cfg Config) (llms.Model, embeddings.EmbedderClient, error) {
	switch cfg.Backend {
	case BackendGoogleAI:
		opts := []googleai.Option{
			googleai.WithAPIKey(cfg.APIKey),
			googleai.WithDefaultModel(cfg.Model),
		}
		if cfg.EmbedModel != "" {
			opts = append(opts, googleai.WithDefaultEmbeddingModel(cfg.EmbedModel))
		}
		if cfg.Temperature >= 0 {
			opts = append(opts, googleai.WithDefaultTemperature(cfg.Temperature))
		}
		if cfg.TopP >= 0 {
			opts = append(opts, googleai.WithDefaultTopP(cfg.TopP))
		}
		if cfg.MaxTokens > 0 {
			opts = append(opts, googleai.WithDefaultMaxTokens(cfg.MaxTokens))
		}
		g, err := googleai.New(ctx, opts...)
		if err != nil {
			return nil, nil, err
		}
		// *googleai.GoogleAI implements embeddings.EmbedderClient via
		// CreateEmbedding(ctx, texts). Cast is safe at this version.
		var embedder embeddings.EmbedderClient
		if ec, ok := any(g).(embeddings.EmbedderClient); ok {
			embedder = ec
		}
		return g, embedder, nil

	case BackendOpenAI, BackendAnthropic, BackendOllama:
		// Backend is recognised by Config.Validate so a deploy never sees
		// "unsupported backend" silently. Wiring the concrete sub-package
		// is intentionally left as a one-line follow-up so adding it does
		// not pull dead dependency weight into deploys that only use
		// Gemini today:
		//   - openai:    github.com/tmc/langchaingo/llms/openai
		//   - anthropic: github.com/tmc/langchaingo/llms/anthropic
		//   - ollama:    github.com/tmc/langchaingo/llms/ollama
		return nil, nil, fmt.Errorf(
			"%w: backend %q is registered but not yet wired in this build",
			ErrInvalidConfig, cfg.Backend)

	default:
		return nil, nil, fmt.Errorf("%w: unknown backend %q", ErrInvalidConfig, cfg.Backend)
	}
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

// Generate runs a non-streamed chat completion under the configured
// timeout and retry budget. Returns ErrInvalidConfig for an empty
// message list; for other failures see translateError.
func (c *Client) Generate(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	if len(req.Messages) == 0 {
		return nil, fmt.Errorf("%w: at least one message is required", ErrInvalidConfig)
	}

	callCtx, cancel := context.WithTimeout(ctx, c.cfg.Timeout)
	defer cancel()

	content := toLLMContent(req.Messages)
	opts := c.buildCallOptions(req)

	resp, err := c.withRetry(callCtx, func(retryCtx context.Context) (*llms.ContentResponse, error) {
		return c.llm.GenerateContent(retryCtx, content, opts...)
	})
	if err != nil {
		return nil, c.translateError(err)
	}
	if resp == nil || len(resp.Choices) == 0 {
		return nil, ErrDecodeResponse
	}

	choice := resp.Choices[0]
	out := &ChatResponse{
		Model:        firstNonEmpty(req.Model, c.cfg.Model),
		Content:      choice.Content,
		FinishReason: choice.StopReason,
	}
	out.Usage = readUsage(choice.GenerationInfo)
	return out, nil
}

// GenerateStream runs a streaming chat completion. Each token chunk is
// pushed to onChunk; the final invocation receives Done=true plus
// optional Usage info. onChunk MUST NOT block — the caller's pipeline
// owns downstream buffering.
//
// Streaming is not retried automatically: a mid-stream failure would
// require replaying delivered tokens, which is the caller's policy
// decision (e.g. restart at the application layer with a new SSE
// connection).
func (c *Client) GenerateStream(ctx context.Context, req ChatRequest, onChunk func(StreamChunk) error) (*ChatResponse, error) {
	if len(req.Messages) == 0 {
		return nil, fmt.Errorf("%w: at least one message is required", ErrInvalidConfig)
	}
	if onChunk == nil {
		return nil, fmt.Errorf("%w: onChunk is required for streaming", ErrInvalidConfig)
	}

	callCtx, cancel := context.WithTimeout(ctx, c.cfg.Timeout)
	defer cancel()

	content := toLLMContent(req.Messages)
	opts := append(c.buildCallOptions(req),
		llms.WithStreamingFunc(func(_ context.Context, chunk []byte) error {
			return onChunk(StreamChunk{Delta: string(chunk)})
		}),
	)

	resp, err := c.llm.GenerateContent(callCtx, content, opts...)
	if err != nil {
		return nil, c.translateError(err)
	}

	final := StreamChunk{Done: true}
	out := &ChatResponse{Model: firstNonEmpty(req.Model, c.cfg.Model)}
	if resp != nil && len(resp.Choices) > 0 {
		choice := resp.Choices[0]
		out.Content = choice.Content
		out.FinishReason = choice.StopReason
		out.Usage = readUsage(choice.GenerationInfo)
		if out.Usage != (Usage{}) {
			usage := out.Usage
			final.Usage = &usage
		}
	}
	if err := onChunk(final); err != nil {
		return out, err
	}
	return out, nil
}

// Embed produces embeddings for each input string, preserving order.
// Returns ErrUnsupportedOp when the configured backend does not expose
// an embeddings client.
func (c *Client) Embed(ctx context.Context, req EmbedRequest) (*EmbedResponse, error) {
	if len(req.Inputs) == 0 {
		return &EmbedResponse{Model: firstNonEmpty(req.Model, c.cfg.EmbedModel, c.cfg.Model)}, nil
	}
	if c.embedder == nil {
		return nil, ErrUnsupportedOp
	}

	callCtx, cancel := context.WithTimeout(ctx, c.cfg.Timeout)
	defer cancel()

	var vectors [][]float32
	_, err := c.withRetry(callCtx, func(retryCtx context.Context) (*llms.ContentResponse, error) {
		v, err := c.embedder.CreateEmbedding(retryCtx, req.Inputs)
		if err != nil {
			return nil, err
		}
		vectors = v
		return nil, nil
	})
	if err != nil {
		return nil, c.translateError(err)
	}

	return &EmbedResponse{
		Model:   firstNonEmpty(req.Model, c.cfg.EmbedModel, c.cfg.Model),
		Vectors: vectors,
	}, nil
}

// withRetry executes op under the client's MaxRetries / RetryDelay
// budget. Errors that classify as auth / config / rate-limited / context
// length are NOT retried — they will not recover within the same call.
// Transport-shaped errors (i.e. anything that does not look like a
// vendor 4xx) and HTTP 5xx classified by IsRetryable are retried with a
// linear sleep.
func (c *Client) withRetry(ctx context.Context, op func(ctx context.Context) (*llms.ContentResponse, error)) (*llms.ContentResponse, error) {
	var lastErr error
	for attempt := 0; attempt <= c.cfg.MaxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(c.cfg.RetryDelay):
			}
		}

		resp, err := op(ctx)
		if err == nil {
			return resp, nil
		}

		lastErr = err
		if !shouldRetry(err) {
			return nil, err
		}
	}
	return nil, lastErr
}

// shouldRetry classifies an error from the underlying llms.Model call.
// Non-recoverable: auth, config, context-too-large, rate-limit, ctx
// cancellation. Everything else (transport, 5xx, unclassified) is
// considered transient.
func shouldRetry(err error) bool {
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
	if IsAuthError(err) || IsConfigError(err) || IsRateLimited(err) {
		return false
	}
	var api *APIError
	if errors.As(err, &api) {
		return IsRetryable(api.HTTPStatus)
	}
	return true
}

// buildCallOptions translates a ChatRequest plus the Client's defaults
// into the langchaingo per-call option slice. Per-call values win when
// set; otherwise the client config defaults apply.
func (c *Client) buildCallOptions(req ChatRequest) []llms.CallOption {
	opts := []llms.CallOption{}

	model := firstNonEmpty(req.Model, c.cfg.Model)
	if model != "" {
		opts = append(opts, llms.WithModel(model))
	}

	temp := c.cfg.Temperature
	if req.Temperature >= 0 {
		temp = req.Temperature
	}
	if temp >= 0 {
		opts = append(opts, llms.WithTemperature(temp))
	}

	topP := c.cfg.TopP
	if req.TopP >= 0 {
		topP = req.TopP
	}
	if topP >= 0 {
		opts = append(opts, llms.WithTopP(topP))
	}

	maxTokens := c.cfg.MaxTokens
	if req.MaxTokens > 0 {
		maxTokens = req.MaxTokens
	}
	if maxTokens > 0 {
		opts = append(opts, llms.WithMaxTokens(maxTokens))
	}

	if len(req.Stop) > 0 {
		opts = append(opts, llms.WithStopWords(req.Stop))
	}
	if req.JSONMode {
		opts = append(opts, llms.WithJSONMode())
	}
	return opts
}

// toLLMContent converts the package-internal Message slice into the shape
// langchaingo's Model interface expects.
func toLLMContent(msgs []Message) []llms.MessageContent {
	out := make([]llms.MessageContent, 0, len(msgs))
	for _, m := range msgs {
		out = append(out, llms.MessageContent{
			Role:  toLLMRole(m.Role),
			Parts: []llms.ContentPart{llms.TextPart(m.Content)},
		})
	}
	return out
}

func toLLMRole(r Role) llms.ChatMessageType {
	switch r {
	case RoleSystem:
		return llms.ChatMessageTypeSystem
	case RoleAssistant:
		return llms.ChatMessageTypeAI
	case RoleTool:
		return llms.ChatMessageTypeTool
	default:
		return llms.ChatMessageTypeHuman
	}
}

// readUsage attempts to extract token counts from langchaingo's
// GenerationInfo map. Different vendors populate different keys; we
// inspect the well-known ones and return a zero-value Usage when nothing
// matches.
func readUsage(info map[string]any) Usage {
	if info == nil {
		return Usage{}
	}
	var u Usage
	u.PromptTokens = readUsageField(info, "PromptTokens", "input_tokens", "prompt_tokens")
	u.CompletionTokens = readUsageField(info, "CompletionTokens", "output_tokens", "completion_tokens")
	u.TotalTokens = readUsageField(info, "TotalTokens", "total_tokens")
	if u.TotalTokens == 0 {
		u.TotalTokens = u.PromptTokens + u.CompletionTokens
	}
	return u
}

func readUsageField(info map[string]any, keys ...string) int {
	for _, k := range keys {
		v, ok := info[k]
		if !ok {
			continue
		}
		switch x := v.(type) {
		case int:
			return x
		case int32:
			return int(x)
		case int64:
			return int(x)
		case float64:
			return int(x)
		case float32:
			return int(x)
		}
	}
	return 0
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

// translateError lifts a raw langchaingo / vendor error into one of the
// libs sentinels when the shape is recognisable. The adapter layer maps
// these to MathError(BOT_*) codes; see adapter/bot/errors.go.
//
// langchaingo does not expose a uniform error type today; vendor SDKs
// surface plain errors with vendor-specific message strings. We detect
// the well-known shapes (context-window, rate-limit) by substring and
// pass everything else through unchanged.
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

	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "context length") ||
		strings.Contains(msg, "context window") ||
		strings.Contains(msg, "maximum context"):
		return fmt.Errorf("%w: %v", ErrContextTooLarge, err)
	case strings.Contains(msg, "rate limit") ||
		strings.Contains(msg, "too many requests") ||
		strings.Contains(msg, "quota"):
		return fmt.Errorf("%w: %v", ErrRateLimited, err)
	}

	return err
}
