package chat

import (
	"context"
	"errors"

	"math-ai.com/math-ai/internal/domain/shared/mtime"
	"math-ai.com/math-ai/internal/shared/pagination"
)

// ErrDuplicateConversation and ErrDuplicateMessage are returned by the
// repository when a UNIQUE constraint rejects an insert. They exist so the
// command layer can react to a lost race without importing the SQL driver:
// a duplicate dm_key means someone opened the same 1-1 thread a moment
// earlier and the caller should read their row, and a duplicate client_msg_id
// means the client retried and the caller should return the original message.
var (
	ErrDuplicateConversation = errors.New("chat: conversation already exists")
	ErrDuplicateMessage      = errors.New("chat: message already exists")
)

// ListConversationsParams drives the inbox screen. ProfileId is required —
// the listing is always "threads I participate in". Search matches the
// counterpart's display name at the query layer, not here.
type ListConversationsParams struct {
	ProfileId        int64
	ConversationType *string
	// UnreadOnly narrows to threads with unread_count > 0 for this profile.
	UnreadOnly bool
	Page       int64
	Limit      int64
	TakeAll    bool
}

// IRepository owns ma_chat_conversations.
type IRepository interface {
	FindByConversationId(ctx context.Context, conversationId int64) (*Conversation, error)

	// FindByDmKey resolves an existing 1-1 thread. Callers build the key with
	// BuildDmKey; they must never format it themselves.
	FindByDmKey(ctx context.Context, dmKey string) (*Conversation, error)

	// ListByDmKeys batch-resolves 1-1 threads for a whole member list. Without
	// it the classroom member screen would issue one FindByDmKey per member.
	// Keys absent from the result simply have no thread yet.
	ListByDmKeys(ctx context.Context, dmKeys []string) (map[string]*Conversation, error)

	// ListByProfileId joins ma_chat_participants and orders by the
	// denormalised last_message_dt, so the inbox costs one query rather than
	// one per thread.
	ListByProfileId(ctx context.Context, params *ListConversationsParams) ([]*Conversation, *pagination.Pagination, error)

	Create(ctx context.Context, c *Conversation) (*Conversation, error)

	// NextSeqNo allocates the next per-conversation message sequence number.
	//
	// It MUST be called inside a transaction. The implementation increments
	// and then reads the counter, and those two statements only observe each
	// other's effect when they run on the same connection — outside a
	// transaction the pool can route them separately and two senders receive
	// the same seq_no. This is the same own-write rule as ma_seqs.
	NextSeqNo(ctx context.Context, conversationId int64) (int64, error)

	// UpdateLastMessage refreshes the denormalised preview and bumps
	// message_count. Runs in the same transaction as the message insert so the
	// inbox can never show a preview for a message that was rolled back.
	UpdateLastMessage(ctx context.Context, conversationId int64, m *Message, preview string) error

	SoftDeleteByConversationId(ctx context.Context, conversationId int64) error
}

// ListParticipantsParams narrows a participant read.
type ListParticipantsParams struct {
	ConversationId int64
	Status         *string
}

// IParticipantRepository owns ma_chat_participants — membership plus all
// per-user thread state (read watermark, unread badge, mute, pin).
type IParticipantRepository interface {
	FindByConversationAndProfile(ctx context.Context, conversationId, profileId int64) (*Participant, error)
	ListByConversationId(ctx context.Context, params *ListParticipantsParams) ([]*Participant, error)

	// ListByProfileAndConversationIds hydrates one profile's state across a
	// page of threads in a single round trip, preventing N+1 on the inbox and
	// on the classroom member list.
	ListByProfileAndConversationIds(ctx context.Context, profileId int64, conversationIds []int64) (map[int64]*Participant, error)

	Create(ctx context.Context, p *Participant) (*Participant, error)

	// IncUnreadExcept bumps unread_count for every ACTIVE participant of the
	// thread except the sender, and advances their delivery watermark. One
	// UPDATE rather than a read-modify-write per member, so concurrent sends
	// cannot lose a count.
	IncUnreadExcept(ctx context.Context, conversationId, exceptProfileId, seqNo int64) error

	// MarkRead advances the read watermark and clears the badge.
	//
	// The implementation guards with `last_read_seq_no < ?` so the watermark
	// can only move forward. Without that, a delayed or out-of-order client
	// call would drag it backwards and resurrect already-read messages as
	// unread.
	MarkRead(ctx context.Context, conversationId, profileId, seqNo int64, messageId *int64, readDt mtime.MathTime) error

	// SumUnreadByProfileId is the badge on the message tab: the total across
	// every thread the profile participates in.
	SumUnreadByProfileId(ctx context.Context, profileId int64) (int64, error)

	MarkLeft(ctx context.Context, participantId int64, leftDt mtime.MathTime) error
}

// ListMessagesParams drives thread paging. Exactly one direction should be
// set: BeforeSeqNo walks backwards through history (scrolling up), AfterSeqNo
// replays forward from what the client already holds (backfill on reconnect).
//
// ClearedBeforeSeqNo is the caller's own "clear history" watermark; the
// repository filters below it so one side clearing their copy never removes
// anything for the other participants.
type ListMessagesParams struct {
	ConversationId     int64
	BeforeSeqNo        *int64
	AfterSeqNo         *int64
	ClearedBeforeSeqNo int64
	Limit              int64
}

// IMessageRepository owns ma_chat_messages.
type IMessageRepository interface {
	FindByMessageId(ctx context.Context, messageId int64) (*Message, error)

	// FindByClientMsgId resolves a retry. A flaky mobile network makes the
	// client resend; the caller looks the id up first and returns the message
	// it already created instead of storing a duplicate.
	FindByClientMsgId(ctx context.Context, conversationId, senderProfileId int64, clientMsgId string) (*Message, error)

	// ListByConversationId pages by seq_no, never by OFFSET — an offset shifts
	// under the reader every time a new message arrives.
	ListByConversationId(ctx context.Context, params *ListMessagesParams) ([]*Message, error)

	Create(ctx context.Context, m *Message) (*Message, error)

	SoftDeleteByMessageId(ctx context.Context, messageId int64) error
}
