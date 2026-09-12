package bot

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/libs/openai"
)

func TestMapOpenAIError(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		in       error
		wantCode status.StatusCode
	}{
		{"invalid config", fmt.Errorf("x: %w", openai.ErrInvalidConfig), status.BOT_CONFIG_INVALID},
		{"unsupported op", fmt.Errorf("x: %w", openai.ErrUnsupportedOp), status.BOT_UNSUPPORTED_OP},
		{"context too large", fmt.Errorf("x: %w", openai.ErrContextTooLarge), status.BOT_CONTEXT_TOO_LARGE},
		{"rate limited sentinel", fmt.Errorf("x: %w", openai.ErrRateLimited), status.BOT_RATE_LIMITED},
		{"throttling 429", fmt.Errorf("x: %w", &openai.APIError{HTTPStatus: 429}), status.BOT_RATE_LIMITED},
		{"bad key 401 → config invalid", fmt.Errorf("x: %w", &openai.APIError{HTTPStatus: 401}), status.BOT_CONFIG_INVALID},
		{"geo block 403 → config invalid", fmt.Errorf("x: %w", &openai.APIError{HTTPStatus: 403}), status.BOT_CONFIG_INVALID},
		{"unknown model 404 → config invalid", fmt.Errorf("x: %w", &openai.APIError{HTTPStatus: 404}), status.BOT_CONFIG_INVALID},
		{"quota sentinel → config invalid", fmt.Errorf("x: %w", openai.ErrQuotaExhausted), status.BOT_CONFIG_INVALID},
		{"decode failed", fmt.Errorf("x: %w", openai.ErrDecodeResponse), status.BOT_SERIALIZE_FAILED},
		{"unclassified → op failed", errors.New("boom"), status.BOT_OP_FAILED},
		{"408 → op failed", fmt.Errorf("x: %w", &openai.APIError{HTTPStatus: 408}), status.BOT_OP_FAILED},
		{"500 → op failed", fmt.Errorf("x: %w", &openai.APIError{HTTPStatus: 500}), status.BOT_OP_FAILED},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mErr := mapOpenAIError(ctx, tt.in)
			if mErr == nil {
				t.Fatal("mapOpenAIError() = nil, want MathError")
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

// TestMapOpenAIErrorBilling429 is the case that separates this mapper from
// the openrouter one: OpenAI reuses 429 for an exhausted credit balance,
// and that must reach the caller as a config problem, not as transient
// throttling the caller might sit and retry.
func TestMapOpenAIErrorBilling429(t *testing.T) {
	err := fmt.Errorf("x: %w", &openai.APIError{
		HTTPStatus: 429,
		Type:       "rate_limit_error",
		Code:       "credit_balance_exhausted",
		Message:    "Credit balance exhausted",
	})

	mErr := mapOpenAIError(context.Background(), err)
	if mErr == nil {
		t.Fatal("mapOpenAIError() = nil")
	}
	if got := mErr.GetStatusCode(); got != status.BOT_CONFIG_INVALID {
		t.Errorf("status code = %d, want BOT_CONFIG_INVALID (not BOT_RATE_LIMITED)", got)
	}
}

func TestMapOpenAIErrorNil(t *testing.T) {
	if got := mapOpenAIError(context.Background(), nil); got != nil {
		t.Errorf("mapOpenAIError(nil) = %v, want nil", got)
	}
}

// TestMapOpenAIErrorResponsesSentinels covers the two errors only the
// Responses path can produce. A dead chain pointer must stay BOT_OP_FAILED
// (not config, not transient) yet remain recognisable through
// IsChainPointerGone, which is the single predicate the exam module uses
// to decide "rebuild the chain".
func TestMapOpenAIErrorResponsesSentinels(t *testing.T) {
	ctx := context.Background()

	gone := fmt.Errorf("%w: %w", openai.ErrPreviousResponseGone, &openai.APIError{HTTPStatus: 404})
	mErr := mapOpenAIError(ctx, gone)
	if got := mErr.GetStatusCode(); got != status.BOT_OP_FAILED {
		t.Errorf("gone pointer status = %d, want BOT_OP_FAILED", got)
	}
	if !IsChainPointerGone(mErr) {
		t.Error("IsChainPointerGone(mapped) = false, want true")
	}
	if !IsChainPointerGone(gone) {
		t.Error("IsChainPointerGone(raw sentinel) = false, want true")
	}

	refused := fmt.Errorf("x: %w", openai.ErrContentRefused)
	if got := mapOpenAIError(ctx, refused).GetStatusCode(); got != status.BOT_CONTENT_BLOCKED {
		t.Errorf("refusal status = %d, want BOT_CONTENT_BLOCKED", got)
	}

	// Negative space: an ordinary failure must not be mistaken for a dead
	// pointer, or the module would rebuild a healthy chain on every 500.
	for _, other := range []error{
		mapOpenAIError(ctx, fmt.Errorf("x: %w", &openai.APIError{HTTPStatus: 500})),
		mapOpenAIError(ctx, fmt.Errorf("x: %w", openai.ErrDecodeResponse)),
		errors.New("boom"),
		nil,
	} {
		if IsChainPointerGone(other) {
			t.Errorf("IsChainPointerGone(%v) = true, want false", other)
		}
	}
}
