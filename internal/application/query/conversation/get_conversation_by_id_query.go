package query

import (
	"context"

	"math-ai.com/math-ai/internal/domain/conversation"
)

// GetConversationByIdQuery loads a conversation plus its most-recent
// messages (ascending seq_no, capped at MessageLimit). Ownership is NOT
// enforced here — the module service compares UserId() after load and
// returns AI_CONVERSATION_NOT_OWNED.
type GetConversationByIdQuery struct {
	ConversationID int64
	MessageLimit   int64
}

type GetConversationByIdResult struct {
	Conversation *conversation.Conversation
	Messages     []*conversation.Message
}

type GetConversationByIdQueryHandler struct {
	convRepo conversation.IRepository
	msgRepo  conversation.IMessageRepository
}

func NewGetConversationByIdQueryHandler(convRepo conversation.IRepository, msgRepo conversation.IMessageRepository) *GetConversationByIdQueryHandler {
	return &GetConversationByIdQueryHandler{convRepo: convRepo, msgRepo: msgRepo}
}

func (h *GetConversationByIdQueryHandler) Handle(ctx context.Context, q GetConversationByIdQuery) (*GetConversationByIdResult, error) {
	conv, err := h.convRepo.FindByConversationId(ctx, q.ConversationID)
	if err != nil {
		return nil, err
	}
	if conv == nil {
		return nil, nil
	}

	messages, err := h.msgRepo.ListRecentByConversationId(ctx, q.ConversationID, q.MessageLimit)
	if err != nil {
		return nil, err
	}
	return &GetConversationByIdResult{Conversation: conv, Messages: messages}, nil
}
