package openai

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"sync/atomic"
	"testing"
)

// okResponse is a minimal completed Responses-API body with one message
// item. The id is what every chained test asserts on.
const okResponse = `{
  "id": "resp_A",
  "model": "gpt-5.6-luna",
  "status": "completed",
  "output": [
    {"type": "message", "role": "assistant",
     "content": [{"type": "output_text", "text": "{\"title\":\"Lớp 1\"}"}]}
  ],
  "usage": {"input_tokens": 120, "output_tokens": 30, "total_tokens": 150}
}`

func newResponsesClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	return newTestClientWith(t, handler, func(c *Config) {
		c.Model = "gpt-5.6-luna"
		c.ReasoningEffort = ReasoningEffortNone
	})
}

func TestRespondFirstTurnRequestShape(t *testing.T) {
	var raw map[string]any
	var path string

	client := newResponsesClient(t, func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
		if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
			t.Errorf("decode request: %v", err)
		}
		writeJSON(t, w, http.StatusOK, okResponse)
	})

	resp, err := client.Respond(context.Background(), RespondRequest{
		Instructions: "You are a math tutor.",
		Input:        []Message{{Role: RoleUser, Content: "make an exam"}},
		Store:        true,
		MaxTokens:    4096,
		JSONMode:     true,
		Temperature:  -1,
		TopP:         -1,
	})
	if err != nil {
		t.Fatalf("Respond() error = %v", err)
	}

	if path != "/responses" {
		t.Errorf("path = %q, want /responses", path)
	}
	if resp.ID != "resp_A" {
		t.Errorf("ID = %q, want resp_A", resp.ID)
	}
	if resp.Content != `{"title":"Lớp 1"}` {
		t.Errorf("Content = %q", resp.Content)
	}
	if resp.Usage.PromptTokens != 120 || resp.Usage.CompletionTokens != 30 {
		t.Errorf("Usage = %+v, want input/output folded into prompt/completion", resp.Usage)
	}

	if raw["instructions"] != "You are a math tutor." {
		t.Errorf("instructions = %v", raw["instructions"])
	}
	if _, sent := raw["previous_response_id"]; sent {
		t.Error("first turn must not carry previous_response_id")
	}
	if raw["store"] != true {
		t.Errorf("store = %v, want explicit true", raw["store"])
	}
	// Field-name traps: the chat-completions spellings would be silently
	// ignored on this endpoint.
	if got, ok := raw["max_output_tokens"].(float64); !ok || got != 4096 {
		t.Errorf("max_output_tokens = %v, want 4096", raw["max_output_tokens"])
	}
	for _, chatOnly := range []string{"max_completion_tokens", "max_tokens", "reasoning_effort", "response_format", "messages"} {
		if _, sent := raw[chatOnly]; sent {
			t.Errorf("sent chat-completions field %q on /responses", chatOnly)
		}
	}
	reasoning, _ := raw["reasoning"].(map[string]any)
	if reasoning == nil || reasoning["effort"] != "none" {
		t.Errorf("reasoning = %v, want nested {effort: none}", raw["reasoning"])
	}
	text, _ := raw["text"].(map[string]any)
	format, _ := text["format"].(map[string]any)
	if format == nil || format["type"] != "json_object" {
		t.Errorf("text.format = %v, want json_object", raw["text"])
	}
	input, _ := raw["input"].([]any)
	if len(input) != 1 {
		t.Fatalf("input = %v, want one message", raw["input"])
	}
	first, _ := input[0].(map[string]any)
	if first["role"] != "user" || first["content"] != "make an exam" {
		t.Errorf("input[0] = %v", first)
	}
}

func TestRespondChainedTurnResendsInstructions(t *testing.T) {
	var raw map[string]any

	client := newResponsesClient(t, func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
			t.Errorf("decode request: %v", err)
		}
		writeJSON(t, w, http.StatusOK, `{"id":"resp_B","status":"completed",
		  "output":[{"type":"message","content":[{"type":"output_text","text":"ok"}]}]}`)
	})

	resp, err := client.Respond(context.Background(), RespondRequest{
		Instructions:       "You are a math tutor.",
		Input:              []Message{{Role: RoleUser, Content: "next exam"}},
		PreviousResponseID: "resp_A",
		Store:              true,
		Temperature:        -1,
		TopP:               -1,
	})
	if err != nil {
		t.Fatalf("Respond() error = %v", err)
	}
	if resp.ID != "resp_B" {
		t.Errorf("ID = %q, want resp_B", resp.ID)
	}
	if raw["previous_response_id"] != "resp_A" {
		t.Errorf("previous_response_id = %v, want resp_A", raw["previous_response_id"])
	}
	// The vendor does not inherit instructions across the chain. Dropping
	// them on turn 2 would silently turn the tutor back into a generic
	// assistant.
	if raw["instructions"] != "You are a math tutor." {
		t.Errorf("chained turn instructions = %v, want them resent", raw["instructions"])
	}
	if raw["store"] != true {
		t.Error("chained turn must keep store: true so the NEXT turn can chain too")
	}
}

