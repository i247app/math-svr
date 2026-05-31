package command

import (
	"context"
	"log"

	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"

	"math-ai.com/math-ai/internal/application/transaction"
)

type ForceDeleteUserCommand struct {
	UserID int64
}

// ForceDeleteUserCommandResult carries side-effect cleanup data the service
// layer needs to act on after the transaction commits — currently the avatar
// S3 keys to evict.
type ForceDeleteUserCommandResult struct {
	AvatarKeys []string
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
//
// Avatar keys are collected BEFORE the profile rows are wiped so the caller
// can drop them from S3 after the tx commits. We deliberately don't delete
// from S3 inside the tx — slow I/O must not hold a DB lock.
func (h *ForceDeleteUserCommandHandler) Handle(ctx context.Context, cmd ForceDeleteUserCommand) (*ForceDeleteUserCommandResult, error) {
	result := &ForceDeleteUserCommandResult{}

	err := h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		keys, err := repos.Profile.ListAvatarKeysByUserId(ctx, cmd.UserID)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		result.AvatarKeys = keys

		if err := repos.Alias.DeleteByUserId(ctx, cmd.UserID); err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if err := repos.User.DeleteByUserId(ctx, cmd.UserID); err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if err := repos.Profile.ForceDeleteByUserId(ctx, cmd.UserID); err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		return nil
	})
	if err != nil {
		log.Printf("[admin.force_delete_user] user_id=%s action=force_delete outcome=error err=%v", cmd.UserID, err)
		return nil, err
	}
	log.Printf("[admin.force_delete_user] user_id=%s action=force_delete outcome=success avatar_keys=%d", cmd.UserID, len(result.AvatarKeys))
	return result, nil
}
