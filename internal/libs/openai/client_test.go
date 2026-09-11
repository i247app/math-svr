package openai

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
		Model:       "gpt-4.1",
		EmbedModel:  "text-embedding-3-small",
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
  "model": "gpt-4.1",
  "choices": [{"index": 0, "message": {"role": "assistant", "content": "4"}, "finish_reason": "stop"}],
  "usage": {"prompt_tokens": 11, "completion_tokens": 1, "total_tokens": 12}
}`

func TestGenerateSuccess(t *testing.T) {
	var gotPath, gotAuth, gotOrg, gotProject string
	var gotBody wireChatRequest

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		gotOrg = r.Header.Get("OpenAI-Organization")
		gotProject = r.Header.Get("OpenAI-Project")
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
	// Neither header is configured in the base test config, so neither may
	// be sent — an empty OpenAI-Organization is not the same as absent.
	if gotOrg != "" || gotProject != "" {
		t.Errorf("unconfigured scoping headers leaked: org=%q project=%q", gotOrg, gotProject)
	}
	if gotBody.Model != "gpt-4.1" {
		t.Errorf("model = %q, want the configured default", gotBody.Model)
	}
	if gotBody.Stream {
		t.Error("stream = true, want false on Generate")
	}
	if gotBody.Store {
		t.Error("store = true, want false by default")
	}
	// Sampling overrides left at "use default" must not reach the wire —
	// sending temperature:0 would be a real, and very different, request.
	if gotBody.Temperature != nil || gotBody.TopP != nil || gotBody.MaxCompletionTokens != nil {
		t.Errorf("unset sampling fields leaked to the wire: %+v", gotBody)
	}

	if res.Content != "4" {
		t.Errorf("Content = %q, want 4", res.Content)
	}
	if res.Usage != (Usage{PromptTokens: 11, CompletionTokens: 1, TotalTokens: 12}) {
		t.Errorf("Usage = %+v", res.Usage)
	}
}

// TestGenerateSendsMaxCompletionTokens pins the deprecation workaround:
// `max_tokens` is rejected outright by the reasoning models, so the client
// must always send `max_completion_tokens` instead.
func TestGenerateSendsMaxCompletionTokens(t *testing.T) {
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

	if _, present := raw["max_tokens"]; present {
		t.Error("deprecated max_tokens was sent; reasoning models reject it")
	}
	if got, ok := raw["max_completion_tokens"].(float64); !ok || got != 256 {
		t.Errorf("max_completion_tokens = %v, want 256", raw["max_completion_tokens"])
	}
	format, _ := raw["response_format"].(map[string]any)
	if format == nil || format["type"] != "json_object" {
		t.Errorf("response_format = %v, want json_object", raw["response_format"])
	}
}

func TestGenerateStoreAndMetadata(t *testing.T) {
	var gotBody wireChatRequest

	client := newTestClientWith(t, func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Errorf("decode request: %v", err)
		}
		writeJSON(t, w, http.StatusOK, okCompletion)
	}, func(c *Config) {
		c.Store = true
		c.Metadata = map[string]string{"env": "production"}
		c.Organization = "org-abc"
		c.Project = "proj-xyz"
	})

	_, err := client.Generate(context.Background(), ChatRequest{
		Messages: []Message{{Role: RoleUser, Content: "hi"}},
	})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if !gotBody.Store {
		t.Error("store = false, want true — this is what makes the call visible at platform.openai.com/logs")
	}
	if gotBody.Metadata["env"] != "production" {
		t.Errorf("metadata = %v, want env=production", gotBody.Metadata)
	}
}

// TestMetadataWithoutStoreIsNotSent: metadata is only retained alongside a
// stored completion, so sending it otherwise is noise on the wire.
func TestMetadataWithoutStoreIsNotSent(t *testing.T) {
	var gotBody wireChatRequest

	client := newTestClientWith(t, func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Errorf("decode request: %v", err)
		}
		writeJSON(t, w, http.StatusOK, okCompletion)
	}, func(c *Config) {
		c.Store = false
		c.Metadata = map[string]string{"env": "production"}
	})

	if _, err := client.Generate(context.Background(), ChatRequest{
		Messages: []Message{{Role: RoleUser, Content: "hi"}},
	}); err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if gotBody.Metadata != nil {
		t.Errorf("metadata = %v, want nil when store is off", gotBody.Metadata)
	}
}

// TestGenerateTruncatedCompletion: finish_reason=length means the body is
// cut off mid-JSON. Reporting it as a decode failure is what turns an
// inscrutable downstream JSON parse error into an actionable one.
func TestGenerateTruncatedCompletion(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, http.StatusOK, `{
		  "model": "gpt-4.1",
		  "choices": [{"index":0,"message":{"role":"assistant","content":"{\"questions\": ["},"finish_reason":"length"}],
		  "usage": {"prompt_tokens": 11, "completion_tokens": 256, "total_tokens": 267}
		}`)
	})

	_, err := client.Generate(context.Background(), ChatRequest{
		Messages: []Message{{Role: RoleUser, Content: "hi"}},
	})
	if !errors.Is(err, ErrDecodeResponse) {
		t.Fatalf("error = %v, want ErrDecodeResponse", err)
	}
	if !strings.Contains(err.Error(), "MaxTokens") {
		t.Errorf("error = %v, want it to name the knob to raise", err)
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
			body:       `{"error":{"message":"Incorrect API key","type":"invalid_authentication_error","code":"invalid_api_key"}}`,
			wantAuth:   true,
			wantConfig: true,
			wantCalls:  1,
		},
		{
			name:       "403 unsupported region",
			status:     http.StatusForbidden,
			body:       `{"error":{"message":"Country not supported","type":"request_forbidden"}}`,
			wantAuth:   true,
			wantConfig: true,
			wantCalls:  1,
		},
		{
			name:       "404 unknown model",
			status:     http.StatusNotFound,
			body:       `{"error":{"message":"The model does not exist","type":"invalid_request_error","code":"model_not_found"}}`,
			wantConfig: true,
			wantCalls:  1,
		},
		{
			name:      "400 context length",
			status:    http.StatusBadRequest,
			body:      `{"error":{"message":"maximum context length is 128000 tokens","type":"invalid_request_error","code":"context_length_exceeded"}}`,
			wantIs:    ErrContextTooLarge,
			wantCalls: 1,
		},
		{
			name:      "429 throttling without retry-after is surfaced",
			status:    http.StatusTooManyRequests,
			body:      `{"error":{"message":"Rate limit reached","type":"rate_limit_error"}}`,
			wantIs:    ErrRateLimited,
			wantCalls: 1,
		},
		{
			// The whole point of the 429 split: billing must not burn the
			// retry budget on a call that can never succeed.
			name:       "429 billing is not retried",
			status:     http.StatusTooManyRequests,
			body:       `{"error":{"message":"Credit balance exhausted","type":"rate_limit_error","code":"credit_balance_exhausted"}}`,
			wantIs:     ErrQuotaExhausted,
			wantConfig: true,
			wantCalls:  1,
		},
		{
			name:      "500 exhausts the retry budget",
			status:    http.StatusInternalServerError,
			body:      `{"error":{"message":"server error","type":"server_error"}}`,
			wantCalls: 3,
		},
		{
			name:      "503 exhausts the retry budget",
			status:    http.StatusServiceUnavailable,
			body:      `{"error":{"message":"model overloaded","type":"service_unavailable_error"}}`,
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

func TestGenerateRetriesThrottlingWithRetryAfter(t *testing.T) {
	var calls int32
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&calls, 1) == 1 {
			w.Header().Set("Retry-After", "1")
			writeJSON(t, w, http.StatusTooManyRequests,
				`{"error":{"message":"Rate limit reached","type":"rate_limit_error"}}`)
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

// TestErrorBodyIsNotLeaked guards the log/leak discipline: an unexpected
// upstream body (proxy HTML, WAF block page) must be reported by size,
// never echoed into the error that travels to the MathError debug field.
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
	if !strings.Contains(err.Error(), "not an OpenAI error envelope") {
		t.Errorf("error = %v, want the size-only fallback message", err)
	}
}

// ---------------------------------------------------------------------------
// Embeddings — the capability that only this provider and langchain have
// ---------------------------------------------------------------------------

func TestEmbed(t *testing.T) {
	var gotPath string
	var gotBody wireEmbedRequest

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Errorf("decode request: %v", err)
		}
		// Deliberately OUT of order: the reference documents an index field
		// but never promises the array is sorted.
		writeJSON(t, w, http.StatusOK, `{
		  "object": "list",
		  "model": "text-embedding-3-small",
		  "data": [
		    {"object":"embedding","index":1,"embedding":[0.3,0.4]},
		    {"object":"embedding","index":0,"embedding":[0.1,0.2]}
		  ],
		  "usage": {"prompt_tokens": 6, "total_tokens": 6}
		}`)
	})

	res, err := client.Embed(context.Background(), EmbedRequest{
		Inputs: []string{"first", "second"},
	})
	if err != nil {
		t.Fatalf("Embed() error = %v", err)
	}

	if gotPath != "/embeddings" {
		t.Errorf("path = %q, want /embeddings", gotPath)
	}
	if gotBody.Model != "text-embedding-3-small" {
		t.Errorf("model = %q, want the configured embed model", gotBody.Model)
	}
	if gotBody.EncodingFormat != "float" {
		t.Errorf("encoding_format = %q, want float", gotBody.EncodingFormat)
	}

	// Vectors[i] must pair with Inputs[i]; trusting array order here would
	// silently attach the wrong embedding to the wrong text.
	if len(res.Vectors) != 2 {
		t.Fatalf("vectors = %d, want 2", len(res.Vectors))
	}
	if res.Vectors[0][0] != 0.1 {
		t.Errorf("Vectors[0] = %v, want the index=0 embedding [0.1 0.2]", res.Vectors[0])
	}
	if res.Vectors[1][0] != 0.3 {
		t.Errorf("Vectors[1] = %v, want the index=1 embedding [0.3 0.4]", res.Vectors[1])
	}
}

func TestEmbedCountMismatch(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, http.StatusOK, `{
		  "model": "text-embedding-3-small",
		  "data": [{"object":"embedding","index":0,"embedding":[0.1]}]
		}`)
	})

	_, err := client.Embed(context.Background(), EmbedRequest{Inputs: []string{"a", "b"}})
	if !errors.Is(err, ErrDecodeResponse) {
		t.Fatalf("error = %v, want ErrDecodeResponse on a short embedding batch", err)
	}
}

func TestEmbedEmptyInputs(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("upstream must not be called for an empty input list")
	})

	_, err := client.Embed(context.Background(), EmbedRequest{})
	if !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("error = %v, want ErrInvalidConfig", err)
	}
}

// ---------------------------------------------------------------------------
// Streaming
// ---------------------------------------------------------------------------

const sseBody = `: keep-alive

