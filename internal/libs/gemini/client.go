package gemini

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"math-ai.com/math-ai/internal/infrastructure/logger"
	"math-ai.com/math-ai/internal/shared/http_client"
)

// Client is the direct Gemini LLM client. It speaks the v1beta REST API
// over the project's shared http_client — no vendor SDK, no private
// http.Client — and exposes the same narrow surface as the sibling libs
// packages.
//
// Capability matrix: Chat, Stream and Embed are all supported.
//
// The client is safe for concurrent use; http_client wraps a stdlib
// http.Client which pools connections internally. Nothing to Close —
// constructed once at boot and retained on the bot adapter for the
// process lifetime.
type Client struct {
	cfg  Config
	http *http_client.Client
}

// Model returns the configured default model id, as supplied in config
// (without the "models/" URL prefix normalisation).
func (c *Client) Model() string { return c.cfg.Model }

// EmbedModel returns the configured default embedding model id. May be
// empty, in which case Embed falls back to the chat model.
func (c *Client) EmbedModel() string { return c.cfg.EmbedModel }

// BaseURL returns the resolved REST root. Exposed for log/diagnostic
// surfaces; never includes the credential.
func (c *Client) BaseURL() string { return c.cfg.BaseURL }

// NewClient validates cfg, builds the shared HTTP client with Gemini's
// auth header, and (when cfg.RequireAtBoot is true) runs a tiny probe call
// to fail fast on a bad credential.
//
// Errors:
//   - cfg.Validate failure → wraps ErrInvalidConfig (adapter factory maps
//     to MathError(BOT_CONFIG_INVALID)).
//   - probe failure under RequireAtBoot=true → wraps the underlying error
//     (adapter factory maps to MathError(BOT_CONNECT_FAILED)).
func NewClient(ctx context.Context, cfg Config) (*Client, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	cfg.applyDefaults()

	c := &Client{
		cfg: cfg,
		http: http_client.NewClient(
			http_client.WithBaseURL(cfg.BaseURL),
			http_client.WithTimeout(cfg.Timeout),
			// Header, never ?key=<secret>: the query-parameter form would
			// put the credential into every access log and proxy trace.
			http_client.WithHeader(apiKeyHeader, cfg.APIKey),
			http_client.WithContentType("application/json"),
			// Drop http_client's default [Logging, Timing] chain. The
			// logging interceptor dumps every request header — including
			// the api key above — and the full response body. Neither may
			// reach the logs. This package emits its own operator-metadata
			// log lines instead. Same posture as libs/openai,
			// libs/openrouter and the Twilio client.
			http_client.WithInterceptors(),
		),
	}

	if cfg.RequireAtBoot {
		if err := c.probe(ctx); err != nil {
			return nil, fmt.Errorf("gemini: boot probe failed: %w", err)
		}
	} else {
		logger.From(ctx).Infof("gemini: client ready (model=%s, embed_model=%s, base_url=%s, probe-skipped)",
			cfg.Model, cfg.EmbedModel, cfg.BaseURL)
	}

	return c, nil
}

// probe makes a one-token generation to verify credentials and
// connectivity. Used only when RequireAtBoot=true.
func (c *Client) probe(ctx context.Context) error {
	_, err := c.Generate(ctx, ChatRequest{
		Model:     c.cfg.Model,
		Messages:  []Message{{Role: RoleUser, Content: "ping"}},
		MaxTokens: 1,
	})
	return err
}

