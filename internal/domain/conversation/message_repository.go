package conversation

import (
	"context"
)

// IMessageRepository owns ma_ai_messages persistence.
//
// ListRecentByConversationId returns at most `limit` of the most recent
// messages but ordered ASCENDING by seq_no, so the caller can append them
// to a prompt in chronological order. This is the "history window" read on
// the per-turn path; limit is supplied by the conversation config.
type IMessageRepository interface {
	Create(ctx context.Context, m *Message) (*Message, error)
	ListRecentByConversationId(ctx context.Context, conversationId int64, limit int64) ([]*Message, error)
}
