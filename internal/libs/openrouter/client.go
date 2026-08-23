package openrouter

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

// Client is the OpenRouter-backed LLM client. It speaks the REST API
// directly over the project's shared http_client — no vendor SDK, no
// private http.Client — and exposes the same narrow chat surface as
// libs/eino.Client, minus embeddings (OpenRouter's chat-completions API
// has no embedding endpoint; the adapter reports Embed as unsupported).
//
// The client is safe for concurrent use; http_client wraps a stdlib
// http.Client which pools connections internally. Nothing to Close —
// constructed once at boot and retained on the bot adapter for the
// process lifetime.
type Client struct {
	cfg  Config
	http *http_client.Client
}

// Model returns the configured default chat model id.
func (c *Client) Model() string { return c.cfg.Model }

// BaseURL returns the resolved REST root. Exposed for log/diagnostic
// surfaces; never includes the credential.
func (c *Client) BaseURL() string { return c.cfg.BaseURL }

// NewClient validates cfg, builds the shared HTTP client with OpenRouter's
// auth + attribution headers, and (when cfg.RequireAtBoot is true) runs a
// tiny probe call to fail fast on a bad credential.
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

	opts := []http_client.Option{
		http_client.WithBaseURL(cfg.BaseURL),
		http_client.WithTimeout(cfg.Timeout),
		http_client.WithBearerToken(cfg.APIKey),
		http_client.WithContentType("application/json"),
		// Drop http_client's default [Logging, Timing] chain. The logging
		// interceptor dumps every request header — including the
		// Authorization bearer above — and the full response body. Neither
		// may reach the logs. This package emits its own operator-metadata
		// log line at the end of Generate instead. Same posture as the
		// Twilio client; see http_client.WithInterceptors.
		http_client.WithInterceptors(http_client.NewTimingInterceptor()),
	}
	// Attribution headers are optional and affect only OpenRouter's public
	// rankings, never whether a call succeeds.
	if s := strings.TrimSpace(cfg.SiteURL); s != "" {
		opts = append(opts, http_client.WithHeader("HTTP-Referer", s))
	}
	if s := strings.TrimSpace(cfg.AppTitle); s != "" {
		opts = append(opts, http_client.WithHeader("X-OpenRouter-Title", s))
	}

	c := &Client{cfg: cfg, http: http_client.NewClient(opts...)}

	if cfg.RequireAtBoot {
		if err := c.probe(ctx); err != nil {
			return nil, fmt.Errorf("openrouter: boot probe failed: %w", err)
		}
	} else {
		logger.From(ctx).Infof("openrouter: client ready (model=%s, base_url=%s, probe-skipped)",
			cfg.Model, cfg.BaseURL)
	}

	return c, nil
}

// probe makes a one-token completion to verify credentials and
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
// timeout and retry budget. Returns ErrInvalidConfig for an empty message
// list; for other failures see translateError.
func (c *Client) Generate(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	if len(req.Messages) == 0 {
		return nil, fmt.Errorf("%w: at least one message is required", ErrInvalidConfig)
	}

	callCtx, cancel := context.WithTimeout(ctx, c.cfg.Timeout)
	defer cancel()

	body, err := c.doChat(callCtx, c.buildRequest(req, false))
	if err != nil {
		return nil, err
	}

	var wire wireChatResponse
	if err := json.Unmarshal(body, &wire); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDecodeResponse, err)
	}
	// A 200 can still carry an error envelope — OpenRouter uses that shape
	// when the routed upstream failed after the response headers went out.
	if wire.Error != nil {
		return nil, classifyAPIError(wire.Error.toAPIError(200))
	}
	if len(wire.Choices) == 0 {
		return nil, fmt.Errorf("%w: response carried no choices", ErrDecodeResponse)
	}
	choice := wire.Choices[0]
	if choice.Error != nil {
		return nil, classifyAPIError(choice.Error.toAPIError(200))
	}
	if choice.FinishReason == finishReasonError {
		return nil, fmt.Errorf("openrouter: upstream reported finish_reason=error")
	}

	resp := &ChatResponse{
		Model:        firstNonEmpty(wire.Model, req.Model, c.cfg.Model),
		Content:      choice.Message.Content,
		FinishReason: choice.FinishReason,
		Usage:        wire.Usage.toUsage(),
	}

	// Operator metadata only — never the prompt or the response body.
	logger.From(ctx).Infof("openrouter.generate model=%s finish=%s prompt_tokens=%d completion_tokens=%d total_tokens=%d",
		resp.Model, resp.FinishReason,
		resp.Usage.PromptTokens, resp.Usage.CompletionTokens, resp.Usage.TotalTokens)
	return resp, nil
}

