package command

import (
	"context"
	"fmt"

	"math-ai.com/math-ai/internal/application/transaction"
	"math-ai.com/math-ai/internal/domain/llmroute"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
)

// UpdateLLMRouteCommand retargets one route. Provider and Model overwrite
// the stored values; nil clears either (the route then falls through to
// DEFAULT / env). Whether Provider is registered is the module's check —
// it needs the LLM registry, which this layer does not hold.
type UpdateLLMRouteCommand struct {
	Code     llmroute.Code
	Provider *string
	Model    *string
	AdminUID *int64 // nil = no uid (API-key admin)
}

type UpdateLLMRouteCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewUpdateLLMRouteCommandHandler(uow transaction.UnitOfWork) *UpdateLLMRouteCommandHandler {
	return &UpdateLLMRouteCommandHandler{uow: uow}
}

// Handle writes first and reads back after, in the same transaction: the
// UPDATE itself reports whether the row exists.
func (h *UpdateLLMRouteCommandHandler) Handle(ctx context.Context, cmd UpdateLLMRouteCommand) (*llmroute.LLMRoute, error) {
	var saved *llmroute.LLMRoute
	err := h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		updated, err := repos.LLMRoute.UpdateTarget(ctx, cmd.Code, cmd.Provider, cmd.Model, cmd.AdminUID)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if !updated {
			return errs.NewError(ctx, status.LLM_ROUTE_NOT_FOUND, nil,
				fmt.Errorf("llm route: ma_llm_routes has no live row %s", cmd.Code))
		}
		if saved, err = repos.LLMRoute.FindByCode(ctx, cmd.Code); err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return saved, nil
}
