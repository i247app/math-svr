package command

import (
	"context"
	"errors"

	"math-ai.com/math-ai/internal/application/transaction"
	notificationdom "math-ai.com/math-ai/internal/domain/notification"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/errs"
)

// MarkReadCommand marks a single owned notification read. Ownership is
// re-checked inside the tx so a stale/forged id cannot mutate another user's
// notification.
type MarkReadCommand struct {
	NotificationID int64
	UID            int64
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
			return errs.NewError(ctx, notificationdom.NotificationNotFound, nil,
				errors.New("notification not found"))
		}
		if n.UID() != cmd.UID {
			return errs.NewError(ctx, notificationdom.NotificationNotOwned, nil,
				errors.New("notification not owned by user"))
		}
		if err := repos.Notification.MarkReadByNotificationId(ctx, cmd.NotificationID); err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		return nil
	})
}
