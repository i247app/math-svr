package openrouter

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// The OpenRouter client speaks plain REST, so the natural double is a real
// httptest server: it exercises the actual request encoding, header set,
// status handling and retry loop rather than a hand-rolled transport seam.

// newTestClient wires a Client against handler. RetryDelay is 1ms so the
// retry-budget cases stay fast.
func newTestClient(t *testing.T, handler http.HandlerFunc) (*Client, *httptest.Server) {
	t.Helper()

	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	client, err := NewClient(context.Background(), Config{
		APIKey:      "test-key",
		BaseURL:     srv.URL,
		Model:       "openai/gpt-4o-mini",
		SiteURL:     "https://math-ai.example",
		AppTitle:    "math-svr-test",
		Temperature: -1,
		TopP:        -1,
		Timeout:     5 * time.Second,
		MaxRetries:  2,
		RetryDelay:  time.Millisecond,
	})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	return client, srv
}

func writeJSON(t *testing.T, w http.ResponseWriter, status int, body string) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if _, err := w.Write([]byte(body)); err != nil {
		t.Errorf("write response: %v", err)
	}
}

const okCompletion = `{
  "model": "openai/gpt-4o-mini",
  "choices": [{"index": 0, "message": {"role": "assistant", "content": "4"}, "finish_reason": "stop"}],
  "usage": {"prompt_tokens": 11, "completion_tokens": 1, "total_tokens": 12}
}`

func TestGenerateSuccess(t *testing.T) {
	var gotPath, gotAuth, gotReferer, gotTitle string
	var gotBody wireChatRequest

	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		gotReferer = r.Header.Get("HTTP-Referer")
		gotTitle = r.Header.Get("X-OpenRouter-Title")
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Errorf("decode request: %v", err)
		}
		writeJSON(t, w, http.StatusOK, okCompletion)
	})

	res, err := client.Generate(context.Background(), ChatRequest{
		Messages:    []Message{{Role: RoleUser, Content: "2+2?"}},
		Temperature: -1,
		TopP:        -1,
	})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if gotPath != "/chat/completions" {
		t.Errorf("path = %q, want /chat/completions", gotPath)
	}
	if gotAuth != "Bearer test-key" {
		t.Errorf("Authorization = %q, want Bearer test-key", gotAuth)
	}
	if gotReferer != "https://math-ai.example" {
		t.Errorf("HTTP-Referer = %q", gotReferer)
	}
	if gotTitle != "math-svr-test" {
		t.Errorf("X-OpenRouter-Title = %q", gotTitle)
	}
	if gotBody.Model != "openai/gpt-4o-mini" {
		t.Errorf("model = %q, want the configured default", gotBody.Model)
	}
	if gotBody.Stream {
		t.Error("stream = true, want false on Generate")
	}
	// Sampling overrides left at "use default" must not reach the wire —
	// sending temperature:0 would be a real, and very different, request.
	if gotBody.Temperature != nil || gotBody.TopP != nil || gotBody.MaxTokens != nil {
		t.Errorf("unset sampling fields leaked to the wire: %+v", gotBody)
	}
	if gotBody.ResponseFormat != nil {
		t.Error("response_format sent without JSONMode")
	}
	if len(gotBody.Messages) != 1 || gotBody.Messages[0].Role != "user" {
		t.Errorf("messages = %+v", gotBody.Messages)
	}

	if res.Content != "4" {
		t.Errorf("Content = %q, want 4", res.Content)
	}
	if res.FinishReason != "stop" {
		t.Errorf("FinishReason = %q, want stop", res.FinishReason)
	}
	if res.Usage != (Usage{PromptTokens: 11, CompletionTokens: 1, TotalTokens: 12}) {
		t.Errorf("Usage = %+v", res.Usage)
	}
}