func TestRespondRejectsChainWithoutStore(t *testing.T) {
	var calls int32
	client := newResponsesClient(t, func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		writeJSON(t, w, http.StatusOK, okResponse)
	})

	_, err := client.Respond(context.Background(), RespondRequest{
		Input:              []Message{{Role: RoleUser, Content: "x"}},
		PreviousResponseID: "resp_A",
		Store:              false,
	})
	if !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("error = %v, want ErrInvalidConfig", err)
	}
	if n := atomic.LoadInt32(&calls); n != 0 {
		t.Errorf("upstream calls = %d, want 0 — rejected before the wire", n)
	}
}

func TestRespondEmptyInput(t *testing.T) {
	client := newResponsesClient(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("upstream must not be called")
	})
	if _, err := client.Respond(context.Background(), RespondRequest{Store: true}); !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("error = %v, want ErrInvalidConfig", err)
	}
}

// The chain pointer can die two ways on the wire; both must surface as
// the one sentinel the caller branches on, and neither may be retried.
func TestRespondPreviousResponseGone(t *testing.T) {
	tests := []struct {
		name   string
		status int
		body   string
	}{
		{"404 object gone", http.StatusNotFound,
			`{"error":{"message":"Response with id 'resp_A' not found.","type":"invalid_request_error","code":"not_found"}}`},
		{"400 param rejected", http.StatusBadRequest,
			`{"error":{"message":"Invalid 'previous_response_id'","type":"invalid_request_error","param":"previous_response_id","code":"invalid_value"}}`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var calls int32
			client := newResponsesClient(t, func(w http.ResponseWriter, r *http.Request) {
				atomic.AddInt32(&calls, 1)
				writeJSON(t, w, tc.status, tc.body)
			})

			_, err := client.Respond(context.Background(), RespondRequest{
				Input:              []Message{{Role: RoleUser, Content: "x"}},
				PreviousResponseID: "resp_A",
				Store:              true,
				Temperature:        -1,
				TopP:               -1,
			})
			if !IsPreviousResponseGone(err) {
				t.Fatalf("error = %v, want ErrPreviousResponseGone", err)
			}
			var api *APIError
			if !errors.As(err, &api) {
				t.Error("APIError must remain reachable through the sentinel")
			}
			if n := atomic.LoadInt32(&calls); n != 1 {
				t.Errorf("upstream calls = %d, want 1 — a dead pointer is not retried", n)
			}
		})
	}
}

// A 404 on an UNCHAINED call is not a dead pointer — there was no pointer.
// It stays a plain config-class error (unknown model, wrong base URL).
func TestRespondUnchained404IsNotGone(t *testing.T) {
	client := newResponsesClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, http.StatusNotFound,
			`{"error":{"message":"The model 'nope' does not exist","type":"invalid_request_error","code":"model_not_found"}}`)
	})
	_, err := client.Respond(context.Background(), RespondRequest{
		Input:       []Message{{Role: RoleUser, Content: "x"}},
		Store:       true,
		Temperature: -1,
		TopP:        -1,
	})
	if err == nil {
		t.Fatal("error = nil")
	}
	if IsPreviousResponseGone(err) {
		t.Error("unchained 404 was reported as a dead chain pointer")
	}
	if !IsConfigError(err) {
		t.Errorf("error = %v, want a config-class error", err)
	}
}

func TestRespondIncompleteIsNotChainable(t *testing.T) {
	client := newResponsesClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, http.StatusOK, `{"id":"resp_T","status":"incomplete",
		  "incomplete_details":{"reason":"max_output_tokens"},
		  "output":[{"type":"message","content":[{"type":"output_text","text":"{\"title\":\"Lớp"}]}],
		  "usage":{"input_tokens":10,"output_tokens":4096}}`)
	})
	resp, err := client.Respond(context.Background(), RespondRequest{
		Input: []Message{{Role: RoleUser, Content: "x"}}, Store: true, Temperature: -1, TopP: -1,
	})
	if !errors.Is(err, ErrDecodeResponse) {
		t.Fatalf("error = %v, want ErrDecodeResponse", err)
	}
	if resp == nil || resp.ID != "resp_T" || resp.IncompleteReason != "max_output_tokens" {
		t.Errorf("partial response = %+v, want id + reason attached for logging", resp)
	}
}