data: {"model":"gpt-4.1","choices":[{"index":0,"delta":{"role":"assistant","content":""}}]}

data: {"model":"gpt-4.1","choices":[{"index":0,"delta":{"content":"He"}}]}

data: {"model":"gpt-4.1","choices":[{"index":0,"delta":{"content":"llo"}}]}

data: {"model":"gpt-4.1","choices":[{"index":0,"delta":{},"finish_reason":"stop"}]}

data: {"model":"gpt-4.1","choices":[],"usage":{"prompt_tokens":7,"completion_tokens":2,"total_tokens":9}}

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
	// Empty deltas (role-only and finish chunks) must not be forwarded.
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
			`{"error":{"message":"Incorrect API key","type":"invalid_authentication_error"}}`)
	}))
	t.Cleanup(srv.Close)

	base := Config{
		APIKey:      "bad-key",
		BaseURL:     srv.URL,
		Model:       "gpt-4.1",
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

// unsupportedTemperature is the exact envelope a reasoning model returns
// when the request carries an explicit temperature.
const unsupportedTemperature = `{"error":{
  "message":"Unsupported value: 'temperature' does not support 0.2 with this model. Only the default (1) value is supported.",
  "type":"invalid_request_error",
  "param":"temperature",
  "code":"unsupported_value"}}`

func TestGenerateDropsSamplingWhenModelRejectsIt(t *testing.T) {
	var bodies []map[string]any

	client := newTestClientWith(t, func(w http.ResponseWriter, r *http.Request) {
		var raw map[string]any
		if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
			t.Errorf("decode request: %v", err)
		}
		bodies = append(bodies, raw)

		if _, sent := raw["temperature"]; sent {
			writeJSON(t, w, http.StatusBadRequest, unsupportedTemperature)
			return
		}
		writeJSON(t, w, http.StatusOK, okCompletion)
	}, func(c *Config) { c.Model = "gpt-5-mini" })

	resp, err := client.Generate(context.Background(), ChatRequest{
		Messages:    []Message{{Role: RoleUser, Content: "hi"}},
		Temperature: 0.2,
	})
	if err != nil {
		t.Fatalf("Generate() error = %v, want the sampling override to be dropped and retried", err)
	}
	if resp.Content == "" {
		t.Error("Content is empty after the retry succeeded")
	}

	if len(bodies) != 2 {
		t.Fatalf("upstream calls = %d, want 2 (rejected, then retried without sampling)", len(bodies))
	}
	if got, sent := bodies[0]["temperature"]; !sent || got != 0.2 {
		t.Errorf("first attempt temperature = %v (sent=%v), want 0.2", got, sent)
	}
	if _, sent := bodies[1]["temperature"]; sent {
		t.Error("retry still carried temperature")
	}
	if _, sent := bodies[1]["top_p"]; sent {
		t.Error("retry still carried top_p; both sampling fields must go together")
	}

	// The refusal is memoised: a later call must build the correct body
	// first time rather than paying the round trip again.
	if _, err := client.Generate(context.Background(), ChatRequest{
		Messages:    []Message{{Role: RoleUser, Content: "again"}},
		Temperature: 0.2,
	}); err != nil {
		t.Fatalf("second Generate() error = %v", err)
	}
	if len(bodies) != 3 {
		t.Errorf("upstream calls = %d, want 3 — the second call must not re-probe", len(bodies))
	}
}

