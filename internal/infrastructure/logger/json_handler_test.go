package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestJSONHandler_EmitsStructuredFields(t *testing.T) {
	var buf bytes.Buffer
	req := httptest.NewRequest(http.MethodPost, "/users/create", nil)

	lg := New(context.Background(), req, Options{ConsoleFormat: FormatJSON, ConsoleWriter: &buf})
	lg.Info("user.created", "uid", int64(42), "email", "a@b.com")

	var got map[string]any
	if err := json.Unmarshal(bytes.TrimSpace(buf.Bytes()), &got); err != nil {
		t.Fatalf("log line is not valid JSON: %v\nline: %s", err, buf.String())
	}

	if got["level"] != "info" {
		t.Errorf("level = %v, want info", got["level"])
	}
	if got["msg"] != "user.created" {
		t.Errorf("msg = %v, want user.created", got["msg"])
	}
	if got["method"] != "POST" {
		t.Errorf("method = %v, want POST", got["method"])
	}
	if got["path"] != "/users/create" {
		t.Errorf("path = %v, want /users/create", got["path"])
	}
	if got["email"] != "a@b.com" {
		t.Errorf("email attr = %v, want a@b.com", got["email"])
	}
	if _, ok := got["caller"]; !ok {
		t.Errorf("caller field missing: %v", got)
	}
	// trace_id must be absent while TraceContextFn is unset (Phase 2 state).
	if _, ok := got["trace_id"]; ok {
		t.Errorf("trace_id should be absent when TraceContextFn is nil")
	}
}

func TestJSONHandler_TraceContextHook(t *testing.T) {
	orig := TraceContextFn
	TraceContextFn = func(context.Context) (string, string) { return "abc123", "span9" }
	defer func() { TraceContextFn = orig }()

	var buf bytes.Buffer
	lg := New(context.Background(), nil, Options{ConsoleFormat: FormatJSON, ConsoleWriter: &buf})
	lg.Error("db.failed")

	var got map[string]any
	if err := json.Unmarshal(bytes.TrimSpace(buf.Bytes()), &got); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if got["trace_id"] != "abc123" {
		t.Errorf("trace_id = %v, want abc123", got["trace_id"])
	}
	if got["span_id"] != "span9" {
		t.Errorf("span_id = %v, want span9", got["span_id"])
	}
	if got["level"] != "error" {
		t.Errorf("level = %v, want error", got["level"])
	}
}
