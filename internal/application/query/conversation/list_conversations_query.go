package query

import (
	"context"

	"math-ai.com/math-ai/internal/domain/conversation"
	"math-ai.com/math-ai/internal/shared/pagination"
)

// ListConversationsQuery lists the caller's conversations, most-recent
// first, scoped to UserID for tenant isolation.
type ListConversationsQuery struct {
	UserID int64
	Page   int64
	Limit  int64
}

type ListConversationsQueryHandler struct {
	convRepo conversation.IRepository
}

func NewListConversationsQueryHandler(convRepo conversation.IRepository) *ListConversationsQueryHandler {
	return &ListConversationsQueryHandler{convRepo: convRepo}
}

func (h *ListConversationsQueryHandler) Handle(ctx context.Context, q ListConversationsQuery) ([]*conversation.Conversation, *pagination.Pagination, error) {
	return h.convRepo.ListByUserId(ctx, &conversation.ListConversationsParams{
		UserID: q.UserID,
		Page:   q.Page,
		Limit:  q.Limit,
	})
}
