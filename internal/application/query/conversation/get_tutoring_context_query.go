package query

import (
	"context"

	"math-ai.com/math-ai/internal/domain/conversation"
)

// GetTutoringContextQuery loads the recent messages of a profile's
// long-lived tutoring thread (by purpose), most-recent capped at Limit and
// returned ascending by seq_no. Returns nil when the profile has no such
// thread yet (cold start) — callers treat that as "no prior context".
type GetTutoringContextQuery struct {
	ProfileID int64
	Purpose   string
	Limit     int64
}

type GetTutoringContextQueryHandler struct {
	convRepo conversation.IRepository
	msgRepo  conversation.IMessageRepository
}

func NewGetTutoringContextQueryHandler(convRepo conversation.IRepository, msgRepo conversation.IMessageRepository) *GetTutoringContextQueryHandler {
	return &GetTutoringContextQueryHandler{convRepo: convRepo, msgRepo: msgRepo}
}

func (h *GetTutoringContextQueryHandler) Handle(ctx context.Context, q GetTutoringContextQuery) ([]*conversation.Message, error) {
	if q.Limit <= 0 {
		return nil, nil
	}
	conv, err := h.convRepo.FindLatestActiveByProfileAndPurpose(ctx, q.ProfileID, q.Purpose)
	if err != nil {
		return nil, err
	}
	if conv == nil {
		return nil, nil
	}
	return h.msgRepo.ListRecentByConversationId(ctx, conv.ConversationId(), q.Limit)
}
