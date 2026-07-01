package command

import (
	"context"

	"math-ai.com/math-ai/internal/application/transaction"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
)

// MarkAllReadCommand marks every unread notification of a user read.
type MarkAllReadCommand struct {
	UserID int64
}

type MarkAllReadCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewMarkAllReadCommandHandler(uow transaction.UnitOfWork) *MarkAllReadCommandHandler {
	return &MarkAllReadCommandHandler{uow: uow}
}

func (h *MarkAllReadCommandHandler) Handle(ctx context.Context, cmd MarkAllReadCommand) error {
	return h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		if err := repos.Notification.MarkAllReadByUserId(ctx, cmd.UserID); err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		return nil
	})
}
