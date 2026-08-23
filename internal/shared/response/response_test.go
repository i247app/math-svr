package response

import (
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"

	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
)

// A MathError carries a localized status.StatusMessage (a named string type).
// WriteJson must emit that message, not the raw base-error text.
func TestWriteJsonKeepsMathErrorMessage(t *testing.T) {
	mathErr := errs.NewError(context.Background(), status.USER_NOT_FOUND, nil,
		errors.New("user repo find: sql: no rows"))

	w := httptest.NewRecorder()
	WriteJson(w, nil, mathErr)

	var got map[string]any
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	want := string(mathErr.GetStatusMessage())
	if got["mmessage"] != want {
		t.Errorf("mmessage = %q, want %q", got["mmessage"], want)
	}
	if got["debug"] != mathErr.Error() {
		t.Errorf("debug = %q, want %q", got["debug"], mathErr.Error())
	}
	if got["mstatus"] != float64(mathErr.GetStatusCode()) {
		t.Errorf("mstatus = %v, want %v", got["mstatus"], mathErr.GetStatusCode())
	}
}
