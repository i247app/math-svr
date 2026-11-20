package llmroute

import (
	"context"

	command "math-ai.com/math-ai/internal/application/command/llmroute"
	"math-ai.com/math-ai/internal/application/dto"
	query "math-ai.com/math-ai/internal/application/query/llmroute"
	"math-ai.com/math-ai/internal/application/transaction"
	domain "math-ai.com/math-ai/internal/domain/llmroute"
	"math-ai.com/math-ai/internal/infrastructure/logger"
	llmProvider "math-ai.com/math-ai/internal/provider/llm"
	sctx "math-ai.com/math-ai/internal/shared/context"
	"math-ai.com/math-ai/internal/shared/utils"
)

// Service is the admin surface over ma_llm_routes: list every route with
// what it resolves to, retarget one. Rows are never created or deleted
// here — the catalog is code (domain/llmroute/catalog.go).
type Service struct {
	registry  *llmProvider.Registry
	listQuery *query.ListLLMRoutesQuery
	resolve   *query.ResolveLLMRouteQuery
	updateCmd *command.UpdateLLMRouteCommandHandler
	// models caches each provider's model list for ModelListTTL.
	models *modelCache
}

// NewService wires the module. registry may be nil (no LLM configured):
// then every route resolves to nothing and only clearing is accepted.
func NewService(repo domain.IRepository, resolve *query.ResolveLLMRouteQuery, uow transaction.UnitOfWork, registry *llmProvider.Registry) *Service {
	return &Service{
		registry:  registry,
		listQuery: query.NewListLLMRoutesQuery(repo),
		resolve:   resolve,
		updateCmd: command.NewUpdateLLMRouteCommandHandler(uow),
		models:    newModelCache(ModelListTTL),
	}
}

func (s *Service) ListLLMRoutes(ctx context.Context, _ *dto.ListLLMRoutesReq) (*dto.ListLLMRoutesRes, error) {
	if _, err := requireAdminGate(ctx); err != nil {
		return nil, err
	}
	routes, err := s.listQuery.Handle(ctx)
	if err != nil {
		return nil, err
	}
	res := &dto.ListLLMRoutesRes{
		LLMRoutes:           make([]*dto.LLMRouteResponse, 0, len(routes)),
		RegisteredProviders: []string{},
	}
	for _, r := range routes {
		res.LLMRoutes = append(res.LLMRoutes, dto.LLMRouteToResponse(r, s.effective(ctx, r.RouteCode())))
	}
	if s.registry != nil {
		for _, n := range s.registry.Names() {
			res.RegisteredProviders = append(res.RegisteredProviders, n.String())
		}
	}
	return res, nil
}

// UpdateLLMRoute retargets one route. The next call of that kind, on every
// instance, reads the new target (no cache).
func (s *Service) UpdateLLMRoute(ctx context.Context, req *dto.UpdateLLMRouteReq) (*dto.LLMRouteRes, error) {
	via, err := requireAdminGate(ctx)
	if err != nil {
		return nil, err
	}
	v, err := ValidateUpdateLLMRoute(ctx, req, s.registry)
	if err != nil {
		return nil, err
	}
	saved, err := s.updateCmd.Handle(ctx, command.UpdateLLMRouteCommand{
		Code:     v.Code,
		Provider: v.Provider,
		Model:    v.Model,
		AdminUID: adminActor(ctx),
	})
	if err != nil {
		return nil, err
	}
	eff := s.effective(ctx, saved.RouteCode())
	logger.From(ctx).Infof("llm_route.updated admin_uid=%d via=%s code=%s provider=%s model=%s effective=%s",
		sctx.UID(ctx), via, saved.RouteCode(), utils.DerefString(saved.LLMProvider()), utils.DerefString(saved.LLMModel()), eff.Provider)

	res := &dto.LLMRouteRes{
		LLMRoute: dto.LLMRouteToResponse(saved, eff),
	}

	return res, nil
}

func (s *Service) effective(ctx context.Context, code domain.Code) dto.LLMRouteEffective {
	r := s.resolve.Handle(ctx, code)

	res := dto.LLMRouteEffective{
		Provider: r.Provider.String(),
		Model:    r.Model,
		Source:   r.Source.String(),
	}

	return res
}
