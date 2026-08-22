package chat

import (
	"context"

	domain "math-ai.com/math-ai/internal/domain/chat"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/enum"
)

type ListMessagesQuery struct {
	ConversationID int64
	ProfileID      int64
	BeforeSeqNo    *int64
	AfterSeqNo     *int64
	Limit          int64
}

type ListMessagesQueryHandler struct {
	messageRepo     domain.IMessageRepository
	participantRepo domain.IParticipantRepository
}

func NewListMessagesQueryHandler(messageRepo domain.IMessageRepository, participantRepo domain.IParticipantRepository) *ListMessagesQueryHandler {
	return &ListMessagesQueryHandler{messageRepo: messageRepo, participantRepo: participantRepo}
}

// Handle reads a page of a thread.
//
// The participant lookup is the authorization check as well as a data read:
// there is no other gate between an arbitrary conversation_id in the request
// body and this thread's history, so it must run on every call — not only when
// the client happens to send a cursor.
func (h *ListMessagesQueryHandler) Handle(ctx context.Context, q *ListMessagesQuery) ([]*domain.Message, error) {
	participant, err := h.participantRepo.FindByConversationAndProfile(ctx, q.ConversationID, q.ProfileID)
	if err != nil {
		return nil, errs.NewError(ctx, status.CHAT_NOT_PARTICIPANT, nil, err)
	}
	if participant == nil || participant.ParticipantStatus() == nil ||
		*participant.ParticipantStatus() != string(enum.ChatParticipantStatusActive) {
		return nil, errs.NewError(ctx, status.CHAT_NOT_PARTICIPANT, nil, ErrNotParticipant)
	}

	messages, err := h.messageRepo.ListByConversationId(ctx, &domain.ListMessagesParams{
		ConversationId: q.ConversationID,
		BeforeSeqNo:    q.BeforeSeqNo,
		AfterSeqNo:     q.AfterSeqNo,
		// The caller's own "clear history" watermark, so one side clearing
		// their copy hides nothing from the other participant.
		ClearedBeforeSeqNo: participant.ClearedBeforeSeqNo(),
		Limit:              q.Limit,
	})
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	return messages, nil
}
