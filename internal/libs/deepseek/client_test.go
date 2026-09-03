package deepseek

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

// This client speaks plain REST, so the natural double is a real httptest
// server: it exercises the actual request encoding, header set, status
// handling and retry loop rather than a hand-rolled transport seam.

func newTestClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	return newTestClientWith(t, handler, func(*Config) {})
}

func newTestClientWith(t *testing.T, handler http.HandlerFunc, mutate func(*Config)) *Client {
	t.Helper()

	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	cfg := Config{
		APIKey:      "test-key",
		BaseURL:     srv.URL,
		Model:       "deepseek-v4-flash",
		Temperature: -1,
		TopP:        -1,
		Timeout:     5 * time.Second,
		MaxRetries:  2,
		RetryDelay:  time.Millisecond,
	}
	mutate(&cfg)

	client, err := NewClient(context.Background(), cfg)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	return client
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
  "model": "deepseek-v4-flash",
  "choices": [{"index": 0, "message": {"role": "assistant", "content": "4"}, "finish_reason": "stop"}],
  "usage": {
    "prompt_tokens": 11, "completion_tokens": 1, "total_tokens": 12,
    "prompt_cache_hit_tokens": 8, "prompt_cache_miss_tokens": 3
  }
}`

func TestGenerateSuccess(t *testing.T) {
	var gotPath, gotAuth string
	var gotBody wireChatRequest

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
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
		t.Errorf("Authorization = %q", gotAuth)
	}
	if gotBody.Model != "deepseek-v4-flash" {
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
	// Thinking is unset by default, so the block must be omitted entirely
	// and the model's own default left in place.
	if gotBody.Thinking != nil {
		t.Errorf("thinking = %+v, want omitted when unconfigured", gotBody.Thinking)
	}

	if res.Content != "4" {
		t.Errorf("Content = %q, want 4", res.Content)
	}
	if res.Usage.PromptTokens != 11 || res.Usage.TotalTokens != 12 {
		t.Errorf("Usage = %+v", res.Usage)
	}
	// Cache counters are DeepSeek extensions kept for cost reporting.
	if res.Usage.CacheHitTokens != 8 || res.Usage.CacheMissTokens != 3 {
		t.Errorf("cache tokens = hit %d / miss %d, want 8 / 3",
			res.Usage.CacheHitTokens, res.Usage.CacheMissTokens)
	}
}

// TestGenerateSendsMaxTokensNotMaxCompletionTokens is the field-name
// difference that makes this client necessary: libs/openai always sends
// max_completion_tokens, which DeepSeek does not read — the cap would
// silently not apply.
func TestGenerateSendsMaxTokensNotMaxCompletionTokens(t *testing.T) {
	var raw map[string]any

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
			t.Errorf("decode request: %v", err)
		}
		writeJSON(t, w, http.StatusOK, okCompletion)
	})

	_, err := client.Generate(context.Background(), ChatRequest{
		Messages:  []Message{{Role: RoleUser, Content: "hi"}},
		MaxTokens: 256,
		JSONMode:  true,
	})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if got, ok := raw["max_tokens"].(float64); !ok || got != 256 {
		t.Errorf("max_tokens = %v, want 256", raw["max_tokens"])
	}
	if _, present := raw["max_completion_tokens"]; present {
		t.Error("max_completion_tokens was sent; DeepSeek does not read it and the cap would not apply")
	}
	format, _ := raw["response_format"].(map[string]any)
	if format == nil || format["type"] != "json_object" {
		t.Errorf("response_format = %v, want json_object", raw["response_format"])
	}
}

func TestGenerateSendsThinking(t *testing.T) {
	var gotBody wireChatRequest

	client := newTestClientWith(t, func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Errorf("decode request: %v", err)
		}
		writeJSON(t, w, http.StatusOK, okCompletion)
	}, func(c *Config) {
		c.Thinking = ThinkingDisabled
		c.ReasoningEffort = ReasoningEffortLow
	})

	if _, err := client.Generate(context.Background(), ChatRequest{
		Messages: []Message{{Role: RoleUser, Content: "hi"}},
	}); err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if gotBody.Thinking == nil {
		t.Fatal("thinking = nil, want the configured block")
	}
	if gotBody.Thinking.Type != ThinkingDisabled {
		t.Errorf("thinking.type = %q, want disabled", gotBody.Thinking.Type)
	}
	if gotBody.Thinking.ReasoningEffort != ReasoningEffortLow {
		t.Errorf("thinking.reasoning_effort = %q, want low", gotBody.Thinking.ReasoningEffort)
	}
}

// TestGenerateCapturesReasoningContent: the field is captured for
// reporting but must never be echoed back on an outbound message.
func TestGenerateCapturesReasoningContent(t *testing.T) {
	var gotBody wireChatRequest

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Errorf("decode request: %v", err)
		}
		writeJSON(t, w, http.StatusOK, `{
		  "model": "deepseek-v4-pro",
		  "choices": [{"index":0,"message":{"role":"assistant","content":"4","reasoning_content":"2+2 is 4"},"finish_reason":"stop"}],
		  "usage": {"prompt_tokens":11,"completion_tokens":9,"total_tokens":20,
		            "completion_tokens_details":{"reasoning_tokens":8}}
		}`)
	})

	res, err := client.Generate(context.Background(), ChatRequest{
		Messages: []Message{
			{Role: RoleAssistant, Content: "earlier answer"},
			{Role: RoleUser, Content: "2+2?"},
		},
	})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if res.Content != "4" {
		t.Errorf("Content = %q, want just the answer", res.Content)
	}
	if res.ReasoningContent != "2+2 is 4" {
		t.Errorf("ReasoningContent = %q", res.ReasoningContent)
	}
	if res.Usage.ReasoningTokens != 8 {
		t.Errorf("ReasoningTokens = %d, want 8", res.Usage.ReasoningTokens)
	}
	for i, m := range gotBody.Messages {
		if m.ReasoningContent != "" {
			t.Errorf("messages[%d] carried reasoning_content outbound; the API rejects it", i)
		}
	}
}

// TestGenerateFinishReasons covers the three finish_reason values that
// mean an HTTP 200 is not a usable answer.
func TestGenerateFinishReasons(t *testing.T) {
	tests := []struct {
		reason string
		wantIs error
		want   string
	}{
		{finishReasonFilter, ErrContentFiltered, "content_filter"},
		{finishReasonNoResource, ErrServerOverloaded, "insufficient_system_resource"},
		{finishReasonLength, ErrDecodeResponse, "MaxTokens"},
	}

	for _, tt := range tests {
		t.Run(tt.reason, func(t *testing.T) {
			client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
				writeJSON(t, w, http.StatusOK, `{
				  "model": "deepseek-v4-flash",
				  "choices": [{"index":0,"message":{"role":"assistant","content":"{\"q\": ["},"finish_reason":"`+tt.reason+`"}],
				  "usage": {"prompt_tokens":11,"completion_tokens":256,"total_tokens":267}
				}`)
			})

			_, err := client.Generate(context.Background(), ChatRequest{
				Messages: []Message{{Role: RoleUser, Content: "hi"}},
			})
			if !errors.Is(err, tt.wantIs) {
				t.Fatalf("error = %v, want %v", err, tt.wantIs)
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Errorf("error = %v, want it to mention %q", err, tt.want)
			}
		})
	}
}

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
			name:       "401 bad key",
			status:     http.StatusUnauthorized,
			body:       `{"error":{"message":"Authentication Fails","type":"authentication_error"}}`,
			wantAuth:   true,
			wantConfig: true,
			wantCalls:  1,
		},
		{
			name:       "402 insufficient balance",
			status:     http.StatusPaymentRequired,
			body:       `{"error":{"message":"Insufficient Balance","type":"insufficient_balance"}}`,
			wantIs:     ErrInsufficientBalance,
			wantConfig: true,
			wantCalls:  1,
		},
		{
			name:      "400 invalid format",
			status:    http.StatusBadRequest,
			body:      `{"error":{"message":"Invalid request body","type":"invalid_request_error"}}`,
			wantCalls: 1,
		},
		{
			name:      "422 invalid parameters",
			status:    http.StatusUnprocessableEntity,
			body:      `{"error":{"message":"Invalid parameters","type":"invalid_request_error"}}`,
			wantCalls: 1,
		},
		{
			name:      "429 without retry-after is surfaced",
			status:    http.StatusTooManyRequests,
			body:      `{"error":{"message":"Rate Limit Reached","type":"rate_limit_error"}}`,
			wantIs:    ErrRateLimited,
			wantCalls: 1,
		},
		{
			name:      "500 exhausts the retry budget",
			status:    http.StatusInternalServerError,
			body:      `{"error":{"message":"Server Error","type":"server_error"}}`,
			wantCalls: 3,
		},
		{
			name:      "503 exhausts the retry budget",
			status:    http.StatusServiceUnavailable,
			body:      `{"error":{"message":"Server Overloaded","type":"server_overloaded"}}`,
			wantIs:    ErrServerOverloaded,
			wantCalls: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var calls int32
			client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
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

func TestGenerateRetriesWithRetryAfter(t *testing.T) {
	var calls int32
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&calls, 1) == 1 {
			w.Header().Set("Retry-After", "1")
			writeJSON(t, w, http.StatusTooManyRequests,
				`{"error":{"message":"Rate Limit Reached","type":"rate_limit_error"}}`)
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

func TestErrorBodyIsNotLeaked(t *testing.T) {
	const secret = "<html>gateway: api_key=sk-should-never-surface</html>"

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
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
	if strings.Contains(err.Error(), "sk-should-never-surface") || strings.Contains(err.Error(), "<html>") {
		t.Fatalf("raw upstream body leaked into the error: %v", err)
	}
	if !strings.Contains(err.Error(), "not a DeepSeek error envelope") {
		t.Errorf("error = %v, want the size-only fallback message", err)
	}
}

func TestGenerateEmptyMessages(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("upstream must not be called for an empty message list")
	})

	if _, err := client.Generate(context.Background(), ChatRequest{}); !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("error = %v, want ErrInvalidConfig", err)
	}
}

// ---------------------------------------------------------------------------
// Streaming
// ---------------------------------------------------------------------------

// A thinking-mode stream interleaves reasoning deltas with answer deltas.
const sseBody = `: keep-alive

