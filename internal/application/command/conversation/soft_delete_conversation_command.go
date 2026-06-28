package command

import (
	"context"
	"errors"

	"math-ai.com/math-ai/internal/application/transaction"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
)

// SoftDeleteConversationCommand marks an owned conversation DELETED. Ownership
// is re-checked inside the tx so a stale/forged id cannot delete another
// user's conversation.
type SoftDeleteConversationCommand struct {
	ConversationID int64
	UserID         int64
}

type SoftDeleteConversationCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewSoftDeleteConversationCommandHandler(uow transaction.UnitOfWork) *SoftDeleteConversationCommandHandler {
	return &SoftDeleteConversationCommandHandler{uow: uow}
}

func (h *SoftDeleteConversationCommandHandler) Handle(ctx context.Context, cmd SoftDeleteConversationCommand) error {
	return h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		conv, err := repos.Conversation.FindByConversationId(ctx, cmd.ConversationID)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if conv == nil {
			return errs.NewError(ctx, status.AI_CONVERSATION_NOT_FOUND, nil,
				errors.New("conversation not found"))
		}
		if conv.UserId() != cmd.UserID {
			return errs.NewError(ctx, status.AI_CONVERSATION_NOT_OWNED, nil,
				errors.New("conversation not owned by user"))
		}
		if err := repos.Conversation.SoftDeleteByConversationId(ctx, cmd.ConversationID); err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		return nil
	})
}
