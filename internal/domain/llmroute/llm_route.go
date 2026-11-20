package llmroute

import (
	"math-ai.com/math-ai/internal/domain/shared/audit"
)

// LLMRoute says which LLM provider (and optionally model) one kind of LLM
// call goes through. A nil provider falls through to the DEFAULT route, then
// to env LLM_PROVIDER; a nil model uses the provider's model from env. The
// provider is a name checked against the LLM registry by the API, not here.
type LLMRoute struct {
	audit.Audit

	llmRouteId  int64
	routeCode   Code
	llmProvider *string
	llmModel    *string
	description *string

	llmRouteStatus *string
}

func NewLLMRoute() *LLMRoute { return &LLMRoute{} }

func (r *LLMRoute) LLMRouteId() int64      { return r.llmRouteId }
func (r *LLMRoute) SetLLMRouteId(id int64) { r.llmRouteId = id }

func (r *LLMRoute) RouteCode() Code     { return r.routeCode }
func (r *LLMRoute) SetRouteCode(c Code) { r.routeCode = c }

func (r *LLMRoute) LLMProvider() *string     { return r.llmProvider }
func (r *LLMRoute) SetLLMProvider(p *string) { r.llmProvider = p }

func (r *LLMRoute) LLMModel() *string     { return r.llmModel }
func (r *LLMRoute) SetLLMModel(m *string) { r.llmModel = m }

func (r *LLMRoute) Description() *string     { return r.description }
func (r *LLMRoute) SetDescription(s *string) { r.description = s }

func (r *LLMRoute) LLMRouteStatus() *string     { return r.llmRouteStatus }
func (r *LLMRoute) SetLLMRouteStatus(s *string) { r.llmRouteStatus = s }