// Generate runs a non-streamed generation under the configured timeout and
// retry budget.
func (c *Client) Generate(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	if len(req.Messages) == 0 {
		return nil, fmt.Errorf("%w: at least one message is required", ErrInvalidConfig)
	}

	callCtx, cancel := context.WithTimeout(ctx, c.cfg.Timeout)
	defer cancel()

	model := normalizeModel(firstNonEmpty(req.Model, c.cfg.Model))
	body, err := c.post(callCtx, "/"+model+":generateContent", c.buildRequest(req))
	if err != nil {
		return nil, err
	}

	var wire wireGenerateResponse
	if err := json.Unmarshal(body, &wire); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDecodeResponse, err)
	}
	if wire.Error != nil {
		return nil, classifyAPIError(wire.Error.toAPIError(200, 0))
	}

	// A blocked PROMPT comes back HTTP 200 with zero candidates and the
	// reason only in promptFeedback. Checking this before the empty-
	// candidates branch is what turns a baffling "no candidates" into an
	// actionable "the filter refused this prompt".
	if wire.PromptFeedback != nil && wire.PromptFeedback.BlockReason != "" {
		return nil, fmt.Errorf("%w: prompt blocked (%s)",
			ErrContentBlocked, wire.PromptFeedback.BlockReason)
	}
	if len(wire.Candidates) == 0 {
		return nil, fmt.Errorf("%w: response carried no candidates", ErrDecodeResponse)
	}

	cand := wire.Candidates[0]
	// A blocked ANSWER also comes back 200, with the reason on the
	// candidate instead.
	if isBlockedReason(cand.FinishReason) {
		return nil, fmt.Errorf("%w: response blocked (%s)", ErrContentBlocked, cand.FinishReason)
	}

	resp := &ChatResponse{
		Model:        firstNonEmpty(wire.ModelVersion, req.Model, c.cfg.Model),
		Content:      cand.text(),
		FinishReason: cand.FinishReason,
		Usage:        wire.UsageMetadata.toUsage(),
	}

	// A generation cut off by the token cap is still HTTP 200 but the JSON
	// the caller is about to parse is truncated. Surface it as a decode
	// failure rather than letting quiz generation fail on malformed JSON
	// with no explanation.
	if cand.FinishReason == finishReasonMaxTokens {
		return resp, fmt.Errorf("%w: generation truncated (finishReason=MAX_TOKENS, completion_tokens=%d) — raise MaxTokens",
			ErrDecodeResponse, resp.Usage.CompletionTokens)
	}

	// Operator metadata only — never the prompt or the response body.
	logger.From(ctx).Infof("gemini.generate model=%s finish=%s prompt_tokens=%d completion_tokens=%d total_tokens=%d",
		resp.Model, resp.FinishReason,
		resp.Usage.PromptTokens, resp.Usage.CompletionTokens, resp.Usage.TotalTokens)
	return resp, nil
}

