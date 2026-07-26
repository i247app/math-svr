package command

import (
	"context"

	"math-ai.com/math-ai/internal/application/command/shared/seqgen"
	"math-ai.com/math-ai/internal/application/transaction"
	"math-ai.com/math-ai/internal/domain/notification"
	"math-ai.com/math-ai/internal/domain/seq"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/enum"
)

// CreateNotificationCommand persists one notification row for a recipient
// (UserID). The external id is minted via ma_seqs inside the UoW. Push
// delivery is orchestrated separately by the module service (the FCM call must
// not hold a transaction open).
type CreateNotificationCommand struct {
	UserID     int64
	Title      string
	ShortText  string
	Category   *string
	ActionType *string
	ActionData *string // opaque JSON
	Priority   *string
	Note       *string
	CreatorUID *int64
}

type CreateNotificationCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewCreateNotificationCommandHandler(uow transaction.UnitOfWork) *CreateNotificationCommandHandler {
	return &CreateNotificationCommandHandler{uow: uow}
}

func (h *CreateNotificationCommandHandler) Handle(ctx context.Context, cmd CreateNotificationCommand) (*notification.Notification, error) {
	var created *notification.Notification

	err := h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		id, err := seqgen.Next(ctx, repos.Seq, seq.NameNotification)
		if err != nil {
			return err
		}

		n := notification.NewNotification()
		n.SetNotificationId(id)
		n.SetUserId(cmd.UserID)
		n.SetTitle(cmd.Title)
		n.SetShortText(cmd.ShortText)
		n.SetCategory(cmd.Category)
		n.SetIsRead(false)
		n.SetActionType(cmd.ActionType)
		n.SetActionData(cmd.ActionData)

		priority := string(enum.NotificationPriorityTypeNormal)
		if cmd.Priority != nil && *cmd.Priority != "" {
			priority = *cmd.Priority
		}
		n.SetPriority(&priority)

		n.SetNote(cmd.Note)
		active := string(enum.NotificationStatusTypeActive)
		n.SetNotificationStatus(&active)
		n.SetCreateId(cmd.CreatorUID)

		saved, err := repos.Notification.Create(ctx, n)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		created = saved
		return nil
	})
	if err != nil {
		return nil, err
	}
	return created, nil
}
