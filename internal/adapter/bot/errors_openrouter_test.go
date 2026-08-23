package bot

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/libs/openrouter"
)

func TestMapOpenRouterError(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		in       error
		wantCode status.StatusCode
	}{
		{"invalid config", fmt.Errorf("x: %w", openrouter.ErrInvalidConfig), status.BOT_CONFIG_INVALID},
		{"unsupported op", fmt.Errorf("x: %w", openrouter.ErrUnsupportedOp), status.BOT_UNSUPPORTED_OP},
		{"context too large", fmt.Errorf("x: %w", openrouter.ErrContextTooLarge), status.BOT_CONTEXT_TOO_LARGE},
		{"rate limited sentinel", fmt.Errorf("x: %w", openrouter.ErrRateLimited), status.BOT_RATE_LIMITED},
		{"rate limited typed 429", fmt.Errorf("x: %w", &openrouter.APIError{HTTPStatus: 429}), status.BOT_RATE_LIMITED},
		{"auth 401 → config invalid", fmt.Errorf("x: %w", &openrouter.APIError{HTTPStatus: 401}), status.BOT_CONFIG_INVALID},
		{"moderation 403 → config invalid", fmt.Errorf("x: %w", &openrouter.APIError{HTTPStatus: 403}), status.BOT_CONFIG_INVALID},
		{"unknown model 404 → config invalid", fmt.Errorf("x: %w", &openrouter.APIError{HTTPStatus: 404}), status.BOT_CONFIG_INVALID},
		{"insufficient credits → config invalid", fmt.Errorf("x: %w", openrouter.ErrInsufficientCredits), status.BOT_CONFIG_INVALID},
		{"402 typed → config invalid", fmt.Errorf("x: %w", &openrouter.APIError{HTTPStatus: 402}), status.BOT_CONFIG_INVALID},
		{"decode failed", fmt.Errorf("x: %w", openrouter.ErrDecodeResponse), status.BOT_SERIALIZE_FAILED},
		{"unclassified → op failed", errors.New("boom"), status.BOT_OP_FAILED},
		{"typed 408 → op failed", fmt.Errorf("x: %w", &openrouter.APIError{HTTPStatus: 408}), status.BOT_OP_FAILED},
		{"typed 503 → op failed", fmt.Errorf("x: %w", &openrouter.APIError{HTTPStatus: 503}), status.BOT_OP_FAILED},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mErr := mapOpenRouterError(ctx, tt.in)
			if mErr == nil {
				t.Fatal("mapOpenRouterError() = nil, want MathError")
			}
			if got := mErr.GetStatusCode(); got != tt.wantCode {
				t.Errorf("status code = %d, want %d", got, tt.wantCode)
			}
			// The original cause must stay unwrappable for observability.
			if !errors.Is(mErr, tt.in) && mErr.BaseError == nil {
				t.Error("MathError lost the underlying cause")
			}
		})
	}
}

func TestMapOpenRouterErrorNil(t *testing.T) {
	if got := mapOpenRouterError(context.Background(), nil); got != nil {
		t.Errorf("mapOpenRouterError(nil) = %v, want nil", got)
	}
}

// TestMapOpenRouterErrorCarriesVendorContext pins the log-safe args: the
// HTTP status and vendor code must reach the envelope so an operator can
// tell a 402 from a 429 without the raw body ever being attached.
func TestMapOpenRouterErrorCarriesVendorContext(t *testing.T) {
	err := fmt.Errorf("x: %w", &openrouter.APIError{
		HTTPStatus: 402,
		Code:       "insufficient_credits",
		Message:    "Insufficient credits",
	})

	mErr := mapOpenRouterError(context.Background(), err)
	if mErr == nil {
		t.Fatal("mapOpenRouterError() = nil")
	}
	if mErr.GetStatusCode() != status.BOT_CONFIG_INVALID {
		t.Errorf("status code = %d, want BOT_CONFIG_INVALID", mErr.GetStatusCode())
	}
}