func TestGenerateStreamDropsSamplingWhenModelRejectsIt(t *testing.T) {
	var calls int32

	client := newTestClientWith(t, func(w http.ResponseWriter, r *http.Request) {
		var raw map[string]any
		if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
			t.Errorf("decode request: %v", err)
		}
		atomic.AddInt32(&calls, 1)
		if _, sent := raw["temperature"]; sent {
			writeJSON(t, w, http.StatusBadRequest, unsupportedTemperature)
			return
		}
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("data: {\"choices\":[{\"delta\":{\"content\":\"ok\"}}]}\n\ndata: [DONE]\n\n")); err != nil {
			t.Errorf("write stream: %v", err)
		}
	}, func(c *Config) { c.Model = "gpt-5-mini" })

	var got string
	resp, err := client.GenerateStream(context.Background(), ChatRequest{
		Messages:    []Message{{Role: RoleUser, Content: "hi"}},
		Temperature: 0.2,
	}, func(chunk StreamChunk) error {
		got += chunk.Delta
		return nil
	})
	if err != nil {
		t.Fatalf("GenerateStream() error = %v", err)
	}
	if got != "ok" || resp.Content != "ok" {
		t.Errorf("streamed content = %q / %q, want \"ok\"", got, resp.Content)
	}
	if n := atomic.LoadInt32(&calls); n != 2 {
		t.Errorf("upstream calls = %d, want 2", n)
	}
}