data: {"model":"deepseek-v4-pro","choices":[{"index":0,"delta":{"role":"assistant","content":""}}]}

data: {"model":"deepseek-v4-pro","choices":[{"index":0,"delta":{"reasoning_content":"let me think"}}]}

data: {"model":"deepseek-v4-pro","choices":[{"index":0,"delta":{"content":"He"}}]}

data: {"model":"deepseek-v4-pro","choices":[{"index":0,"delta":{"content":"llo"}}]}

data: {"model":"deepseek-v4-pro","choices":[{"index":0,"delta":{},"finish_reason":"stop"}]}

data: {"model":"deepseek-v4-pro","choices":[],"usage":{"prompt_tokens":7,"completion_tokens":2,"total_tokens":9,"completion_tokens_details":{"reasoning_tokens":5}}}

data: [DONE]

`

func TestGenerateStream(t *testing.T) {
	var gotBody wireChatRequest

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
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
		t.Error("stream = false, want true")
	}
	if gotBody.StreamOptions == nil || !gotBody.StreamOptions.IncludeUsage {
		t.Errorf("stream_options = %+v, want include_usage", gotBody.StreamOptions)
	}

	// Reasoning deltas must NOT reach the consumer — mixing them into the
	// answer stream would corrupt the JSON the caller assembles.
	if len(deltas) != 2 || deltas[0] != "He" || deltas[1] != "llo" {
		t.Errorf("deltas = %q, want [He llo] with no reasoning text", deltas)
	}
	if doneCount != 1 {
		t.Errorf("Done chunks = %d, want exactly 1", doneCount)
	}
	if doneUsage == nil || doneUsage.ReasoningTokens != 5 {
		t.Errorf("final Usage = %+v, want reasoning_tokens 5", doneUsage)
	}
	if res.Content != "Hello" {
		t.Errorf("Content = %q, want Hello", res.Content)
	}
	// It is still accumulated on the response for reporting.
	if res.ReasoningContent != "let me think" {
		t.Errorf("ReasoningContent = %q", res.ReasoningContent)
	}
	if res.FinishReason != finishReasonStop {
		t.Errorf("FinishReason = %q, want stop", res.FinishReason)
	}
}

func TestGenerateStreamAbortedMidStream(t *testing.T) {
	tests := []struct {
		reason string
		wantIs error
	}{
		{finishReasonFilter, ErrContentFiltered},
		{finishReasonNoResource, ErrServerOverloaded},
	}

	for _, tt := range tests {
		t.Run(tt.reason, func(t *testing.T) {
			client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/event-stream")
				w.WriteHeader(http.StatusOK)
				body := `data: {"choices":[{"index":0,"delta":{"content":"par"}}]}

