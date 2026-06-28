package conversation

import (
	"context"

	"math-ai.com/math-ai/internal/shared/pagination"
)

// ListConversationsParams narrows the per-user conversation listing.
// Results are scoped to UserID (ownership isolation) and ordered
// most-recent-first; Page/Limit drive pagination.
type ListConversationsParams struct {
	UserID int64
	Page   int64
	Limit  int64
}

// IRepository owns ma_ai_conversations persistence. All reads exclude
// soft-deleted/inactive rows. IncMessageCount advances the counter inside
// the same tx as a message insert so the derived seq_no stays race-free.
type IRepository interface {
	FindByConversationId(ctx context.Context, conversationId int64) (*Conversation, error)
	ListByUserId(ctx context.Context, params *ListConversationsParams) ([]*Conversation, *pagination.Pagination, error)
	Create(ctx context.Context, c *Conversation) (*Conversation, error)
	IncMessageCount(ctx context.Context, conversationId int64, delta int64) error
	SoftDeleteByConversationId(ctx context.Context, conversationId int64) error
}