// GenerateStream runs :streamGenerateContent?alt=sse and replays the
// resulting SSE events through onChunk: one call per non-empty text delta,
// then exactly one final call with Done=true carrying Usage.
//
// LIMITATION — the deltas are parsed from a fully-buffered response, not
// delivered as they arrive on the wire. http_client.Do reads the whole
// body before returning, and this package is deliberately built on that
// shared client rather than its own http.Client. Token sequence and the
// assembled result are correct, but the first chunk is not handed over any
// sooner than the last. Identical constraint to libs/openai and
// libs/openrouter.
//
// Streaming is not retried: a mid-stream failure would require replaying
// delivered tokens, which is the caller's policy decision.
func (c *Client) GenerateStream(ctx context.Context, req ChatRequest, onChunk func(StreamChunk) error) (*ChatResponse, error) {
	if len(req.Messages) == 0 {
		return nil, fmt.Errorf("%w: at least one message is required", ErrInvalidConfig)
	}
	if onChunk == nil {
		return nil, fmt.Errorf("%w: onChunk is required for streaming", ErrInvalidConfig)
	}

	callCtx, cancel := context.WithTimeout(ctx, c.cfg.Timeout)
	defer cancel()

	model := normalizeModel(firstNonEmpty(req.Model, c.cfg.Model))
	body, err := c.post(callCtx, "/"+model+":streamGenerateContent", c.buildRequest(req),
		// Without alt=sse the endpoint streams a JSON ARRAY instead of SSE
		// frames, which parseSSE cannot read.
		http_client.WithRequestQueryParam("alt", "sse"))
	if err != nil {
		return nil, err
	}

	events, err := parseSSE(body)
	if err != nil {
		return nil, err
	}

	out := &ChatResponse{Model: firstNonEmpty(req.Model, c.cfg.Model)}
	var content strings.Builder

	for _, ev := range events {
		if ev.Error != nil {
			return nil, classifyAPIError(ev.Error.toAPIError(200, 0))
		}
		if ev.ModelVersion != "" {
			out.Model = ev.ModelVersion
		}
		if ev.UsageMetadata != nil {
			out.Usage = ev.UsageMetadata.toUsage()
		}
		if ev.PromptFeedback != nil && ev.PromptFeedback.BlockReason != "" {
			return nil, fmt.Errorf("%w: prompt blocked (%s)",
				ErrContentBlocked, ev.PromptFeedback.BlockReason)
		}
		for _, cand := range ev.Candidates {
			if isBlockedReason(cand.FinishReason) {
				return nil, fmt.Errorf("%w: response blocked mid-stream (%s)",
					ErrContentBlocked, cand.FinishReason)
			}
			if cand.FinishReason != "" {
				out.FinishReason = cand.FinishReason
			}
			delta := cand.text()
			if delta == "" {
				continue
			}
			content.WriteString(delta)
			if cbErr := onChunk(StreamChunk{Delta: delta}); cbErr != nil {
				return nil, cbErr
			}
		}
	}
	out.Content = content.String()

	final := StreamChunk{Done: true}
	if out.Usage != (Usage{}) {
		usage := out.Usage
		final.Usage = &usage
	}
	if cbErr := onChunk(final); cbErr != nil {
		return out, cbErr
	}

	logger.From(ctx).Infof("gemini.stream model=%s finish=%s prompt_tokens=%d completion_tokens=%d",
		out.Model, out.FinishReason, out.Usage.PromptTokens, out.Usage.CompletionTokens)
	return out, nil
}

// Embed produces one vector per input via :batchEmbedContents.
//
// Model precedence is EmbedRequest.Model → Config.EmbedModel →
// Config.Model, matching libs/langchain and libs/openai.
//
// Order: unlike OpenAI's embeddings endpoint, the Gemini reference states
// the embeddings come back "in the same order as provided in the batch
// request", and the elements carry no index to re-sort by. The count check
// below is therefore the only guard available.
func (c *Client) Embed(ctx context.Context, req EmbedRequest) (*EmbedResponse, error) {
	if len(req.Inputs) == 0 {
		return nil, fmt.Errorf("%w: at least one input is required", ErrInvalidConfig)
	}

	name := firstNonEmpty(req.Model, c.cfg.EmbedModel, c.cfg.Model)
	if name == "" {
		return nil, fmt.Errorf("%w: no embedding model configured", ErrInvalidConfig)
	}
	model := normalizeModel(name)

	callCtx, cancel := context.WithTimeout(ctx, c.cfg.Timeout)
	defer cancel()

	reqs := make([]wireEmbedContentRequest, 0, len(req.Inputs))
	for _, in := range req.Inputs {
		reqs = append(reqs, wireEmbedContentRequest{
			Model:   model,
			Content: wireContent{Parts: []wirePart{{Text: in}}},
		})
	}

	body, err := c.post(callCtx, "/"+model+":batchEmbedContents",
		wireBatchEmbedRequest{Requests: reqs})
	if err != nil {
		return nil, err
	}

	var wire wireBatchEmbedResponse
	if err := json.Unmarshal(body, &wire); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDecodeResponse, err)
	}
	if wire.Error != nil {
		return nil, classifyAPIError(wire.Error.toAPIError(200, 0))
	}
	if len(wire.Embeddings) != len(req.Inputs) {
		return nil, fmt.Errorf("%w: got %d embeddings for %d inputs",
			ErrDecodeResponse, len(wire.Embeddings), len(req.Inputs))
	}

	vectors := make([][]float32, 0, len(wire.Embeddings))
	for _, e := range wire.Embeddings {
		vectors = append(vectors, e.Values)
	}

	resp := &EmbedResponse{
		Model:   name,
		Vectors: vectors,
		Usage:   wire.UsageMetadata.toUsage(),
	}
	logger.From(ctx).Infof("gemini.embed model=%s inputs=%d prompt_tokens=%d",
		resp.Model, len(req.Inputs), resp.Usage.PromptTokens)
	return resp, nil
}