func TestGenerateSendsOverrides(t *testing.T) {
	var gotBody wireChatRequest

	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Errorf("decode request: %v", err)
		}
		writeJSON(t, w, http.StatusOK, okCompletion)
	})

	_, err := client.Generate(context.Background(), ChatRequest{
		Model:       "anthropic/claude-3.5-sonnet",
		Messages:    []Message{{Role: RoleSystem, Content: "be terse"}, {Role: RoleUser, Content: "hi"}},
		Temperature: 0.2,
		TopP:        0.9,
		MaxTokens:   256,
		Stop:        []string{"###"},
		JSONMode:    true,
	})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if gotBody.Model != "anthropic/claude-3.5-sonnet" {
		t.Errorf("model = %q, want the per-call override", gotBody.Model)
	}
	if gotBody.Temperature == nil || *gotBody.Temperature != 0.2 {
		t.Errorf("temperature = %v, want 0.2", gotBody.Temperature)
	}
	if gotBody.TopP == nil || *gotBody.TopP != 0.9 {
		t.Errorf("top_p = %v, want 0.9", gotBody.TopP)
	}
	if gotBody.MaxTokens == nil || *gotBody.MaxTokens != 256 {
		t.Errorf("max_tokens = %v, want 256", gotBody.MaxTokens)
	}
	if len(gotBody.Stop) != 1 || gotBody.Stop[0] != "###" {
		t.Errorf("stop = %v", gotBody.Stop)
	}
	if gotBody.ResponseFormat == nil || gotBody.ResponseFormat.Type != "json_object" {
		t.Errorf("response_format = %+v, want json_object", gotBody.ResponseFormat)
	}
}

func TestGenerateEmptyMessages(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("upstream must not be called for an empty message list")
	})

	_, err := client.Generate(context.Background(), ChatRequest{})
	if !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("error = %v, want ErrInvalidConfig", err)
	}
}

// TestGenerateErrorStatuses pins the HTTP-status → sentinel taxonomy that
// adapter/bot.mapOpenRouterError switches on.
func TestGenerateErrorStatuses(t *testing.T) {
	tests := []struct {
		name       string
		status     int
		body       string
		wantIs     error
		wantAuth   bool
		wantConfig bool
		wantCalls  int32
	}{
		{
			name:       "401 auth",
			status:     http.StatusUnauthorized,
			body:       `{"error":{"code":401,"message":"No auth credentials found"}}`,
			wantAuth:   true,
			wantConfig: true,
			wantCalls:  1,
		},
		{
			name:       "402 out of credits",
			status:     http.StatusPaymentRequired,
			body:       `{"error":{"code":402,"message":"Insufficient credits"}}`,
			wantIs:     ErrInsufficientCredits,
			wantConfig: true,
			wantCalls:  1,
		},
		{
			name:       "403 moderation block",
			status:     http.StatusForbidden,
			body:       `{"error":{"code":403,"message":"Flagged by moderation"}}`,
			wantAuth:   true,
			wantConfig: true,
			wantCalls:  1,
		},
		{
			name:       "404 unknown model",
			status:     http.StatusNotFound,
			body:       `{"error":{"code":404,"message":"No endpoints found for model"}}`,
			wantConfig: true,
			wantCalls:  1,
		},
		{
			name:      "400 context too large",
			status:    http.StatusBadRequest,
			body:      `{"error":{"code":400,"message":"This model's maximum context length is 8192 tokens"}}`,
			wantIs:    ErrContextTooLarge,
			wantCalls: 1,
		},
		{
			name:      "400 plain invalid request",
			status:    http.StatusBadRequest,
			body:      `{"error":{"code":400,"message":"messages must be an array"}}`,
			wantCalls: 1,
		},
		{
			name:      "429 without retry-after is surfaced",
			status:    http.StatusTooManyRequests,
			body:      `{"error":{"code":429,"message":"Rate limit exceeded"}}`,
			wantIs:    ErrRateLimited,
			wantCalls: 1,
		},
		{
			// 408 / 5xx are transient: the full budget is spent first.
			name:      "503 exhausts the retry budget",
			status:    http.StatusServiceUnavailable,
			body:      `{"error":{"code":503,"message":"No instances available"}}`,
			wantCalls: 3,
		},
		{
			name:      "408 exhausts the retry budget",
			status:    http.StatusRequestTimeout,
			body:      `{"error":{"code":408,"message":"Request timed out"}}`,
			wantCalls: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var calls int32
			client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
				atomic.AddInt32(&calls, 1)
				writeJSON(t, w, tt.status, tt.body)
			})

			_, err := client.Generate(context.Background(), ChatRequest{
				Messages: []Message{{Role: RoleUser, Content: "hi"}},
			})
			if err == nil {
				t.Fatal("Generate() error = nil, want error")
			}

			if got := atomic.LoadInt32(&calls); got != tt.wantCalls {
				t.Errorf("upstream calls = %d, want %d", got, tt.wantCalls)
			}
			if tt.wantIs != nil && !errors.Is(err, tt.wantIs) {
				t.Errorf("error %v does not wrap %v", err, tt.wantIs)
			}
			if got := IsAuthError(err); got != tt.wantAuth {
				t.Errorf("IsAuthError() = %v, want %v", got, tt.wantAuth)
			}
			if got := IsConfigError(err); got != tt.wantConfig {
				t.Errorf("IsConfigError() = %v, want %v", got, tt.wantConfig)
			}

			var api *APIError
			if !errors.As(err, &api) {
				t.Fatalf("error %v is not an *APIError", err)
			}
			if api.HTTPStatus != tt.status {
				t.Errorf("HTTPStatus = %d, want %d", api.HTTPStatus, tt.status)
			}
		})
	}
}

