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
	// FindLatestActiveByProfileAndPurpose returns the most recently active
	// conversation owned by a profile for a given purpose (e.g. the
	// per-profile QUIZ_TUTORING thread), or (nil, nil) when none exists.
	// Used by feature modules that keep a single long-lived thread per
	// subject rather than per chat session.
	FindLatestActiveByProfileAndPurpose(ctx context.Context, profileId int64, purpose string) (*Conversation, error)
	// FindLatestActiveByUserAndPurpose returns the most recently active
	// conversation owned by a user for a given purpose (e.g. the per-user
	// CHAT session primed by /ai/shake), or (nil, nil) when none exists.
	FindLatestActiveByUserAndPurpose(ctx context.Context, userId int64, purpose string) (*Conversation, error)
	ListByUserId(ctx context.Context, params *ListConversationsParams) ([]*Conversation, *pagination.Pagination, error)
	Create(ctx context.Context, c *Conversation) (*Conversation, error)
	IncMessageCount(ctx context.Context, conversationId int64, delta int64) error
	SoftDeleteByConversationId(ctx context.Context, conversationId int64) error
}
