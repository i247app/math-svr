package bot

import (
	"context"
	"errors"

	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/libs/eino"
	"math-ai.com/math-ai/internal/libs/gemini"
	"math-ai.com/math-ai/internal/libs/langchain"
	"math-ai.com/math-ai/internal/libs/openai"
	"math-ai.com/math-ai/internal/libs/openrouter"
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

// mapOpenRouterError translates a libs/openrouter error into a
// domain-layer MathError, mirroring mapEinoError's policy so all three
// providers surface identical status codes for identical failure shapes:
//
//   - openrouter.ErrInvalidConfig / IsConfigError → BOT_CONFIG_INVALID.
//   - openrouter.ErrInsufficientCredits (HTTP 402) → BOT_CONFIG_INVALID.
//   - openrouter.ErrUnsupportedOp                 → BOT_UNSUPPORTED_OP.
//   - openrouter.ErrContextTooLarge               → BOT_CONTEXT_TOO_LARGE.
//   - openrouter.ErrRateLimited / IsRateLimited   → BOT_RATE_LIMITED.
//   - openrouter.ErrDecodeResponse                → BOT_SERIALIZE_FAILED.
//   - HTTP 401 / 403 (auth, moderation block)     → BOT_CONFIG_INVALID.
//   - anything else (incl. 408 / 5xx)             → BOT_OP_FAILED.
//
// HTTP 402 rides with the config failures on purpose: an exhausted
// balance cannot recover inside the call and needs an operator, which is
// exactly what BOT_CONFIG_INVALID signals to the caller.
//
// There is no backend arg — OpenRouter routes to the vendor itself, so
// the model id is the closest equivalent and it is carried on the
// underlying error, not read here. API keys are never read here, and the
// raw upstream body never reaches the args (libs/openrouter reports an
// undecodable body by size only).
func mapOpenRouterError(ctx context.Context, err error) *errs.MathError {
	if err == nil {
		return nil
	}

	args := map[string]any{"provider": string(ProviderOpenRouter)}

	var apiErr *openrouter.APIError
	if errors.As(err, &apiErr) {
		args["http_status"] = apiErr.HTTPStatus
		if apiErr.Code != "" {
			args["vendor_code"] = apiErr.Code
		}
	}

	switch {
	case errors.Is(err, openrouter.ErrInvalidConfig),
		errors.Is(err, openrouter.ErrInsufficientCredits),
		openrouter.IsConfigError(err):
		return errs.NewError(ctx, status.BOT_CONFIG_INVALID, args, err)

	case errors.Is(err, openrouter.ErrUnsupportedOp):
		return errs.NewError(ctx, status.BOT_UNSUPPORTED_OP, args, err)

	case errors.Is(err, openrouter.ErrContextTooLarge):
		return errs.NewError(ctx, status.BOT_CONTEXT_TOO_LARGE, args, err)

	case errors.Is(err, openrouter.ErrRateLimited), openrouter.IsRateLimited(err):
		return errs.NewError(ctx, status.BOT_RATE_LIMITED, args, err)

	case errors.Is(err, openrouter.ErrDecodeResponse):
		return errs.NewError(ctx, status.BOT_SERIALIZE_FAILED, args, err)
	}

	return errs.NewError(ctx, status.BOT_OP_FAILED, args, err)
}

// mapOpenAIError translates a libs/openai error into a domain-layer
// MathError, mirroring its siblings so all four providers surface
// identical status codes for identical failure shapes:
//
//   - openai.ErrInvalidConfig / IsConfigError → BOT_CONFIG_INVALID.
//   - openai.ErrQuotaExhausted (billing 429)  → BOT_CONFIG_INVALID.
//   - openai.ErrUnsupportedOp                 → BOT_UNSUPPORTED_OP.
//   - openai.ErrContextTooLarge               → BOT_CONTEXT_TOO_LARGE.
//   - openai.ErrRateLimited / IsRateLimited   → BOT_RATE_LIMITED.
//   - openai.ErrDecodeResponse                → BOT_SERIALIZE_FAILED.
//   - HTTP 401 / 403 (bad key, geo block)     → BOT_CONFIG_INVALID.
//   - anything else (incl. 408 / 5xx)         → BOT_OP_FAILED.
//
// The 429 split is the one asymmetry worth knowing: OpenAI reuses that
// status for both throttling and an exhausted credit balance. Throttling
// is transient (BOT_RATE_LIMITED, retried upstream); billing is not, so it
// rides with the config failures — the same posture as openrouter's HTTP
// 402, which is the equivalent signal on that provider.
//
// API keys are never read here, and the raw upstream body never reaches
// the args (libs/openai reports an undecodable body by size only).
func mapOpenAIError(ctx context.Context, err error) *errs.MathError {
	if err == nil {
		return nil
	}

	args := map[string]any{"provider": string(ProviderOpenAI)}

	var apiErr *openai.APIError
	if errors.As(err, &apiErr) {
		args["http_status"] = apiErr.HTTPStatus
		if apiErr.Type != "" {
			args["vendor_type"] = apiErr.Type
		}
		if apiErr.Code != "" {
			args["vendor_code"] = apiErr.Code
		}
	}

	switch {
	case errors.Is(err, openai.ErrInvalidConfig),
		errors.Is(err, openai.ErrQuotaExhausted),
		openai.IsConfigError(err):
		return errs.NewError(ctx, status.BOT_CONFIG_INVALID, args, err)

	case errors.Is(err, openai.ErrUnsupportedOp):
		return errs.NewError(ctx, status.BOT_UNSUPPORTED_OP, args, err)

	case errors.Is(err, openai.ErrContextTooLarge):
		return errs.NewError(ctx, status.BOT_CONTEXT_TOO_LARGE, args, err)

	case errors.Is(err, openai.ErrRateLimited), openai.IsRateLimited(err):
		return errs.NewError(ctx, status.BOT_RATE_LIMITED, args, err)

	case errors.Is(err, openai.ErrDecodeResponse):
		return errs.NewError(ctx, status.BOT_SERIALIZE_FAILED, args, err)
	}

	return errs.NewError(ctx, status.BOT_OP_FAILED, args, err)
}

// mapGeminiError translates a libs/gemini error into a domain-layer
// MathError, mirroring its siblings so all five providers surface
// identical status codes for identical failure shapes:
//
//   - gemini.ErrInvalidConfig / IsConfigError → BOT_CONFIG_INVALID.
//   - gemini.ErrQuotaExhausted (daily 429)    → BOT_CONFIG_INVALID.
//   - gemini.ErrContentBlocked                → BOT_CONTENT_BLOCKED.
//   - gemini.ErrUnsupportedOp                 → BOT_UNSUPPORTED_OP.
//   - gemini.ErrContextTooLarge               → BOT_CONTEXT_TOO_LARGE.
//   - gemini.ErrRateLimited / IsRateLimited   → BOT_RATE_LIMITED.
//   - gemini.ErrDecodeResponse                → BOT_SERIALIZE_FAILED.
//   - HTTP 401 / 403 (bad key, no permission) → BOT_CONFIG_INVALID.
//   - anything else (incl. 408 / 5xx)         → BOT_OP_FAILED.
//
// Two mappings are Gemini-specific. BOT_CONTENT_BLOCKED exists because a
// safety or recitation refusal is not a transient fault: the same prompt
// is refused every time, so folding it into BOT_OP_FAILED would invite a
// caller to retry forever. And the 429 split mirrors the OpenAI provider's
// — a burst limit is transient, a daily quota is not.
//
// API keys are never read here, and the raw upstream body never reaches
// the args (libs/gemini reports an undecodable body by size only).
func mapGeminiError(ctx context.Context, err error) *errs.MathError {
	if err == nil {
		return nil
	}

	args := map[string]any{"provider": string(ProviderGemini)}

	var apiErr *gemini.APIError
	if errors.As(err, &apiErr) {
		args["http_status"] = apiErr.HTTPStatus
		if apiErr.Status != "" {
			args["vendor_status"] = apiErr.Status
		}
		if apiErr.Code != "" {
			args["vendor_code"] = apiErr.Code
		}
	}

	switch {
	case errors.Is(err, gemini.ErrContentBlocked):
		return errs.NewError(ctx, status.BOT_CONTENT_BLOCKED, args, err)

	case errors.Is(err, gemini.ErrInvalidConfig),
		errors.Is(err, gemini.ErrQuotaExhausted),
		gemini.IsConfigError(err):
		return errs.NewError(ctx, status.BOT_CONFIG_INVALID, args, err)

	case errors.Is(err, gemini.ErrUnsupportedOp):
		return errs.NewError(ctx, status.BOT_UNSUPPORTED_OP, args, err)

	case errors.Is(err, gemini.ErrContextTooLarge):
		return errs.NewError(ctx, status.BOT_CONTEXT_TOO_LARGE, args, err)

	case errors.Is(err, gemini.ErrRateLimited), gemini.IsRateLimited(err):
		return errs.NewError(ctx, status.BOT_RATE_LIMITED, args, err)

	case errors.Is(err, gemini.ErrDecodeResponse):
		return errs.NewError(ctx, status.BOT_SERIALIZE_FAILED, args, err)
	}

	return errs.NewError(ctx, status.BOT_OP_FAILED, args, err)
}
