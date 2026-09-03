package gemini

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
// server: it exercises the actual URL construction, request translation,
// header set, status handling and retry loop rather than a hand-rolled
// transport seam.

func newTestClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()

	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	client, err := NewClient(context.Background(), Config{
		APIKey:      "test-key",
		BaseURL:     srv.URL,
		Model:       "gemini-2.0-flash",
		EmbedModel:  "text-embedding-004",
		Temperature: -1,
		TopP:        -1,
		Timeout:     5 * time.Second,
		MaxRetries:  2,
		RetryDelay:  time.Millisecond,
	})
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

const okGeneration = `{
  "candidates": [{
    "content": {"role": "model", "parts": [{"text": "4"}]},
    "finishReason": "STOP",
    "index": 0
  }],
  "usageMetadata": {"promptTokenCount": 11, "candidatesTokenCount": 1, "totalTokenCount": 12},
  "modelVersion": "gemini-2.0-flash"
}`

func TestGenerateSuccess(t *testing.T) {
	var gotPath, gotKey, gotQuery string
	var gotBody wireGenerateRequest

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		gotKey = r.Header.Get(apiKeyHeader)
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Errorf("decode request: %v", err)
		}
		writeJSON(t, w, http.StatusOK, okGeneration)
	})

	res, err := client.Generate(context.Background(), ChatRequest{
		Messages:    []Message{{Role: RoleUser, Content: "2+2?"}},
		Temperature: -1,
		TopP:        -1,
	})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// The model is part of the PATH here, not the body — the single
	// biggest shape difference from the OpenAI-style providers.
	if gotPath != "/models/gemini-2.0-flash:generateContent" {
		t.Errorf("path = %q, want /models/gemini-2.0-flash:generateContent", gotPath)
	}
	if gotKey != "test-key" {
		t.Errorf("%s = %q, want test-key", apiKeyHeader, gotKey)
	}
	// The credential must never ride in the URL, where it would land in
	// every access log and proxy trace.
	if strings.Contains(gotQuery, "key") {
		t.Errorf("query = %q, want no credential in the URL", gotQuery)
	}

	if len(gotBody.Contents) != 1 || gotBody.Contents[0].Role != "user" {
		t.Fatalf("contents = %+v, want one user turn", gotBody.Contents)
	}
	if gotBody.Contents[0].Parts[0].Text != "2+2?" {
		t.Errorf("part text = %q", gotBody.Contents[0].Parts[0].Text)
	}
	// Nothing to configure → generationConfig must be omitted entirely so
	// the model's own defaults stand.
	if gotBody.GenerationConfig != nil {
		t.Errorf("generationConfig = %+v, want omitted", gotBody.GenerationConfig)
	}

	if res.Content != "4" {
		t.Errorf("Content = %q, want 4", res.Content)
	}
	if res.FinishReason != finishReasonStop {
		t.Errorf("FinishReason = %q, want STOP", res.FinishReason)
	}
	if res.Usage != (Usage{PromptTokens: 11, CompletionTokens: 1, TotalTokens: 12}) {
		t.Errorf("Usage = %+v", res.Usage)
	}
}

// TestSystemInstructionAndRoleMapping is the translation that has no
// equivalent in the OpenAI-shaped providers: Gemini has no system role, so
// system turns must be lifted into a separate top-level object, assistant
// becomes "model", and consecutive same-role turns must be merged because
// the API expects them to alternate.
func TestSystemInstructionAndRoleMapping(t *testing.T) {
	var gotBody wireGenerateRequest

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Errorf("decode request: %v", err)
		}
		writeJSON(t, w, http.StatusOK, okGeneration)
	})

	_, err := client.Generate(context.Background(), ChatRequest{
		Messages: []Message{
			{Role: RoleSystem, Content: "be terse"},
			{Role: RoleSystem, Content: "answer in Vietnamese"},
			{Role: RoleUser, Content: "first"},
			{Role: RoleUser, Content: "second"},
			{Role: RoleAssistant, Content: "ok"},
			{Role: RoleTool, Content: "tool output"},
		},
	})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Both system turns joined into systemInstruction, none left in contents.
	if gotBody.SystemInstruction == nil {
		t.Fatal("systemInstruction = nil, want the system turns lifted out")
	}
	sys := gotBody.SystemInstruction.Parts[0].Text
	if !strings.Contains(sys, "be terse") || !strings.Contains(sys, "answer in Vietnamese") {
		t.Errorf("systemInstruction = %q, want both system turns", sys)
	}
	if gotBody.SystemInstruction.Role != "" {
		t.Errorf("systemInstruction.role = %q, want empty (it has no role)", gotBody.SystemInstruction.Role)
	}

	// user+user merged; assistant → model; tool → user.
	if len(gotBody.Contents) != 3 {
		t.Fatalf("contents = %d turns (%+v), want 3 after merging", len(gotBody.Contents), gotBody.Contents)
	}
	if gotBody.Contents[0].Role != "user" || len(gotBody.Contents[0].Parts) != 2 {
		t.Errorf("contents[0] = %+v, want one user turn with two parts", gotBody.Contents[0])
	}
	if gotBody.Contents[1].Role != "model" {
		t.Errorf("contents[1].role = %q, want model (assistant maps to model)", gotBody.Contents[1].Role)
	}
	if gotBody.Contents[2].Role != "user" {
		t.Errorf("contents[2].role = %q, want user (tool maps to user)", gotBody.Contents[2].Role)
	}
}

