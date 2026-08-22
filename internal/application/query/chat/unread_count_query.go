package chat

import (
	"context"

	domain "math-ai.com/math-ai/internal/domain/chat"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
)

type UnreadCountQueryHandler struct {
	participantRepo domain.IParticipantRepository
}

func NewUnreadCountQueryHandler(participantRepo domain.IParticipantRepository) *UnreadCountQueryHandler {
	return &UnreadCountQueryHandler{participantRepo: participantRepo}
}

// Handle returns the badge for the message tab: the sum of unread counts
// across every thread the profile participates in.
func (h *UnreadCountQueryHandler) Handle(ctx context.Context, profileID int64) (int64, error) {
	total, err := h.participantRepo.SumUnreadByProfileId(ctx, profileID)
	if err != nil {
		return 0, errs.NewError(ctx, status.FAIL, nil, err)
	}
	return total, nil
}