func TestRespondRefusal(t *testing.T) {
	client := newResponsesClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, http.StatusOK, `{"id":"resp_R","status":"completed",
		  "output":[{"type":"message","content":[{"type":"refusal","refusal":"I can't help with that."}]}]}`)
	})
	_, err := client.Respond(context.Background(), RespondRequest{
		Input: []Message{{Role: RoleUser, Content: "x"}}, Store: true, Temperature: -1, TopP: -1,
	})
	if !errors.Is(err, ErrContentRefused) {
		t.Fatalf("error = %v, want ErrContentRefused", err)
	}
}

// Reasoning items and multi-part messages: only message text is
// assembled, in order, and reasoning is never surfaced.
func TestRespondAssemblesOnlyMessageText(t *testing.T) {
	client := newResponsesClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, http.StatusOK, `{"id":"resp_M","status":"completed","output":[
		  {"type":"reasoning","summary":[{"type":"summary_text","text":"thinking..."}]},
		  {"type":"message","content":[{"type":"output_text","text":"{\"a\":"},{"type":"output_text","text":"1}"}]}
		]}`)
	})
	resp, err := client.Respond(context.Background(), RespondRequest{
		Input: []Message{{Role: RoleUser, Content: "x"}}, Store: true, Temperature: -1, TopP: -1,
	})
	if err != nil {
		t.Fatalf("Respond() error = %v", err)
	}
	if resp.Content != `{"a":1}` {
		t.Errorf("Content = %q, want message parts joined and reasoning dropped", resp.Content)
	}
}

func TestRespondNoMessageText(t *testing.T) {
	client := newResponsesClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, http.StatusOK, `{"id":"resp_E","status":"completed","output":[{"type":"reasoning"}]}`)
	})
	_, err := client.Respond(context.Background(), RespondRequest{
		Input: []Message{{Role: RoleUser, Content: "x"}}, Store: true, Temperature: -1, TopP: -1,
	})
	if !errors.Is(err, ErrDecodeResponse) {
		t.Fatalf("error = %v, want ErrDecodeResponse", err)
	}
}

// The sampling memo is per model, not per endpoint: a refusal learned on
// /responses must stop the very next /chat/completions call from paying
// the same 400, and vice versa.
func TestRespondSharesSamplingMemoWithChat(t *testing.T) {
	var calls []string

	client := newResponsesClient(t, func(w http.ResponseWriter, r *http.Request) {
		var raw map[string]any
		if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
			t.Errorf("decode request: %v", err)
		}
		calls = append(calls, r.URL.Path)
		if _, sent := raw["temperature"]; sent {
			writeJSON(t, w, http.StatusBadRequest, unsupportedTemperature)
			return
		}
		if r.URL.Path == "/responses" {
			writeJSON(t, w, http.StatusOK, okResponse)
			return
		}
		writeJSON(t, w, http.StatusOK, okCompletion)
	})

	if _, err := client.Respond(context.Background(), RespondRequest{
		Input: []Message{{Role: RoleUser, Content: "x"}}, Store: true, Temperature: 0.2, TopP: -1,
	}); err != nil {
		t.Fatalf("Respond() error = %v", err)
	}
	if _, err := client.Generate(context.Background(), ChatRequest{
		Messages: []Message{{Role: RoleUser, Content: "x"}}, Temperature: 0.2, TopP: -1,
	}); err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	want := []string{"/responses", "/responses", "/chat/completions"}
	if len(calls) != len(want) {
		t.Fatalf("calls = %v, want %v (one probe, then memo applies across endpoints)", calls, want)
	}
	for i := range want {
		if calls[i] != want[i] {
			t.Errorf("calls[%d] = %q, want %q", i, calls[i], want[i])
		}
	}
}

func TestRespondMetadataOnlyWhenStored(t *testing.T) {
	var raw map[string]any
	client := newTestClientWith(t, func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
			t.Errorf("decode request: %v", err)
		}
		writeJSON(t, w, http.StatusOK, okResponse)
	}, func(c *Config) {
		c.Model = "gpt-5.6-luna"
		c.Metadata = map[string]string{"feature": "exam_generate"}
	})

	for _, store := range []bool{false, true} {
		raw = nil
		if _, err := client.Respond(context.Background(), RespondRequest{
			Input: []Message{{Role: RoleUser, Content: "x"}}, Store: store, Temperature: -1, TopP: -1,
		}); err != nil {
			t.Fatalf("Respond(store=%v) error = %v", store, err)
		}
		_, sent := raw["metadata"]
		if sent != store {
			t.Errorf("store=%v: metadata sent=%v, want %v", store, sent, store)
		}
	}
}