func TestGenerateSendsGenerationConfig(t *testing.T) {
	var gotBody wireGenerateRequest

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Errorf("decode request: %v", err)
		}
		writeJSON(t, w, http.StatusOK, okGeneration)
	})

	_, err := client.Generate(context.Background(), ChatRequest{
		Messages:    []Message{{Role: RoleUser, Content: "hi"}},
		Temperature: 0.2,
		TopP:        0.9,
		MaxTokens:   256,
		Stop:        []string{"###"},
		JSONMode:    true,
	})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	cfg := gotBody.GenerationConfig
	if cfg == nil {
		t.Fatal("generationConfig = nil")
	}
	if cfg.Temperature == nil || *cfg.Temperature != 0.2 {
		t.Errorf("temperature = %v, want 0.2", cfg.Temperature)
	}
	if cfg.TopP == nil || *cfg.TopP != 0.9 {
		t.Errorf("topP = %v, want 0.9", cfg.TopP)
	}
	if cfg.MaxOutputTokens == nil || *cfg.MaxOutputTokens != 256 {
		t.Errorf("maxOutputTokens = %v, want 256", cfg.MaxOutputTokens)
	}
	if len(cfg.StopSequences) != 1 || cfg.StopSequences[0] != "###" {
		t.Errorf("stopSequences = %v", cfg.StopSequences)
	}
	if cfg.ResponseMimeType != "application/json" {
		t.Errorf("responseMimeType = %q, want application/json", cfg.ResponseMimeType)
	}
}

// TestGenerateJoinsMultipleParts: Gemini may split one reply across
// several parts; taking parts[0] alone would silently truncate the answer.
func TestGenerateJoinsMultipleParts(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, http.StatusOK, `{
		  "candidates": [{
		    "content": {"role":"model","parts":[{"text":"{\"a\":"},{"text":"1}"}]},
		    "finishReason": "STOP"
		  }]
		}`)
	})

	res, err := client.Generate(context.Background(), ChatRequest{
		Messages: []Message{{Role: RoleUser, Content: "hi"}},
	})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if res.Content != `{"a":1}` {
		t.Errorf("Content = %q, want every part joined", res.Content)
	}
}

// TestGenerateBlockedPrompt: a blocked prompt returns HTTP 200 with ZERO
// candidates and the reason only in promptFeedback. Without this branch it
// would surface as a baffling "no candidates".
func TestGenerateBlockedPrompt(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, http.StatusOK, `{"promptFeedback": {"blockReason": "SAFETY"}}`)
	})

	_, err := client.Generate(context.Background(), ChatRequest{
		Messages: []Message{{Role: RoleUser, Content: "hi"}},
	})
	if !errors.Is(err, ErrContentBlocked) {
		t.Fatalf("error = %v, want ErrContentBlocked", err)
	}
	if !strings.Contains(err.Error(), "SAFETY") {
		t.Errorf("error = %v, want it to name the block reason", err)
	}
}

// TestGenerateBlockedResponse: a blocked ANSWER also returns 200, with the
// reason on the candidate instead.
func TestGenerateBlockedResponse(t *testing.T) {
	for _, reason := range []string{finishReasonSafety, finishReasonRecitation, finishReasonProhibited} {
		t.Run(reason, func(t *testing.T) {
			client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
				writeJSON(t, w, http.StatusOK, `{"candidates":[{"content":{"role":"model","parts":[]},"finishReason":"`+reason+`"}]}`)
			})

			_, err := client.Generate(context.Background(), ChatRequest{
				Messages: []Message{{Role: RoleUser, Content: "hi"}},
			})
			if !errors.Is(err, ErrContentBlocked) {
				t.Fatalf("error = %v, want ErrContentBlocked", err)
			}
		})
	}
}