func TestGenerateRetriesRateLimitWithRetryAfter(t *testing.T) {
	var calls int32
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&calls, 1) == 1 {
			w.Header().Set("Retry-After", "1")
			writeJSON(t, w, http.StatusTooManyRequests,
				`{"error":{"code":429,"message":"Rate limit exceeded"}}`)
			return
		}
		writeJSON(t, w, http.StatusOK, okCompletion)
	})

	res, err := client.Generate(context.Background(), ChatRequest{
		Messages: []Message{{Role: RoleUser, Content: "hi"}},
	})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if res.Content != "4" {
		t.Errorf("Content = %q, want 4", res.Content)
	}
	if got := atomic.LoadInt32(&calls); got != 2 {
		t.Errorf("upstream calls = %d, want 2 (one 429 + one retry)", got)
	}
}

func TestRetryDelayFor(t *testing.T) {
	base := 500 * time.Millisecond
	tests := []struct {
		name    string
		api     *APIError
		wantOK  bool
		wantDur time.Duration
	}{
		{"nil", nil, false, 0},
		{"429 no retry-after", &APIError{HTTPStatus: 429}, false, 0},
		{"429 with retry-after", &APIError{HTTPStatus: 429, RetryAfter: 2 * time.Second}, true, 2 * time.Second},
		{"429 retry-after too long", &APIError{HTTPStatus: 429, RetryAfter: maxRetryAfter + time.Second}, false, 0},
		{"503", &APIError{HTTPStatus: 503}, true, base},
		{"401", &APIError{HTTPStatus: 401}, false, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := retryDelayFor(tt.api, base)
			if ok != tt.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tt.wantOK)
			}
			if ok && got != tt.wantDur {
				t.Errorf("delay = %v, want %v", got, tt.wantDur)
			}
		})
	}
}

// TestGenerateNonJSONErrorBodyIsNotLeaked guards the log/leak discipline:
// an unexpected upstream body (proxy HTML, WAF block page) must be
// reported by size, never echoed into the error that travels to the
// MathError debug field.
func TestGenerateNonJSONErrorBodyIsNotLeaked(t *testing.T) {
	const secret = "<html>gateway says: api_key=sk-should-never-surface</html>"

	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		if _, err := w.Write([]byte(secret)); err != nil {
			t.Errorf("write response: %v", err)
		}
	})

	_, err := client.Generate(context.Background(), ChatRequest{
		Messages: []Message{{Role: RoleUser, Content: "hi"}},
	})
	if err == nil {
		t.Fatal("Generate() error = nil, want error")
	}
	if strings.Contains(err.Error(), "sk-should-never-surface") ||
		strings.Contains(err.Error(), "<html>") {
		t.Fatalf("raw upstream body leaked into the error: %v", err)
	}
	if !strings.Contains(err.Error(), "not an OpenRouter error envelope") {
		t.Errorf("error = %v, want the size-only fallback message", err)
	}
}

