package bot

import (
	"context"
	"errors"

	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/libs/eino"
	"math-ai.com/math-ai/internal/libs/langchain"
)

// mapLangChainError translates a libs/langchain error into a domain-layer
// MathError, choosing the right status code based on the failure shape
// and attaching log-safe context (provider, backend).
//
// The mapping policy:
//
//   - langchain.ErrInvalidConfig / IsConfigError → BOT_CONFIG_INVALID.
//   - langchain.ErrUnsupportedOp                 → BOT_UNSUPPORTED_OP.
//   - langchain.ErrContextTooLarge               → BOT_CONTEXT_TOO_LARGE.
//   - langchain.ErrRateLimited / IsRateLimited   → BOT_RATE_LIMITED.
//   - langchain.ErrDecodeResponse                → BOT_SERIALIZE_FAILED.
//   - vendor 4xx auth                            → BOT_CONFIG_INVALID.
//   - anything else                              → BOT_OP_FAILED.
//
// API keys are never read here. The Backend value is intentionally
// included so log lines and the JSON envelope identify which LLM vendor
// produced the failure without exposing endpoint URLs or models.
func mapLangChainError(ctx context.Context, backend langchain.Backend, err error) *errs.MathError {
	if err == nil {
		return nil
	}

	args := map[string]any{
		"provider": string(ProviderLangChain),
		"backend":  string(backend),
	}

	switch {
	case errors.Is(err, langchain.ErrInvalidConfig), langchain.IsConfigError(err):
		return errs.NewError(ctx, status.BOT_CONFIG_INVALID, args, err)

	case errors.Is(err, langchain.ErrUnsupportedOp):
		return errs.NewError(ctx, status.BOT_UNSUPPORTED_OP, args, err)

	case errors.Is(err, langchain.ErrContextTooLarge):
		return errs.NewError(ctx, status.BOT_CONTEXT_TOO_LARGE, args, err)

	case errors.Is(err, langchain.ErrRateLimited), langchain.IsRateLimited(err):
		return errs.NewError(ctx, status.BOT_RATE_LIMITED, args, err)

	case errors.Is(err, langchain.ErrDecodeResponse):
		return errs.NewError(ctx, status.BOT_SERIALIZE_FAILED, args, err)
	}

	var apiErr *langchain.APIError
	if errors.As(err, &apiErr) {
		args["http_status"] = apiErr.HTTPStatus
		if apiErr.Code != "" {
			args["vendor_code"] = apiErr.Code
		}
	}
	return errs.NewError(ctx, status.BOT_OP_FAILED, args, err)
}

// mapEinoError translates a libs/eino error into a domain-layer MathError,
// mirroring mapLangChainError's policy one-for-one so the two providers
// surface identical status codes for identical failure shapes:
//
//   - eino.ErrInvalidConfig / IsConfigError → BOT_CONFIG_INVALID.
//   - eino.ErrUnsupportedOp                 → BOT_UNSUPPORTED_OP.
//   - eino.ErrContextTooLarge               → BOT_CONTEXT_TOO_LARGE.
//   - eino.ErrRateLimited / IsRateLimited   → BOT_RATE_LIMITED.
//   - eino.ErrDecodeResponse                → BOT_SERIALIZE_FAILED.
//   - anything else                         → BOT_OP_FAILED.
//
// API keys are never read here. The Backend value is intentionally
// included so log lines and the JSON envelope identify which LLM vendor
// produced the failure without exposing endpoint URLs or models.
func mapEinoError(ctx context.Context, backend eino.Backend, err error) *errs.MathError {
	if err == nil {
		return nil
	}

	args := map[string]any{
		"provider": string(ProviderEino),
		"backend":  string(backend),
	}

	switch {
	case errors.Is(err, eino.ErrInvalidConfig), eino.IsConfigError(err):
		return errs.NewError(ctx, status.BOT_CONFIG_INVALID, args, err)

	case errors.Is(err, eino.ErrUnsupportedOp):
		return errs.NewError(ctx, status.BOT_UNSUPPORTED_OP, args, err)

	case errors.Is(err, eino.ErrContextTooLarge):
		return errs.NewError(ctx, status.BOT_CONTEXT_TOO_LARGE, args, err)

	case errors.Is(err, eino.ErrRateLimited), eino.IsRateLimited(err):
		return errs.NewError(ctx, status.BOT_RATE_LIMITED, args, err)

	case errors.Is(err, eino.ErrDecodeResponse):
		return errs.NewError(ctx, status.BOT_SERIALIZE_FAILED, args, err)
	}

	var apiErr *eino.APIError
	if errors.As(err, &apiErr) {
		args["http_status"] = apiErr.HTTPStatus
		if apiErr.Code != "" {
			args["vendor_code"] = apiErr.Code
		}
	}
	return errs.NewError(ctx, status.BOT_OP_FAILED, args, err)
}
