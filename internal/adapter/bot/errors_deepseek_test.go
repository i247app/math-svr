package bot

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/libs/deepseek"
)

func TestMapDeepSeekError(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		in       error
		wantCode status.StatusCode
	}{
		{"invalid config", fmt.Errorf("x: %w", deepseek.ErrInvalidConfig), status.BOT_CONFIG_INVALID},
		{"unsupported op", fmt.Errorf("x: %w", deepseek.ErrUnsupportedOp), status.BOT_UNSUPPORTED_OP},
		{"context too large", fmt.Errorf("x: %w", deepseek.ErrContextTooLarge), status.BOT_CONTEXT_TOO_LARGE},
		{"rate limited sentinel", fmt.Errorf("x: %w", deepseek.ErrRateLimited), status.BOT_RATE_LIMITED},
		{"429", fmt.Errorf("x: %w", &deepseek.APIError{HTTPStatus: 429}), status.BOT_RATE_LIMITED},
		{"bad key 401 → config invalid", fmt.Errorf("x: %w", &deepseek.APIError{HTTPStatus: 401}), status.BOT_CONFIG_INVALID},
		{"402 typed → config invalid", fmt.Errorf("x: %w", &deepseek.APIError{HTTPStatus: 402}), status.BOT_CONFIG_INVALID},
		{"insufficient balance → config invalid", fmt.Errorf("x: %w", deepseek.ErrInsufficientBalance), status.BOT_CONFIG_INVALID},
		{"content filtered → blocked", fmt.Errorf("x: %w", deepseek.ErrContentFiltered), status.BOT_CONTENT_BLOCKED},
		{"decode failed", fmt.Errorf("x: %w", deepseek.ErrDecodeResponse), status.BOT_SERIALIZE_FAILED},
		{"server overloaded → op failed", fmt.Errorf("x: %w", deepseek.ErrServerOverloaded), status.BOT_OP_FAILED},
		{"unclassified → op failed", errors.New("boom"), status.BOT_OP_FAILED},
		{"422 → op failed", fmt.Errorf("x: %w", &deepseek.APIError{HTTPStatus: 422}), status.BOT_OP_FAILED},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mErr := mapDeepSeekError(ctx, tt.in)
			if mErr == nil {
				t.Fatal("mapDeepSeekError() = nil, want MathError")
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

// TestDeepSeekBillingMatchesSiblings pins the cross-provider consistency
// that makes these six interchangeable: an empty balance must produce the
// same status code whichever vendor reports it, even though DeepSeek and
// OpenRouter say 402 while OpenAI says 429-plus-code.
func TestDeepSeekBillingMatchesSiblings(t *testing.T) {
	err := fmt.Errorf("x: %w", &deepseek.APIError{
		HTTPStatus: 402,
		Type:       "insufficient_balance",
		Message:    "Insufficient Balance",
	})

	mErr := mapDeepSeekError(context.Background(), err)
	if mErr == nil {
		t.Fatal("mapDeepSeekError() = nil")
	}
	if got := mErr.GetStatusCode(); got != status.BOT_CONFIG_INVALID {
		t.Errorf("status code = %d, want BOT_CONFIG_INVALID (not a transient code)", got)
	}
}

func TestMapDeepSeekErrorNil(t *testing.T) {
	if got := mapDeepSeekError(context.Background(), nil); got != nil {
		t.Errorf("mapDeepSeekError(nil) = %v, want nil", got)
	}
}