// TestGenerateErrorEnvelopeOn200 covers OpenRouter answering 200 and then
// reporting the routed upstream's failure in the body.
func TestGenerateErrorEnvelopeOn200(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, http.StatusOK,
			`{"error":{"code":429,"message":"Provider returned rate limit"}}`)
	})

	_, err := client.Generate(context.Background(), ChatRequest{
		Messages: []Message{{Role: RoleUser, Content: "hi"}},
	})
	if !errors.Is(err, ErrRateLimited) {
		t.Fatalf("error = %v, want ErrRateLimited", err)
	}
}

func TestGenerateFinishReasonError(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, http.StatusOK, `{
		  "model": "openai/gpt-4o-mini",
		  "choices": [{"index":0,"message":{"role":"assistant","content":"par"},"finish_reason":"error"}]
		}`)
	})

	_, err := client.Generate(context.Background(), ChatRequest{
		Messages: []Message{{Role: RoleUser, Content: "hi"}},
	})
	if err == nil {
		t.Fatal("Generate() error = nil, want error on finish_reason=error")
	}
	if !strings.Contains(err.Error(), "finish_reason=error") {
		t.Errorf("error = %v, want a finish_reason=error report", err)
	}
}

func TestGenerateNoChoices(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, http.StatusOK, `{"model":"openai/gpt-4o-mini","choices":[]}`)
	})

	_, err := client.Generate(context.Background(), ChatRequest{
		Messages: []Message{{Role: RoleUser, Content: "hi"}},
	})
	if !errors.Is(err, ErrDecodeResponse) {
		t.Fatalf("error = %v, want ErrDecodeResponse", err)
	}
}

func TestGenerateUndecodableSuccessBody(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, http.StatusOK, `not json at all`)
	})

	_, err := client.Generate(context.Background(), ChatRequest{
		Messages: []Message{{Role: RoleUser, Content: "hi"}},
	})
	if !errors.Is(err, ErrDecodeResponse) {
		t.Fatalf("error = %v, want ErrDecodeResponse", err)
	}
}

// ---------------------------------------------------------------------------
// Streaming
// ---------------------------------------------------------------------------

// sseBody is a realistic OpenRouter SSE transcript: a keep-alive comment
// line, role-only first delta, two content deltas, a finish chunk, a
// usage-only chunk, and the [DONE] terminator.
const sseBody = `: OPENROUTER PROCESSING

data: {"model":"openai/gpt-4o-mini","choices":[{"index":0,"delta":{"role":"assistant","content":""}}]}

data: {"model":"openai/gpt-4o-mini","choices":[{"index":0,"delta":{"content":"He"}}]}

: OPENROUTER PROCESSING

data: {"model":"openai/gpt-4o-mini","choices":[{"index":0,"delta":{"content":"llo"}}]}

data: {"model":"openai/gpt-4o-mini","choices":[{"index":0,"delta":{"content":""},"finish_reason":"stop"}]}

data: {"model":"openai/gpt-4o-mini","choices":[],"usage":{"prompt_tokens":7,"completion_tokens":2,"total_tokens":9}}

data: [DONE]

`

func TestGenerateStream(t *testing.T) {
	var gotBody wireChatRequest

	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Errorf("decode request: %v", err)
		}
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(sseBody)); err != nil {
			t.Errorf("write response: %v", err)
		}
	})

	var deltas []string
	var doneCount int
	var doneUsage *Usage

	res, err := client.GenerateStream(context.Background(), ChatRequest{
		Messages: []Message{{Role: RoleUser, Content: "hi"}},
	}, func(c StreamChunk) error {
		if c.Done {
			doneCount++
			doneUsage = c.Usage
			return nil
		}
		deltas = append(deltas, c.Delta)
		return nil
	})
	if err != nil {
		t.Fatalf("GenerateStream() error = %v", err)
	}

	if !gotBody.Stream {
		t.Error("stream = false, want true on GenerateStream")
	}
	if gotBody.StreamOptions == nil || !gotBody.StreamOptions.IncludeUsage {
		t.Errorf("stream_options = %+v, want include_usage", gotBody.StreamOptions)
	}

	// Empty deltas (the role-only and finish chunks) must not be forwarded.
	if len(deltas) != 2 || deltas[0] != "He" || deltas[1] != "llo" {
		t.Errorf("deltas = %q, want [He llo]", deltas)
	}
	if doneCount != 1 {
		t.Errorf("Done chunks = %d, want exactly 1", doneCount)
	}
	if doneUsage == nil || *doneUsage != (Usage{PromptTokens: 7, CompletionTokens: 2, TotalTokens: 9}) {
		t.Errorf("final Usage = %+v", doneUsage)
	}
	if res.Content != "Hello" {
		t.Errorf("Content = %q, want Hello", res.Content)
	}
	if res.FinishReason != "stop" {
		t.Errorf("FinishReason = %q, want stop", res.FinishReason)
	}
}