// A 400 that merely disagrees with the VALUE must still reach the caller.
// Silently retrying it without sampling would turn an operator's bad
// BOT_OPENAI_TEMPERATURE into a working call with different behaviour.
func TestGenerateDoesNotSwallowOtherTemperatureErrors(t *testing.T) {
	var calls int32

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		writeJSON(t, w, http.StatusBadRequest, `{"error":{
  "message":"0.2 is less than the minimum of 1 - 'temperature'",
  "type":"invalid_request_error",
  "param":"temperature",
  "code":"invalid_value"}}`)
	})

	if _, err := client.Generate(context.Background(), ChatRequest{
		Messages:    []Message{{Role: RoleUser, Content: "hi"}},
		Temperature: 0.2,
	}); err == nil {
		t.Fatal("Generate() error = nil, want the 400 surfaced")
	}
	if n := atomic.LoadInt32(&calls); n != 1 {
		t.Errorf("upstream calls = %d, want 1 — an unrecognised 400 must not be retried", n)
	}
}

// The boot probe must not cap completion tokens: a reasoning model spends
// that budget on hidden reasoning and returns finish_reason=length, which
// would fail boot on a perfectly good credential.
func TestProbeDoesNotCapCompletionTokens(t *testing.T) {
	var raw map[string]any

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
			t.Errorf("decode request: %v", err)
		}
		writeJSON(t, w, http.StatusOK, okCompletion)
	}))
	t.Cleanup(srv.Close)

	if _, err := NewClient(context.Background(), Config{
		APIKey:        "test-key",
		BaseURL:       srv.URL,
		Model:         "gpt-5-mini",
		Temperature:   -1,
		TopP:          -1,
		Timeout:       5 * time.Second,
		MaxRetries:    0,
		RetryDelay:    time.Millisecond,
		RequireAtBoot: true,
	}); err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	if _, capped := raw["max_completion_tokens"]; capped {
		t.Error("probe sent max_completion_tokens; a reasoning model would return finish_reason=length and fail boot")
	}
}

