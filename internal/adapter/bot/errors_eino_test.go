package bot

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/libs/eino"
	"math-ai.com/math-ai/internal/libs/langchain"
)

func TestMapEinoError(t *testing.T) {
	ctx := context.Background()
	backend := eino.BackendGoogleAI

	tests := []struct {
		name     string
		in       error
		wantCode status.StatusCode
	}{
		{"invalid config", fmt.Errorf("x: %w", eino.ErrInvalidConfig), status.BOT_CONFIG_INVALID},
		{"unsupported op", fmt.Errorf("x: %w", eino.ErrUnsupportedOp), status.BOT_UNSUPPORTED_OP},
		{"context too large", fmt.Errorf("x: %w", eino.ErrContextTooLarge), status.BOT_CONTEXT_TOO_LARGE},
		{"rate limited sentinel", fmt.Errorf("x: %w", eino.ErrRateLimited), status.BOT_RATE_LIMITED},
		{"rate limited typed 429", fmt.Errorf("x: %w", &eino.APIError{Backend: backend, HTTPStatus: 429}), status.BOT_RATE_LIMITED},
		{"auth 401 → config invalid", fmt.Errorf("x: %w", &eino.APIError{Backend: backend, HTTPStatus: 401}), status.BOT_CONFIG_INVALID},
		{"decode failed", fmt.Errorf("x: %w", eino.ErrDecodeResponse), status.BOT_SERIALIZE_FAILED},
		{"unclassified → op failed", errors.New("boom"), status.BOT_OP_FAILED},
		{"typed 503 → op failed", fmt.Errorf("x: %w", &eino.APIError{Backend: backend, HTTPStatus: 503}), status.BOT_OP_FAILED},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mErr := mapEinoError(ctx, backend, tt.in)
			if mErr == nil {
				t.Fatal("mapEinoError() = nil, want MathError")
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

	if mapEinoError(ctx, backend, nil) != nil {
		t.Error("mapEinoError(nil) should be nil")
	}
}

// TestMapEinoErrorParityWithLangChain locks the guarantee that the two
// providers surface the SAME status code for the same failure shape, so
// swapping BOT_PROVIDER never changes the wire contract.
func TestMapEinoErrorParityWithLangChain(t *testing.T) {
	ctx := context.Background()

	pairs := []struct {
		name    string
		einoErr error
		langErr error
	}{
		{"rate limited", eino.ErrRateLimited, langchain.ErrRateLimited},
		{"context too large", eino.ErrContextTooLarge, langchain.ErrContextTooLarge},
		{"unsupported op", eino.ErrUnsupportedOp, langchain.ErrUnsupportedOp},
		{"decode failed", eino.ErrDecodeResponse, langchain.ErrDecodeResponse},
		{"invalid config", eino.ErrInvalidConfig, langchain.ErrInvalidConfig},
	}

	for _, p := range pairs {
		t.Run(p.name, func(t *testing.T) {
			e := mapEinoError(ctx, eino.BackendGoogleAI, p.einoErr)
			l := mapLangChainError(ctx, langchain.BackendGoogleAI, p.langErr)
			if e.GetStatusCode() != l.GetStatusCode() {
				t.Errorf("status divergence: eino=%d langchain=%d",
					e.GetStatusCode(), l.GetStatusCode())
			}
		})
	}
}