// TestGenerateTruncated: MAX_TOKENS means the body is cut off mid-JSON.
func TestGenerateTruncated(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, http.StatusOK, `{
		  "candidates": [{"content":{"role":"model","parts":[{"text":"{\"questions\": ["}]},"finishReason":"MAX_TOKENS"}],
		  "usageMetadata": {"promptTokenCount": 11, "candidatesTokenCount": 256, "totalTokenCount": 267}
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
			body:       `{"error":{"code":401,"status":"UNAUTHENTICATED","message":"API key not valid"}}`,
			wantAuth:   true,
			wantConfig: true,
			wantCalls:  1,
		},
		{
			name:       "403 no permission",
			status:     http.StatusForbidden,
			body:       `{"error":{"code":403,"status":"PERMISSION_DENIED","message":"Permission denied"}}`,
			wantAuth:   true,
			wantConfig: true,
			wantCalls:  1,
		},
		{
			name:       "404 unknown model",
			status:     http.StatusNotFound,
			body:       `{"error":{"code":404,"status":"NOT_FOUND","message":"models/nope is not found"}}`,
			wantConfig: true,
			wantCalls:  1,
		},
		{
			name:       "400 FAILED_PRECONDITION (billing)",
			status:     http.StatusBadRequest,
			body:       `{"error":{"code":400,"status":"FAILED_PRECONDITION","message":"enable billing"}}`,
			wantConfig: true,
			wantCalls:  1,
		},
		{
			name:      "400 context too large",
			status:    http.StatusBadRequest,
			body:      `{"error":{"code":400,"status":"INVALID_ARGUMENT","message":"The input token count exceeds the maximum"}}`,
			wantIs:    ErrContextTooLarge,
			wantCalls: 1,
		},
		{
			name:      "429 burst without hint is surfaced",
			status:    http.StatusTooManyRequests,
			body:      `{"error":{"code":429,"status":"RESOURCE_EXHAUSTED","message":"Resource has been exhausted"}}`,
			wantIs:    ErrRateLimited,
			wantCalls: 1,
		},
		{
			name:       "429 daily quota is not retried",
			status:     http.StatusTooManyRequests,
			body:       `{"error":{"code":429,"status":"RESOURCE_EXHAUSTED","message":"You exceeded your current quota exceeded for the free tier"}}`,
			wantIs:     ErrQuotaExhausted,
			wantConfig: true,
			wantCalls:  1,
		},
		{
			name:      "500 exhausts the retry budget",
			status:    http.StatusInternalServerError,
			body:      `{"error":{"code":500,"status":"INTERNAL","message":"internal error"}}`,
			wantCalls: 3,
		},
		{
			name:      "503 exhausts the retry budget",
			status:    http.StatusServiceUnavailable,
			body:      `{"error":{"code":503,"status":"UNAVAILABLE","message":"model overloaded"}}`,
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

// TestGenerateRetriesWithRetryInfo: the backoff hint lives in
// error.details, not in a Retry-After header.
func TestGenerateRetriesWithRetryInfo(t *testing.T) {
	var calls int32
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&calls, 1) == 1 {
			writeJSON(t, w, http.StatusTooManyRequests, `{"error":{
			  "code": 429,
			  "status": "RESOURCE_EXHAUSTED",
			  "message": "Resource has been exhausted",
			  "details": [{"@type":"type.googleapis.com/google.rpc.RetryInfo","retryDelay":"1s"}]
			}}`)
			return
		}
		writeJSON(t, w, http.StatusOK, okGeneration)
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
	const secret = "<html>gateway: api_key=AIza-should-never-surface</html>"

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
	if strings.Contains(err.Error(), "AIza-should-never-surface") || strings.Contains(err.Error(), "<html>") {
		t.Fatalf("raw upstream body leaked into the error: %v", err)
	}
	if !strings.Contains(err.Error(), "not a Gemini error envelope") {
		t.Errorf("error = %v, want the size-only fallback message", err)
	}
}

// ---------------------------------------------------------------------------
// Embeddings
// ---------------------------------------------------------------------------

func TestEmbed(t *testing.T) {
	var gotPath string
	var gotBody wireBatchEmbedRequest

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Errorf("decode request: %v", err)
		}
		writeJSON(t, w, http.StatusOK, `{
		  "embeddings": [{"values": [0.1, 0.2]}, {"values": [0.3, 0.4]}]
		}`)
	})

	res, err := client.Embed(context.Background(), EmbedRequest{
		Inputs: []string{"first", "second"},
	})
	if err != nil {
		t.Fatalf("Embed() error = %v", err)
	}

	if gotPath != "/models/text-embedding-004:batchEmbedContents" {
		t.Errorf("path = %q, want the embed model's batch endpoint", gotPath)
	}
	if len(gotBody.Requests) != 2 {
		t.Fatalf("requests = %d, want 2", len(gotBody.Requests))
	}
	// Each sub-request repeats the model in models/<id> form.
	if gotBody.Requests[0].Model != "models/text-embedding-004" {
		t.Errorf("sub-request model = %q", gotBody.Requests[0].Model)
	}
	if gotBody.Requests[0].Content.Parts[0].Text != "first" {
		t.Errorf("sub-request text = %q", gotBody.Requests[0].Content.Parts[0].Text)
	}

	if len(res.Vectors) != 2 || res.Vectors[0][0] != 0.1 || res.Vectors[1][0] != 0.3 {
		t.Errorf("Vectors = %v", res.Vectors)
	}
}

func TestEmbedCountMismatch(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, http.StatusOK, `{"embeddings":[{"values":[0.1]}]}`)
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

	if _, err := client.Embed(context.Background(), EmbedRequest{}); !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("error = %v, want ErrInvalidConfig", err)
	}
}

// ---------------------------------------------------------------------------
// Streaming
// ---------------------------------------------------------------------------

// Gemini's SSE stream carries no [DONE] terminator — it simply ends.
const sseBody = `: keep-alive

data: {"candidates":[{"content":{"role":"model","parts":[{"text":"He"}]}}],"modelVersion":"gemini-2.0-flash"}

data: {"candidates":[{"content":{"role":"model","parts":[{"text":"llo"}]}}]}

data: {"candidates":[{"content":{"role":"model","parts":[{"text":""}]},"finishReason":"STOP"}],"usageMetadata":{"promptTokenCount":7,"candidatesTokenCount":2,"totalTokenCount":9}}

`

func TestGenerateStream(t *testing.T) {
	var gotPath, gotQuery string

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
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

	if gotPath != "/models/gemini-2.0-flash:streamGenerateContent" {
		t.Errorf("path = %q, want the streaming endpoint", gotPath)
	}
	// Without alt=sse the endpoint streams a JSON array instead of SSE
	// frames, which parseSSE cannot read.
	if gotQuery != "alt=sse" {
		t.Errorf("query = %q, want alt=sse", gotQuery)
	}

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
	if res.FinishReason != finishReasonStop {
		t.Errorf("FinishReason = %q, want STOP", res.FinishReason)
	}
}

func TestGenerateStreamBlockedMidStream(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		body := `data: {"candidates":[{"content":{"role":"model","parts":[{"text":"par"}]}}]}

data: {"candidates":[{"content":{"role":"model","parts":[]},"finishReason":"SAFETY"}]}

`
		if _, err := w.Write([]byte(body)); err != nil {
			t.Errorf("write response: %v", err)
		}
	})

	_, err := client.GenerateStream(context.Background(), ChatRequest{
		Messages: []Message{{Role: RoleUser, Content: "hi"}},
	}, func(StreamChunk) error { return nil })

	if !errors.Is(err, ErrContentBlocked) {
		t.Fatalf("error = %v, want ErrContentBlocked", err)
	}
}

func TestParseSSE(t *testing.T) {
	t.Run("skips comments, ends at EOF without DONE", func(t *testing.T) {
		events, err := parseSSE([]byte(sseBody))
		if err != nil {
			t.Fatalf("parseSSE() error = %v", err)
		}
		if len(events) != 3 {
			t.Fatalf("events = %d, want 3", len(events))
		}
	})

	t.Run("empty stream", func(t *testing.T) {
		if _, err := parseSSE([]byte(": keep-alive\n")); !errors.Is(err, ErrDecodeResponse) {
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
			`{"error":{"code":401,"status":"UNAUTHENTICATED","message":"API key not valid"}}`)
	}))
	t.Cleanup(srv.Close)

	base := Config{
		APIKey:      "bad-key",
		BaseURL:     srv.URL,
		Model:       "gemini-2.0-flash",
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
