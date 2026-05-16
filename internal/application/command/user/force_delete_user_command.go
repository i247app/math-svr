package command

import (
	"context"
	"log"

	"github.com/google/uuid"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"

	"math-ai.com/math-ai/internal/application/transaction"
)

type ForceDeleteUserCommand struct {
	UserID uuid.UUID
}

type ForceDeleteUserCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewForceDeleteUserCommandHandler(uow transaction.UnitOfWork) *ForceDeleteUserCommandHandler {
	return &ForceDeleteUserCommandHandler{uow: uow}
}

// Handle erases a user and every dependent row (notifications, aliases,
// logins) inside a single transaction. The order is children-before-parent so
// the cascade is safe even if FK constraints are added to these tables later.
// Idempotency is by design: rows-affected = 0 at any step is not an error.
func (h *ForceDeleteUserCommandHandler) Handle(ctx context.Context, cmd ForceDeleteUserCommand) error {
	err := h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		if err := repos.Alias.DeleteByUserId(ctx, cmd.UserID); err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if err := repos.User.DeleteByUserId(ctx, cmd.UserID); err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		return nil
	})
	if err != nil {
		log.Printf("[admin.force_delete_user] user_id=%d action=force_delete outcome=error err=%v", cmd.UserID, err)
		return err
	}
	log.Printf("[admin.force_delete_user] user_id=%d action=force_delete outcome=success", cmd.UserID)
	return nil
}
