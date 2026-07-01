package command

import (
	"context"
	"errors"

	"math-ai.com/math-ai/internal/application/transaction"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
)

// SoftDeleteNotificationCommand marks an owned notification DELETED. Ownership
// is re-checked inside the tx so a stale/forged id cannot delete another
// user's notification.
type SoftDeleteNotificationCommand struct {
	NotificationID int64
	UserID         int64
}

type SoftDeleteNotificationCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewSoftDeleteNotificationCommandHandler(uow transaction.UnitOfWork) *SoftDeleteNotificationCommandHandler {
	return &SoftDeleteNotificationCommandHandler{uow: uow}
}

func (h *SoftDeleteNotificationCommandHandler) Handle(ctx context.Context, cmd SoftDeleteNotificationCommand) error {
	return h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		n, err := repos.Notification.FindByNotificationId(ctx, cmd.NotificationID)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if n == nil {
			return errs.NewError(ctx, status.NOTIFICATION_NOT_FOUND, nil,
				errors.New("notification not found"))
		}
		if n.UserId() != cmd.UserID {
			return errs.NewError(ctx, status.NOTIFICATION_NOT_OWNED, nil,
				errors.New("notification not owned by user"))
		}
		if err := repos.Notification.SoftDeleteByNotificationId(ctx, cmd.NotificationID); err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		return nil
	})
}