// buildRequest translates a ChatRequest plus the Client's defaults into
// the wire body.
//
// This is the only genuinely lossy step in the package, because Gemini's
// conversation model is narrower than the adapter's:
//
//   - system messages have no role here; they are lifted out and joined
//     into the top-level systemInstruction.
//   - assistant becomes "model"; user stays "user".
//   - tool becomes "user". Gemini's real equivalent is a functionResponse
//     part, but the adapter's tool messages are plain text with no call to
//     correlate them to, so "user" is the honest mapping.
//   - consecutive messages of the same role are merged into one content
//     with several parts, because the API expects the turns to alternate.
func (c *Client) buildRequest(req ChatRequest) wireGenerateRequest {
	var systemParts []string
	var contents []wireContent

	for _, m := range req.Messages {
		if m.Role == RoleSystem {
			systemParts = append(systemParts, m.Content)
			continue
		}

		role := "user"
		if m.Role == RoleAssistant {
			role = "model"
		}

		if n := len(contents); n > 0 && contents[n-1].Role == role {
			contents[n-1].Parts = append(contents[n-1].Parts, wirePart{Text: m.Content})
			continue
		}
		contents = append(contents, wireContent{
			Role:  role,
			Parts: []wirePart{{Text: m.Content}},
		})
	}

	out := wireGenerateRequest{Contents: contents}
	if len(systemParts) > 0 {
		out.SystemInstruction = &wireContent{
			Parts: []wirePart{{Text: strings.Join(systemParts, "\n\n")}},
		}
	}

	cfg := wireGenerationConfig{StopSequences: req.Stop}
	empty := true

	temp := c.cfg.Temperature
	if req.Temperature >= 0 {
		temp = req.Temperature
	}
	if temp >= 0 {
		cfg.Temperature = &temp
		empty = false
	}

	topP := c.cfg.TopP
	if req.TopP >= 0 {
		topP = req.TopP
	}
	if topP >= 0 {
		cfg.TopP = &topP
		empty = false
	}

	maxTokens := c.cfg.MaxTokens
	if req.MaxTokens > 0 {
		maxTokens = req.MaxTokens
	}
	if maxTokens > 0 {
		cfg.MaxOutputTokens = &maxTokens
		empty = false
	}

	if req.JSONMode {
		cfg.ResponseMimeType = "application/json"
		empty = false
	}

	if !empty || len(req.Stop) > 0 {
		out.GenerationConfig = &cfg
	}
	return out
}

// post sends body to path under the retry budget and returns the raw 2xx
// response bytes.
//
// Retry policy: transport failures and the statuses IsRetryable accepts
// are re-issued. Auth, permission, unknown-model, context-length, blocked
// content and daily-quota failures are not, because they cannot recover
// within the call. A throttling 429 is retried only when it carries a
// RetryInfo hint the client is willing to wait out (see maxRetryAfter).
func (c *Client) post(ctx context.Context, path string, body any, opts ...http_client.RequestOption) ([]byte, error) {
	var lastErr error
	delay := c.cfg.RetryDelay

	for attempt := 0; attempt <= c.cfg.MaxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
			}
		}

		// Post rebuilds the *http.Request on every attempt, so the JSON
		// body is re-encoded rather than replayed from a consumed reader.
		resp, err := c.http.Post(ctx, path, body, opts...)
		if err != nil {
			if ctxErr := ctx.Err(); ctxErr != nil {
				return nil, ctxErr
			}
			// Transport-shaped failure (DNS, dial, TLS, read): transient.
			lastErr = fmt.Errorf("gemini: request failed: %w", err)
			delay = c.cfg.RetryDelay
			continue
		}

		// Stands in for http_client's LoggingInterceptor, which this client
		// switches off in NewClient because it would dump the api key
		// header and the full body. Status + size + attempt is the operator
		// signal; the payload stays out of the logs.
		logger.From(ctx).Infof("gemini.http path=%s status=%d bytes=%d attempt=%d/%d",
			path, resp.StatusCode, len(resp.Body), attempt+1, c.cfg.MaxRetries+1)

		if resp.IsSuccess() {
			return resp.Body, nil
		}

		apiErr := errorFromResponse(resp)
		lastErr = classifyAPIError(apiErr)

		retryIn, ok := retryDelayFor(apiErr, c.cfg.RetryDelay)
		if !ok {
			return nil, lastErr
		}
		delay = retryIn
	}

	return nil, lastErr
}

