package command

import (
	"context"

	"math-ai.com/math-ai/internal/application/transaction"
	"math-ai.com/math-ai/internal/domain/conversation"
	"math-ai.com/math-ai/internal/domain/seq"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/enum"
)

// CreateConversationCommand creates an empty conversation thread and returns
// it. Used by the chat flow to mint a thread up-front (so the framework
// memory store has a conversation_id to append to) before the first turn.
type CreateConversationCommand struct {
	UserID    int64
	ProfileID *int64
	Purpose   string
	Title     *string
}

type CreateConversationCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewCreateConversationCommandHandler(uow transaction.UnitOfWork) *CreateConversationCommandHandler {
	return &CreateConversationCommandHandler{uow: uow}
}

func (h *CreateConversationCommandHandler) Handle(ctx context.Context, cmd CreateConversationCommand) (*conversation.Conversation, error) {
	var created *conversation.Conversation

	err := h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		id, err := nextSeqID(ctx, repos, seq.NameAiConversation)
		if err != nil {
			return err
		}

		c := conversation.NewConversation()
		c.SetConversationId(id)
		c.SetUserId(cmd.UserID)
		c.SetProfileId(cmd.ProfileID)
		if cmd.Purpose != "" {
			p := cmd.Purpose
			c.SetPurpose(&p)
		}
		if cmd.Title != nil {
			c.SetTitle(cmd.Title)
		}
		c.SetMessageCount(0)
		active := string(enum.ConversationStatusTypeActive)
		c.SetConversationStatus(&active)
		c.SetCreateId(&cmd.UserID)

		saved, err := repos.Conversation.Create(ctx, c)
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
