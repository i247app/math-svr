package query

import (
	"context"
	"strings"

	"math-ai.com/math-ai/internal/domain/llmroute"
	"math-ai.com/math-ai/internal/infrastructure/logger"
	llmProvider "math-ai.com/math-ai/internal/provider/llm"
)

// Source says which step of the resolution picked a route's provider.
type Source string

const (
	SourceRow      Source = "row"      // the code's own row
	SourceDefault  Source = "default"  // the DEFAULT row
	SourceEnv      Source = "env"      // no row names a provider: env LLM_PROVIDER
	SourceFallback Source = "fallback" // a row named a provider this instance lacks, or the read failed
)

func (s Source) String() string {
	return string(s)
}

// Route is the provider (and optional model) one LLM call goes through.
// An empty Model means the provider's model from env.
type Route struct {
	Provider llmProvider.LLMProviderName
	Model    string
	Source   Source
}

// ResolveLLMRouteQuery picks the provider of one kind of LLM call:
// the code's row → the DEFAULT row → env LLM_PROVIDER. A provider not
// registered on this instance (no API key) and a failed read both fall back
// to env with a warning: routing must never take a model call down.
//
// It reads the database on every call, with no cache, so an admin update
// reaches every instance on its next call.
type ResolveLLMRouteQuery struct {
	repo     llmroute.IRepository
	registry *llmProvider.Registry
}

func NewResolveLLMRouteQuery(repo llmroute.IRepository, registry *llmProvider.Registry) *ResolveLLMRouteQuery {
	return &ResolveLLMRouteQuery{repo: repo, registry: registry}
}

func (q *ResolveLLMRouteQuery) Handle(ctx context.Context, code llmroute.Code) Route {
	if q.registry == nil {
		// No LLM configured at all: callers already fail on the nil provider.
		return Route{Source: SourceEnv}
	}
	envRoute := Route{Provider: q.registry.DefaultName(), Source: SourceEnv}
	log := logger.From(ctx)

	row, source, err := q.target(ctx, code)
	if err != nil {
		log.Warnf("llm.route.fallback code=%s reason=read_failed used=%s err=%v", code, envRoute.Provider, err)
		envRoute.Source = SourceFallback
		return envRoute
	}
	if row == nil {
		return envRoute
	}

	// Normalised here, not only by the API validator: a row edited by hand
	// in the database ("gpt-4o-mini  ") would otherwise reach the vendor as
	// an invalid model id.
	name := llmProvider.LLMProviderName(strings.ToLower(strings.TrimSpace(*row.LLMProvider())))
	if !q.registry.Has(name) {
		// The model belongs to the provider that is missing, so it goes too.
		log.Warnf("llm.route.fallback code=%s reason=not_registered wanted=%s used=%s", code, name, envRoute.Provider)
		envRoute.Source = SourceFallback
		return envRoute
	}
	route := Route{Provider: name, Source: source}
	if m := row.LLMModel(); m != nil {
		route.Model = strings.TrimSpace(*m)
	}
	return route
}

// target returns the first row that names a provider — the code's own, then
// DEFAULT — or nil when neither does.
func (q *ResolveLLMRouteQuery) target(ctx context.Context, code llmroute.Code) (*llmroute.LLMRoute, Source, error) {
	row, err := q.repo.FindByCode(ctx, code)
	if err != nil {
		return nil, "", err
	}
	if row != nil && row.LLMProvider() != nil {
		return row, SourceRow, nil
	}
	if code == llmroute.CodeDefault {
		return nil, "", nil
	}
	def, err := q.repo.FindByCode(ctx, llmroute.CodeDefault)
	if err != nil {
		return nil, "", err
	}
	if def != nil && def.LLMProvider() != nil {
		return def, SourceDefault, nil
	}
	return nil, "", nil
}
