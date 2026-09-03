package bot

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/libs/gemini"
)

func TestMapGeminiError(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		in       error
		wantCode status.StatusCode
	}{
		{"invalid config", fmt.Errorf("x: %w", gemini.ErrInvalidConfig), status.BOT_CONFIG_INVALID},
		{"unsupported op", fmt.Errorf("x: %w", gemini.ErrUnsupportedOp), status.BOT_UNSUPPORTED_OP},
		{"context too large", fmt.Errorf("x: %w", gemini.ErrContextTooLarge), status.BOT_CONTEXT_TOO_LARGE},
		{"rate limited sentinel", fmt.Errorf("x: %w", gemini.ErrRateLimited), status.BOT_RATE_LIMITED},
		{"burst 429", fmt.Errorf("x: %w", &gemini.APIError{HTTPStatus: 429}), status.BOT_RATE_LIMITED},
		{"bad key 401 → config invalid", fmt.Errorf("x: %w", &gemini.APIError{HTTPStatus: 401}), status.BOT_CONFIG_INVALID},
		{"permission 403 → config invalid", fmt.Errorf("x: %w", &gemini.APIError{HTTPStatus: 403}), status.BOT_CONFIG_INVALID},
		{"unknown model 404 → config invalid", fmt.Errorf("x: %w", &gemini.APIError{HTTPStatus: 404}), status.BOT_CONFIG_INVALID},
		{"daily quota → config invalid", fmt.Errorf("x: %w", gemini.ErrQuotaExhausted), status.BOT_CONFIG_INVALID},
		{"decode failed", fmt.Errorf("x: %w", gemini.ErrDecodeResponse), status.BOT_SERIALIZE_FAILED},
		{"unclassified → op failed", errors.New("boom"), status.BOT_OP_FAILED},
		{"500 → op failed", fmt.Errorf("x: %w", &gemini.APIError{HTTPStatus: 500}), status.BOT_OP_FAILED},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mErr := mapGeminiError(ctx, tt.in)
			if mErr == nil {
				t.Fatal("mapGeminiError() = nil, want MathError")
			}
			if got := mErr.GetStatusCode(); got != tt.wantCode {
				t.Errorf("status code = %d, want %d", got, tt.wantCode)
			}
			if !errors.Is(mErr, tt.in) && mErr.BaseError == nil {
				t.Error("MathError lost the underlying cause")
			}
		})
	}
}

// TestMapGeminiErrorContentBlocked is the mapping unique to this provider.
// A safety refusal must NOT land on BOT_OP_FAILED, which the adapter logs
// as transient and a caller may reasonably retry — the same prompt is
// refused every time.
func TestMapGeminiErrorContentBlocked(t *testing.T) {
	err := fmt.Errorf("%w: prompt blocked (SAFETY)", gemini.ErrContentBlocked)

	mErr := mapGeminiError(context.Background(), err)
	if mErr == nil {
		t.Fatal("mapGeminiError() = nil")
	}
	if got := mErr.GetStatusCode(); got != status.BOT_CONTENT_BLOCKED {
		t.Errorf("status code = %d, want BOT_CONTENT_BLOCKED", got)
	}
	// The three-file lockstep must actually carry a message for the code.
	if msg := mErr.GetStatusMessage(); msg == "" {
		t.Error("BOT_CONTENT_BLOCKED has no localized message")
	}
}

func TestMapGeminiErrorNil(t *testing.T) {
	if got := mapGeminiError(context.Background(), nil); got != nil {
		t.Errorf("mapGeminiError(nil) = %v, want nil", got)
	}
}
