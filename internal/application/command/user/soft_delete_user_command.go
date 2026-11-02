package command

import (
	"context"

	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"

	"math-ai.com/math-ai/internal/application/transaction"
	"math-ai.com/math-ai/internal/infrastructure/logger"
)

type SoftDeleteUserCommand struct {
	UID int64
}

type SoftDeleteUserCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewSoftDeleteUserCommandHandler(uow transaction.UnitOfWork) *SoftDeleteUserCommandHandler {
	return &SoftDeleteUserCommandHandler{uow: uow}
}

// Handle soft-deletes the user and every row keyed by their uid in one
// transaction: user, aliases, profiles, then the rest (UserOwnedData).
func (h *SoftDeleteUserCommandHandler) Handle(ctx context.Context, cmd SoftDeleteUserCommand) error {
	err := h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		// if err := repos.User.SoftDeleteByUid(ctx, cmd.UID); err != nil {
		// 	return errs.NewError(ctx, status.FAIL, nil, err)
		// }
		// if err := repos.Alias.SoftDeleteByUid(ctx, cmd.UID); err != nil {
		// 	return errs.NewError(ctx, status.FAIL, nil, err)
		// }
		// if err := repos.Profile.SoftDeleteByUid(ctx, cmd.UID); err != nil {
		// 	return errs.NewError(ctx, status.FAIL, nil, err)
		// }
		if err := repos.UserOwnedData.SoftDeleteByUid(ctx, cmd.UID); err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		return nil
	})
	if err != nil {
		logger.From(ctx).Warnf("admin.soft_delete_user uid=%d outcome=error err=%v", cmd.UID, err)
		return err
	}
	logger.From(ctx).Infof("admin.soft_delete_user uid=%d outcome=success", cmd.UID)
	return nil
}
