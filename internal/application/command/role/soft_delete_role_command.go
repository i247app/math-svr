package command

import (
	"context"

	"math-ai.com/math-ai/internal/application/transaction"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
)

// SoftDeleteRoleCommand hides the role from every read and frees its code;
// the row stays for audit/recovery.
type SoftDeleteRoleCommand struct {
	RoleID int64
}

type SoftDeleteRoleCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewSoftDeleteRoleCommandHandler(uow transaction.UnitOfWork) *SoftDeleteRoleCommandHandler {
	return &SoftDeleteRoleCommandHandler{uow: uow}
}

func (h *SoftDeleteRoleCommandHandler) Handle(ctx context.Context, cmd SoftDeleteRoleCommand) error {
	return h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		existing, err := repos.Role.FindByRoleId(ctx, cmd.RoleID)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if existing == nil {
			return errs.NewError(ctx, status.ROLE_NOT_FOUND, nil, ErrRoleNotFound)
		}
		if err := repos.Role.SoftDeleteByRoleId(ctx, cmd.RoleID); err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		return nil
	})
}
