package command

import (
	"context"
	"errors"

	"math-ai.com/math-ai/internal/application/transaction"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
)

// MarkReadCommand marks a single owned notification read. Ownership is
// re-checked inside the tx so a stale/forged id cannot mutate another user's
// notification.
type MarkReadCommand struct {
	NotificationID int64
	UserID         int64
}

type MarkReadCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewMarkReadCommandHandler(uow transaction.UnitOfWork) *MarkReadCommandHandler {
	return &MarkReadCommandHandler{uow: uow}
}

func (h *MarkReadCommandHandler) Handle(ctx context.Context, cmd MarkReadCommand) error {
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
		if err := repos.Notification.MarkReadByNotificationId(ctx, cmd.NotificationID); err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		return nil
	})
}
