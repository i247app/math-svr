package command

import (
	"context"

	"math-ai.com/math-ai/internal/application/transaction"
	roledom "math-ai.com/math-ai/internal/domain/role"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/errs"
)

// ForceDeleteRoleCommand physically removes the role row.
type ForceDeleteRoleCommand struct {
	RoleID int64
}

type ForceDeleteRoleCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewForceDeleteRoleCommandHandler(uow transaction.UnitOfWork) *ForceDeleteRoleCommandHandler {
	return &ForceDeleteRoleCommandHandler{uow: uow}
}

func (h *ForceDeleteRoleCommandHandler) Handle(ctx context.Context, cmd ForceDeleteRoleCommand) error {
	return h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		existing, err := repos.Role.FindByRoleId(ctx, cmd.RoleID)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if existing == nil {
			return errs.NewError(ctx, roledom.RoleNotFound, nil, ErrRoleNotFound)
		}
		if err := repos.Role.ForceDeleteByRoleId(ctx, cmd.RoleID); err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		return nil
	})
}