// retryDelayFor decides whether apiErr is worth another attempt and how
// long to wait first. Returns ok=false to surface the error immediately.
func retryDelayFor(apiErr *APIError, base time.Duration) (time.Duration, bool) {
	if apiErr == nil {
		return base, false
	}
	if apiErr.HTTPStatus == 429 || apiErr.Status == "RESOURCE_EXHAUSTED" {
		// A daily quota will still be exhausted a second from now.
		if isQuotaExhaustion(apiErr) {
			return 0, false
		}
		// Without a RetryInfo hint there is no safe interval to guess at,
		// and with an excessive one waiting would hold the caller's HTTP
		// handler open — surface both cases instead.
		if apiErr.RetryAfter <= 0 || apiErr.RetryAfter > maxRetryAfter {
			return 0, false
		}
		return apiErr.RetryAfter, true
	}
	if IsRetryable(apiErr.HTTPStatus) {
		return base, true
	}
	return 0, false
}

// errorFromResponse lifts a non-2xx response into a typed *APIError.
//
// Log/leak discipline: an undecodable body is reported by size only. The
// raw bytes never enter the error, because an upstream proxy or WAF can
// put arbitrary content there and this message travels into the MathError
// debug field.
func errorFromResponse(resp *http_client.Response) *APIError {
	retryAfter := parseRetryAfter(resp.Header("Retry-After"))

	var envelope struct {
		Error *wireError `json:"error"`
	}
	if err := json.Unmarshal(resp.Body, &envelope); err != nil || envelope.Error == nil {
		return &APIError{
			HTTPStatus: resp.StatusCode,
			RetryAfter: retryAfter,
			Message: fmt.Sprintf("unexpected response (%d bytes, not a Gemini error envelope)",
				len(resp.Body)),
		}
	}
	return envelope.Error.toAPIError(resp.StatusCode, retryAfter)
}

// parseSSE decodes a Gemini SSE body into its JSON events, in order.
//
// Gemini's stream carries no [DONE] terminator — it simply ends — so the
// loop runs to EOF. Comment lines (":") are skipped before any JSON
// parsing, and a [DONE] is tolerated in case a proxy adds one.
func parseSSE(body []byte) ([]wireGenerateResponse, error) {
	var events []wireGenerateResponse

	scanner := bufio.NewScanner(bytes.NewReader(body))
	// SSE payloads carrying a whole JSON chunk can exceed bufio's 64 KB
	// default line cap.
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, ":") {
			continue
		}
		payload, ok := strings.CutPrefix(line, "data:")
		if !ok {
			// Any other SSE field (event:, id:, retry:) is not part of this
			// API's contract; ignore rather than fail.
			continue
		}
		payload = strings.TrimSpace(payload)
		if payload == "[DONE]" {
			break
		}

		var ev wireGenerateResponse
		if err := json.Unmarshal([]byte(payload), &ev); err != nil {
			return nil, fmt.Errorf("%w: stream chunk: %v", ErrDecodeResponse, err)
		}
		events = append(events, ev)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("%w: reading stream: %v", ErrDecodeResponse, err)
	}
	if len(events) == 0 {
		return nil, fmt.Errorf("%w: stream carried no events", ErrDecodeResponse)
	}
	return events, nil
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
