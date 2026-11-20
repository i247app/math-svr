package dto

import (
	domain "math-ai.com/math-ai/internal/domain/llmroute"
	"math-ai.com/math-ai/internal/shared/mtime"
)

// ListLLMRoutesReq takes no filter: the catalog is a handful of rows.
type ListLLMRoutesReq struct{}

// UpdateLLMRouteReq retargets one route. Both fields overwrite the stored
// value: null (or "") clears it, so the route falls through to DEFAULT, then
// to env LLM_PROVIDER. llm_model needs an llm_provider.
type UpdateLLMRouteReq struct {
	RouteCode   string  `json:"route_code"`
	LLMProvider *string `json:"llm_provider"`
	LLMModel    *string `json:"llm_model"`
}

// LLMRouteResponse is one stored route plus what it resolves to on the
// instance that answered: effective_* shows a fall-through to DEFAULT / env,
// or a fallback when the stored provider has no API key here
// (effective_source = "fallback"). effective_model "" = the provider's model
// from env.
type LLMRouteResponse struct {
	LLMRouteID        int64          `json:"llm_route_id"`
	RouteCode         string         `json:"route_code"`
	LLMProvider       *string        `json:"llm_provider"`
	LLMModel          *string        `json:"llm_model"`
	Description       *string        `json:"description"`
	EffectiveProvider string         `json:"effective_provider"`
	EffectiveModel    string         `json:"effective_model"`
	EffectiveSource   string         `json:"effective_source"`
	ModifyID          *int64         `json:"modify_id,omitempty"`
	ModifyDt          mtime.MathTime `json:"modify_dt"`
}

type LLMRouteRes struct {
	LLMRoute *LLMRouteResponse `json:"llm_route"`
}

// ListLLMRoutesRes also lists the providers registered on the answering
// instance — the only values /llm-routes/update accepts.
type ListLLMRoutesRes struct {
	LLMRoutes           []*LLMRouteResponse `json:"llm_routes"`
	RegisteredProviders []string            `json:"registered_providers"`
}

// LLMRouteEffective is what a route resolves to; filled by the module.
type LLMRouteEffective struct {
	Provider string
	Model    string
	Source   string
}

func LLMRouteToResponse(r *domain.LLMRoute, eff LLMRouteEffective) *LLMRouteResponse {
	if r == nil {
		return nil
	}
	return &LLMRouteResponse{
		LLMRouteID:        r.LLMRouteId(),
		RouteCode:         r.RouteCode().String(),
		LLMProvider:       r.LLMProvider(),
		LLMModel:          r.LLMModel(),
		Description:       r.Description(),
		EffectiveProvider: eff.Provider,
		EffectiveModel:    eff.Model,
		EffectiveSource:   eff.Source,
		ModifyID:          r.ModifyId(),
		ModifyDt:          mtime.NewTime(r.ModifyDt()),
	}
}

// ListLLMModelsReq names one provider, or none for every registered one.
// chat_only (default false) keeps only the models that can serve a chat
// call, by each vendor's own hints (llm.IsChatModel).
type ListLLMModelsReq struct {
	Provider *string `json:"provider"`
	ChatOnly bool    `json:"chat_only"`
}

// LLMModel is one model a provider lists. id is the value llm_model takes
// on that provider; the other fields are the vendor's own hints, omitted
// when it reports none (owned_by: openai / deepseek; output_modalities,
// input_modalities: openrouter; supported_methods: gemini).
type LLMModel struct {
	ID               string   `json:"id"`
	Name             string   `json:"name,omitempty"`
	OwnedBy          string   `json:"owned_by,omitempty"`
	ContextLength    int      `json:"context_length,omitempty"`
	InputModalities  []string `json:"input_modalities,omitempty"`
	OutputModalities []string `json:"output_modalities,omitempty"`
	SupportedMethods []string `json:"supported_methods,omitempty"`
}

// LLMModelsError is why one provider's list is missing: the status code and
// its message only — never the vendor's raw text, which can quote the key.
type LLMModelsError struct {
	MStatus  int    `json:"mstatus"`
	MMessage string `json:"mmessage"`
}

// LLMProviderModels is one provider's list. On failure models is empty and
// error is set; the other providers are unaffected. fetched_at is when the
// vendor was called (cached = true: served from this instance's cache).
// model_count is what models holds; total_count is the vendor's list before
// chat_only filtering (equal when chat_only is false).
type LLMProviderModels struct {
	Provider   string          `json:"provider"`
	Models     []LLMModel      `json:"models"`
	ModelCount int             `json:"model_count"`
	TotalCount int             `json:"total_count"`
	FetchedAt  *mtime.MathTime `json:"fetched_at"`
	Cached     bool            `json:"cached"`
	Error      *LLMModelsError `json:"error"`
}

type ListLLMModelsRes struct {
	Providers []*LLMProviderModels `json:"providers"`
}
