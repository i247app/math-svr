package command

import (
	"context"
	"log"

	"github.com/google/uuid"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/enum"

	"math-ai.com/math-ai/internal/application/transaction"
)

type SoftDeleteUserCommand struct {
	UserID uuid.UUID
}

type SoftDeleteUserCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewSoftDeleteUserCommandHandler(uow transaction.UnitOfWork) *SoftDeleteUserCommandHandler {
	return &SoftDeleteUserCommandHandler{uow: uow}
}

func (h *SoftDeleteUserCommandHandler) Handle(ctx context.Context, cmd SoftDeleteUserCommand) error {
	err := h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		if err := repos.User.MarkStatusByUserId(ctx, cmd.UserID, enum.UserStatusTypeDeleted); err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if err := repos.Alias.MarkStatusByUserId(ctx, cmd.UserID, enum.UserAliasStatusTypeDeleted); err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		return nil
	})
	if err != nil {
		log.Printf("[admin.soft_delete_user] user_id=%d action=soft_delete outcome=error err=%v", cmd.UserID, err)
		return err
	}
	log.Printf("[admin.soft_delete_user] user_id=%d action=soft_delete outcome=success", cmd.UserID)
	return nil
}
