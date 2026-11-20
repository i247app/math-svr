package llmroute

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"math-ai.com/math-ai/internal/application/dto"
	domain "math-ai.com/math-ai/internal/domain/llmroute"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	llmProvider "math-ai.com/math-ai/internal/provider/llm"
)

// MaxModelRunes matches ma_llm_routes.llm_model VARCHAR(128).
const MaxModelRunes = 128

var (
	ErrRouteCodeRequired  = errors.New("route_code is required")
	ErrModelWithoutTarget = errors.New("llm_model needs an llm_provider")
	ErrModelTooLong       = errors.New("llm_model is longer than 128 characters")
)

// validatedUpdate is an update with its values normalised: code upper-case,
// provider lower-case, blanks turned into nil (= clear).
type validatedUpdate struct {
	Code     domain.Code
	Provider *string
	Model    *string
}

// ValidateUpdateLLMRoute checks an update against the catalog and against
// the providers registered on this instance (registry may be nil when no
// LLM is configured: then no provider is accepted, only clearing).
func ValidateUpdateLLMRoute(ctx context.Context, req *dto.UpdateLLMRouteReq, registry *llmProvider.Registry) (validatedUpdate, error) {
	code := domain.Code(strings.ToUpper(strings.TrimSpace(req.RouteCode)))
	if code == "" {
		return validatedUpdate{}, errs.NewError(ctx, status.LLM_ROUTE_MISSING_CODE, nil, ErrRouteCodeRequired)
	}
	if !domain.IsKnown(code) {
		return validatedUpdate{}, errs.NewError(ctx, status.LLM_ROUTE_NOT_FOUND, nil,
			fmt.Errorf("route_code %q is not in the catalog", code))
	}

	v := validatedUpdate{Code: code, Provider: blankToNil(req.LLMProvider, true), Model: blankToNil(req.LLMModel, false)}
	if v.Provider != nil {
		name := llmProvider.LLMProviderName(*v.Provider)
		if registry == nil || !registry.Has(name) {
			return validatedUpdate{}, errs.NewError(ctx, status.LLM_ROUTE_INVALID_PROVIDER, nil,
				fmt.Errorf("llm_provider %q is not registered (registered: %s)", name, registeredList(registry)))
		}
	}
	if v.Model != nil {
		if v.Provider == nil {
			return validatedUpdate{}, errs.NewError(ctx, status.LLM_ROUTE_INVALID_MODEL, nil, ErrModelWithoutTarget)
		}
		if utf8.RuneCountInString(*v.Model) > MaxModelRunes {
			return validatedUpdate{}, errs.NewError(ctx, status.LLM_ROUTE_INVALID_MODEL, nil, ErrModelTooLong)
		}
	}
	return v, nil
}

// blankToNil trims s; nil or blank → nil. Provider names are lower-case
// (llm.Provider* constants); model ids are kept as written.
func blankToNil(s *string, lower bool) *string {
	if s == nil {
		return nil
	}
	t := strings.TrimSpace(*s)
	if lower {
		t = strings.ToLower(t)
	}
	if t == "" {
		return nil
	}
	return &t
}

func registeredList(registry *llmProvider.Registry) string {
	if registry == nil {
		return "none"
	}
	names := registry.Names()
	out := make([]string, len(names))
	for i, n := range names {
		out[i] = string(n)
	}
	return strings.Join(out, ", ")
}