func TestGenerateSendsReasoningEffort(t *testing.T) {
	var raw map[string]any

	client := newTestClientWith(t, func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
			t.Errorf("decode request: %v", err)
		}
		writeJSON(t, w, http.StatusOK, okCompletion)
	}, func(c *Config) {
		c.Model = "gpt-5.6-luna"
		c.ReasoningEffort = ReasoningEffortNone
	})

	if _, err := client.Generate(context.Background(), ChatRequest{
		Messages: []Message{{Role: RoleUser, Content: "hi"}},
	}); err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if got := raw["reasoning_effort"]; got != "none" {
		t.Errorf("reasoning_effort = %v, want \"none\"", got)
	}
	// Chat Completions takes the FLAT field. The nested Responses-API
	// shape {"reasoning":{"effort":...}} would be ignored here, and the
	// model would silently keep reasoning at full latency.
	if _, nested := raw["reasoning"]; nested {
		t.Error("sent the nested Responses-API `reasoning` object; chat-completions needs flat reasoning_effort")
	}
}

func TestReasoningEffortIsOmittedWhenUnset(t *testing.T) {
	var raw map[string]any

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
			t.Errorf("decode request: %v", err)
		}
		writeJSON(t, w, http.StatusOK, okCompletion)
	})

	if _, err := client.Generate(context.Background(), ChatRequest{
		Messages: []Message{{Role: RoleUser, Content: "hi"}},
	}); err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if _, sent := raw["reasoning_effort"]; sent {
		t.Error("reasoning_effort was sent while unset; a non-reasoning model may 400 on it")
	}
}

// reasoning_effort must survive the sampling-drop retry — otherwise a
// gpt-5.6 model configured for none would quietly go back to full
// reasoning latency on every call that carried a temperature.
func TestReasoningEffortSurvivesSamplingRetry(t *testing.T) {
	var bodies []map[string]any

	client := newTestClientWith(t, func(w http.ResponseWriter, r *http.Request) {
		var raw map[string]any
		if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
			t.Errorf("decode request: %v", err)
		}
		bodies = append(bodies, raw)
		if _, sent := raw["temperature"]; sent {
			writeJSON(t, w, http.StatusBadRequest, unsupportedTemperature)
			return
		}
		writeJSON(t, w, http.StatusOK, okCompletion)
	}, func(c *Config) {
		c.Model = "gpt-5.6-luna"
		c.ReasoningEffort = ReasoningEffortNone
	})

	if _, err := client.Generate(context.Background(), ChatRequest{
		Messages:    []Message{{Role: RoleUser, Content: "hi"}},
		Temperature: 0.2,
	}); err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if len(bodies) != 2 {
		t.Fatalf("upstream calls = %d, want 2", len(bodies))
	}
	if got := bodies[1]["reasoning_effort"]; got != "none" {
		t.Errorf("retry reasoning_effort = %v, want \"none\"", got)
	}
}

func TestConfigRejectsUnknownReasoningEffort(t *testing.T) {
	base := Config{APIKey: "k", Model: "gpt-5.6-luna", Temperature: -1, TopP: -1}

	for _, effort := range []string{"", ReasoningEffortNone, ReasoningEffortMinimal,
		ReasoningEffortLow, ReasoningEffortMedium, ReasoningEffortHigh,
		ReasoningEffortXHigh, ReasoningEffortMax} {
		cfg := base
		cfg.ReasoningEffort = effort
		if err := cfg.Validate(); err != nil {
			t.Errorf("Validate(ReasoningEffort=%q) error = %v, want nil", effort, err)
		}
	}

	for _, effort := range []string{"None", "off", "disabled", "0", "lowest"} {
		cfg := base
		cfg.ReasoningEffort = effort
		if err := cfg.Validate(); !errors.Is(err, ErrInvalidConfig) {
			t.Errorf("Validate(ReasoningEffort=%q) error = %v, want ErrInvalidConfig", effort, err)
		}
	}
}