data: {"choices":[{"index":0,"delta":{},"finish_reason":"` + tt.reason + `"}]}

data: [DONE]
`
				if _, err := w.Write([]byte(body)); err != nil {
					t.Errorf("write response: %v", err)
				}
			})

			_, err := client.GenerateStream(context.Background(), ChatRequest{
				Messages: []Message{{Role: RoleUser, Content: "hi"}},
			}, func(StreamChunk) error { return nil })

			if !errors.Is(err, tt.wantIs) {
				t.Fatalf("error = %v, want %v", err, tt.wantIs)
			}
		})
	}
}

func TestParseSSE(t *testing.T) {
	t.Run("skips comments and stops at DONE", func(t *testing.T) {
		events, err := parseSSE([]byte(sseBody))
		if err != nil {
			t.Fatalf("parseSSE() error = %v", err)
		}
		if len(events) != 6 {
			t.Fatalf("events = %d, want 6", len(events))
		}
	})

	t.Run("empty stream", func(t *testing.T) {
		if _, err := parseSSE([]byte("data: [DONE]\n")); !errors.Is(err, ErrDecodeResponse) {
			t.Fatalf("error = %v, want ErrDecodeResponse", err)
		}
	})

	t.Run("malformed chunk", func(t *testing.T) {
		if _, err := parseSSE([]byte("data: {not json\n")); !errors.Is(err, ErrDecodeResponse) {
			t.Fatalf("error = %v, want ErrDecodeResponse", err)
		}
	})
}

func TestNewClientProbesWhenRequiredAtBoot(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		writeJSON(t, w, http.StatusUnauthorized,
			`{"error":{"message":"Authentication Fails","type":"authentication_error"}}`)
	}))
	t.Cleanup(srv.Close)

	base := Config{
		APIKey:      "bad-key",
		BaseURL:     srv.URL,
		Model:       "deepseek-v4-flash",
		Temperature: -1,
		TopP:        -1,
		Timeout:     5 * time.Second,
		MaxRetries:  0,
		RetryDelay:  time.Millisecond,
	}

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
