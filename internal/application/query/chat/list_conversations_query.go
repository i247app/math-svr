// Package chat holds the read side of messaging. Queries take repositories
// directly and never run inside a transaction.
package chat

import (
	"context"

	domain "math-ai.com/math-ai/internal/domain/chat"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/pagination"
)

type ListConversationsQueryHandler struct {
	repo            domain.IRepository
	participantRepo domain.IParticipantRepository
}

func NewListConversationsQueryHandler(repo domain.IRepository, participantRepo domain.IParticipantRepository) *ListConversationsQueryHandler {
	return &ListConversationsQueryHandler{repo: repo, participantRepo: participantRepo}
}

// ListConversationsResult pairs the page of threads with the caller's own
// participant row for each, since unread count and read watermark are private
// to the reader and do not live on the conversation.
type ListConversationsResult struct {
	Conversations []*domain.Conversation
	Participants  map[int64]*domain.Participant
	Pagination    *pagination.Pagination
}

func (h *ListConversationsQueryHandler) Handle(ctx context.Context, params *domain.ListConversationsParams) (*ListConversationsResult, error) {
	conversations, pg, err := h.repo.ListByProfileId(ctx, params)
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}

	// One batched lookup for the whole page rather than one per row.
	ids := make([]int64, 0, len(conversations))
	for _, c := range conversations {
		ids = append(ids, c.ConversationId())
	}
	participants, err := h.participantRepo.ListByProfileAndConversationIds(ctx, params.ProfileId, ids)
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}

	return &ListConversationsResult{
		Conversations: conversations,
		Participants:  participants,
		Pagination:    pg,
	}, nil
}