// TestGenerateStreamMidStreamError covers the documented trap: the
// transport stays 200 and the failure is only visible as
// finish_reason:"error" inside a chunk.
func TestGenerateStreamMidStreamError(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		body := `data: {"choices":[{"index":0,"delta":{"content":"par"}}]}

data: {"choices":[{"index":0,"delta":{},"finish_reason":"error"}]}

data: [DONE]
`
		if _, err := w.Write([]byte(body)); err != nil {
			t.Errorf("write response: %v", err)
		}
	})

	_, err := client.GenerateStream(context.Background(), ChatRequest{
		Messages: []Message{{Role: RoleUser, Content: "hi"}},
	}, func(StreamChunk) error { return nil })

	if err == nil {
		t.Fatal("GenerateStream() error = nil, want a mid-stream failure")
	}
	if !strings.Contains(err.Error(), "finish_reason=error") {
		t.Errorf("error = %v, want a finish_reason=error report", err)
	}
}

func TestGenerateStreamRequiresCallback(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("upstream must not be called without an onChunk callback")
	})

	_, err := client.GenerateStream(context.Background(), ChatRequest{
		Messages: []Message{{Role: RoleUser, Content: "hi"}},
	}, nil)
	if !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("error = %v, want ErrInvalidConfig", err)
	}
}

func TestParseSSE(t *testing.T) {
	t.Run("skips comments and stops at DONE", func(t *testing.T) {
		events, err := parseSSE([]byte(sseBody))
		if err != nil {
			t.Fatalf("parseSSE() error = %v", err)
		}
		if len(events) != 5 {
			t.Fatalf("events = %d, want 5", len(events))
		}
	})

	t.Run("empty stream", func(t *testing.T) {
		_, err := parseSSE([]byte("data: [DONE]\n"))
		if !errors.Is(err, ErrDecodeResponse) {
			t.Fatalf("error = %v, want ErrDecodeResponse", err)
		}
	})

	t.Run("malformed chunk", func(t *testing.T) {
		_, err := parseSSE([]byte("data: {not json\n"))
		if !errors.Is(err, ErrDecodeResponse) {
			t.Fatalf("error = %v, want ErrDecodeResponse", err)
		}
	})
}

func TestNewClientProbesWhenRequiredAtBoot(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		writeJSON(t, w, http.StatusUnauthorized,
			`{"error":{"code":401,"message":"No auth credentials found"}}`)
	}))
	t.Cleanup(srv.Close)

	base := Config{
		APIKey:      "bad-key",
		BaseURL:     srv.URL,
		Model:       "openai/gpt-4o-mini",
		Temperature: -1,
		TopP:        -1,
		Timeout:     5 * time.Second,
		MaxRetries:  0,
		RetryDelay:  time.Millisecond,
	}

	// RequireAtBoot=false must not touch the network at all.
	if _, err := NewClient(context.Background(), base); err != nil {
		t.Fatalf("NewClient(RequireAtBoot=false) error = %v", err)
	}
	if got := atomic.LoadInt32(&calls); got != 0 {
		t.Fatalf("upstream calls = %d, want 0 when the probe is skipped", got)
	}

	base.RequireAtBoot = true
	if _, err := NewClient(context.Background(), base); err == nil {
		t.Fatal("NewClient(RequireAtBoot=true) error = nil, want the bad credential to fail boot")
	}
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Errorf("upstream calls = %d, want 1 probe", got)
	}
}
