package command

import (
	"context"
	"errors"

	"math-ai.com/math-ai/internal/application/transaction"
	"math-ai.com/math-ai/internal/domain/conversation"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/enum"
)

// AppendAssistantMessageCommand is step 2 of a chat turn, run after the LLM
// reply is available. It re-reads the conversation for the current
// message_count (the user turn already advanced it), persists the assistant
// message with that seq_no, and advances the counter again.
type AppendAssistantMessageCommand struct {
	ConversationID int64
	Content        string
}

type AppendAssistantMessageCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewAppendAssistantMessageCommandHandler(uow transaction.UnitOfWork) *AppendAssistantMessageCommandHandler {
	return &AppendAssistantMessageCommandHandler{uow: uow}
}

func (h *AppendAssistantMessageCommandHandler) Handle(ctx context.Context, cmd AppendAssistantMessageCommand) (*conversation.Message, error) {
	var saved *conversation.Message

	handler := func(ctx context.Context, repos transaction.Repositories) error {
		conv, err := repos.Conversation.FindByConversationId(ctx, cmd.ConversationID)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if conv == nil {
			return errs.NewError(ctx, status.AI_CONVERSATION_NOT_FOUND, nil,
				errors.New("conversation not found"))
		}

		msg, err := appendMessage(ctx, repos, conv.ConversationId(),
			conv.MessageCount(), enum.ConversationRoleTypeAssistant, cmd.Content, nil)
		if err != nil {
			return err
		}
		saved = msg
		return nil
	}

	if err := h.uow.Do(ctx, handler); err != nil {
		return nil, err
	}
	return saved, nil
}
