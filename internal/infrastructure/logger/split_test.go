package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
)

// TestSplitConsoleTextFileJSON verifies one AppLogger fans a single event out
// to a text console AND a JSON file simultaneously.
func TestSplitConsoleTextFileJSON(t *testing.T) {
	var console, file bytes.Buffer
	lg := New(context.Background(), nil, Options{
		ConsoleWriter: &console, ConsoleFormat: FormatText,
		FileWriter: &file, FileFormat: FormatJSON,
		WithColor: true,
	})
	lg.Info("split.test", "k", "v")

	// Console side: a text prefixed line, NOT JSON.
	cs := strings.TrimSpace(console.String())
	if !strings.Contains(cs, "split.test") {
		t.Errorf("console missing message, got: %q", cs)
	}
	if strings.HasPrefix(cs, "{") {
		t.Errorf("console should be text, looks like JSON: %q", cs)
	}

	// File side: valid JSON carrying the structured fields.
	var got map[string]any
	if err := json.Unmarshal(bytes.TrimSpace(file.Bytes()), &got); err != nil {
		t.Fatalf("file line is not JSON: %v; line=%s", err, file.String())
	}
	if got["msg"] != "split.test" || got["k"] != "v" {
		t.Errorf("file JSON missing fields: %v", got)
	}

	// Inline coloring must be OFF whenever a JSON destination exists, so the
	// SQL trace helper won't embed ANSI into the JSON msg field.
	if lg.ColorEnabled() {
		t.Errorf("ColorEnabled() must be false when a JSON destination is present")
	}
}