// GenerateStream runs a chat completion with `stream: true` and replays
// the resulting SSE events through onChunk: one call per non-empty content
// delta, then exactly one final call with Done=true carrying Usage when the
// upstream reported it.
//
// LIMITATION — the deltas are parsed from a fully-buffered response, not
// delivered as they arrive on the wire. http_client.Do reads the whole
// body before returning, and this package is deliberately built on that
// shared client rather than on its own http.Client. So the token sequence
// and the assembled result are correct, but the first chunk is not handed
// over any sooner than the last. Callers that need true incremental
// delivery (e.g. piping tokens to a WebSocket) need a streaming-capable
// transport first.
//
// Streaming is not retried: a mid-stream failure would require replaying
// delivered tokens, which is the caller's policy decision — same contract
// as libs/eino.
func (c *Client) GenerateStream(ctx context.Context, req ChatRequest, onChunk func(StreamChunk) error) (*ChatResponse, error) {
	if len(req.Messages) == 0 {
		return nil, fmt.Errorf("%w: at least one message is required", ErrInvalidConfig)
	}
	if onChunk == nil {
		return nil, fmt.Errorf("%w: onChunk is required for streaming", ErrInvalidConfig)
	}

	callCtx, cancel := context.WithTimeout(ctx, c.cfg.Timeout)
	defer cancel()

	body, err := c.doChat(callCtx, c.buildRequest(req, true))
	if err != nil {
		return nil, err
	}

	out := &ChatResponse{Model: firstNonEmpty(req.Model, c.cfg.Model)}
	var content strings.Builder

	events, err := parseSSE(body)
	if err != nil {
		return nil, err
	}
	for _, ev := range events {
		if ev.Model != "" {
			out.Model = ev.Model
		}
		if ev.Error != nil {
			return nil, classifyAPIError(ev.Error.toAPIError(200))
		}
		if ev.Usage != nil {
			out.Usage = ev.Usage.toUsage()
		}
		for _, choice := range ev.Choices {
			if choice.Error != nil {
				return nil, classifyAPIError(choice.Error.toAPIError(200))
			}
			// A stream that fails midway keeps HTTP 200 and signals the
			// failure here — surface it instead of returning a truncated
			// answer as if it were complete.
			if choice.FinishReason == finishReasonError {
				return nil, fmt.Errorf("openrouter: upstream reported finish_reason=error mid-stream")
			}
			if choice.FinishReason != "" {
				out.FinishReason = choice.FinishReason
			}
			if choice.Delta.Content == "" {
				continue
			}
			content.WriteString(choice.Delta.Content)
			if cbErr := onChunk(StreamChunk{Delta: choice.Delta.Content}); cbErr != nil {
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

	logger.From(ctx).Infof("openrouter.stream model=%s finish=%s prompt_tokens=%d completion_tokens=%d",
		out.Model, out.FinishReason, out.Usage.PromptTokens, out.Usage.CompletionTokens)
	return out, nil
}

// finishReasonError is OpenRouter's marker for "the routed upstream failed
// after we already answered 200".
const finishReasonError = "error"

// buildRequest translates a ChatRequest plus the Client's defaults into
// the wire body. Per-call values win when set; otherwise the client config
// defaults apply. Unset sampling fields are omitted from the JSON entirely
// so the model's own defaults stand.
func (c *Client) buildRequest(req ChatRequest, stream bool) wireChatRequest {
	msgs := make([]wireMessage, 0, len(req.Messages))
	for _, m := range req.Messages {
		msgs = append(msgs, wireMessage{
			Role:    string(m.Role),
			Content: m.Content,
			Name:    m.Name,
		})
	}

	out := wireChatRequest{
		Model:    firstNonEmpty(req.Model, c.cfg.Model),
		Messages: msgs,
		Stream:   stream,
		Stop:     req.Stop,
	}
	if stream {
		out.StreamOptions = &wireStreamOptions{IncludeUsage: true}
	}

	temp := c.cfg.Temperature
	if req.Temperature >= 0 {
		temp = req.Temperature
	}
	if temp >= 0 {
		out.Temperature = &temp
	}

	topP := c.cfg.TopP
	if req.TopP >= 0 {
		topP = req.TopP
	}
	if topP >= 0 {
		out.TopP = &topP
	}

	maxTokens := c.cfg.MaxTokens
	if req.MaxTokens > 0 {
		maxTokens = req.MaxTokens
	}
	if maxTokens > 0 {
		out.MaxTokens = &maxTokens
	}

	if req.JSONMode {
		out.ResponseFormat = &wireResponseFormat{Type: "json_object"}
	}
	return out
}

// doChat POSTs body to /chat/completions under the retry budget and
// returns the raw 2xx response bytes.
//
// Retry policy mirrors libs/eino: transport failures and the statuses
// IsRetryable accepts are re-issued; auth / config / credit / context
// failures are not, because they cannot recover within the same call. A
// 429 is retried only when it carries a Retry-After the client is willing
// to wait out (see maxRetryAfter) — otherwise it is surfaced so business
// -level backoff can apply.
func (c *Client) doChat(ctx context.Context, body wireChatRequest) ([]byte, error) {
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
		resp, err := c.http.Post(ctx, chatCompletionsPath, body)
		if err != nil {
			if ctxErr := ctx.Err(); ctxErr != nil {
				return nil, ctxErr
			}
			// Transport-shaped failure (DNS, dial, TLS, read): transient.
			lastErr = fmt.Errorf("openrouter: request failed: %w", err)
			delay = c.cfg.RetryDelay
			continue
		}

		// Stands in for http_client's LoggingInterceptor, which this client
		// switches off in NewClient because it would dump the Authorization
		// header and the full body. Status + size + attempt is the operator
		// signal; the payload stays out of the logs.
		logger.From(ctx).Infof("openrouter.http status=%d bytes=%d attempt=%d/%d",
			resp.StatusCode, len(resp.Body), attempt+1, c.cfg.MaxRetries+1)

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
	if apiErr.HTTPStatus == 429 {
		// Without a Retry-After there is no safe interval to guess at, and
		// with an excessive one waiting would hold the caller's HTTP
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
	api := &APIError{
		HTTPStatus: resp.StatusCode,
		RetryAfter: parseRetryAfter(resp.Header("Retry-After")),
	}

	var envelope struct {
		Error *wireError `json:"error"`
	}
	if err := json.Unmarshal(resp.Body, &envelope); err != nil || envelope.Error == nil {
		api.Message = fmt.Sprintf("unexpected response (%d bytes, not an OpenRouter error envelope)", len(resp.Body))
		return api
	}

	api.Code = codeString(envelope.Error.Code)
	api.Message = envelope.Error.Message
	return api
}

// sseEvent is one decoded `data:` payload.
type sseEvent = wireStreamChunk

// parseSSE decodes an OpenRouter SSE body into its JSON events, in order.
//
// Per the protocol: lines beginning with ":" are keep-alive comments (e.g.
// ": OPENROUTER PROCESSING") and MUST be skipped before any JSON parsing;
// the terminator is `data: [DONE]`; blank lines separate events.
func parseSSE(body []byte) ([]sseEvent, error) {
	var events []sseEvent

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

		var ev sseEvent
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
