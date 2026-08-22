// Package chat holds the request/response shapes for the messaging endpoints
// and the mappers from domain entities onto them.
package chat

import (
	domain "math-ai.com/math-ai/internal/domain/chat"
	"math-ai.com/math-ai/internal/domain/shared/mtime"
	"math-ai.com/math-ai/internal/shared/pagination"
)

// ---------- shared response shapes ----------

// ChatPeer is the display identity of the other side of a thread. It is
// assembled at the module edge from the profile row plus presence, because
// neither lives on the conversation itself.
type ChatPeer struct {
	ProfileID  int64          `json:"profile_id"`
	Name       string         `json:"name"`
	AvatarKey  *string        `json:"avatar_key"`
	AvatarURL  *string        `json:"avatar_url"`
	IsOnline   bool           `json:"is_online"`
	LastSeenDt mtime.MathTime `json:"last_seen_dt"`
}

type ConversationResponse struct {
	ConversationID   int64   `json:"conversation_id"`
	ConversationType string  `json:"conversation_type"`
	Title            *string `json:"title"`
	ParticipantCount int64   `json:"participant_count"`

	// Counterpart is populated for DIRECT threads only — a group thread has
	// no single "other person".
	Counterpart *ChatPeer `json:"counterpart"`

	LastMessageSeqNo   *int64         `json:"last_message_seq_no"`
	LastMessageType    *string        `json:"last_message_type"`
	LastMessagePreview *string        `json:"last_message_preview"`
	LastMessageDt      mtime.MathTime `json:"last_message_dt"`

	// UnreadCount and LastReadSeqNo are the caller's own state, read from
	// their participant row.
	UnreadCount   int64 `json:"unread_count"`
	LastReadSeqNo int64 `json:"last_read_seq_no"`
}

type MessageResponse struct {
	MessageID        int64          `json:"message_id"`
	ConversationID   int64          `json:"conversation_id"`
	SeqNo            int64          `json:"seq_no"`
	SenderProfileID  *int64         `json:"sender_profile_id"`
	MessageType      string         `json:"message_type"`
	Content          *string        `json:"content"`
	AttachmentCount  int64          `json:"attachment_count"`
	ReplyToMessageID *int64         `json:"reply_to_message_id"`
	ClientMsgID      *string        `json:"client_msg_id"`
	SentDt           mtime.MathTime `json:"sent_dt"`
	IsRevoked        bool           `json:"is_revoked"`
}

// ---------- classroom member list (the message tab) ----------

// ListClassroomMembersReq drives the message tab. There is no server-side
// search: a class roster is small enough that the client filters the page it
// already holds, and adding a Search filter would mean extending the shared
// classroom member repository for a requirement nobody asked for.
type ListClassroomMembersReq struct {
	ProfileID   int64 `json:"profile_id"`
	ClassroomID int64 `json:"classroom_id"`
	Page        int64 `json:"page"`
	Limit       int64 `json:"limit"`
}

// ClassroomChatMember is one row of the member list: who they are, whether
// they are online, and — if the two have talked before — the thread and its
// unread badge, so tapping a row opens straight into history.
type ClassroomChatMember struct {
	ProfileID  int64          `json:"profile_id"`
	Name       string         `json:"name"`
	MemberRole string         `json:"member_role"`
	AvatarKey  *string        `json:"avatar_key"`
	AvatarURL  *string        `json:"avatar_url"`
	IsOnline   bool           `json:"is_online"`
	LastSeenDt mtime.MathTime `json:"last_seen_dt"`

	ConversationID     *int64         `json:"conversation_id"`
	UnreadCount        int64          `json:"unread_count"`
	LastMessagePreview *string        `json:"last_message_preview"`
	LastMessageDt      mtime.MathTime `json:"last_message_dt"`
}

type ListClassroomMembersRes struct {
	Members    []*ClassroomChatMember `json:"members"`
	Pagination *pagination.Pagination `json:"pagination"`
}

// ---------- open conversation ----------

type OpenConversationReq struct {
	ProfileID int64 `json:"profile_id"`
	// ClassroomID is the permission basis, not part of the thread's identity:
	// a direct thread is global, so the same conversation is reached from any
	// classroom the two people share.
	ClassroomID     int64 `json:"classroom_id"`
	TargetProfileID int64 `json:"target_profile_id"`
}

type OpenConversationRes struct {
	Conversation *ConversationResponse `json:"conversation"`
}

// ---------- conversation list (inbox) ----------

type ListConversationsReq struct {
	ProfileID  int64 `json:"profile_id"`
	UnreadOnly bool  `json:"unread_only"`
	Page       int64 `json:"page"`
	Limit      int64 `json:"limit"`
}

type ListConversationsRes struct {
	Conversations []*ConversationResponse `json:"conversations"`
	Pagination    *pagination.Pagination  `json:"pagination"`
}

// ---------- messages ----------

// ListMessagesReq pages a thread. BeforeSeqNo walks backwards through history
// (scroll up); AfterSeqNo replays forward from what the client already holds,
// which is how a reconnecting client fills the gap the socket cannot replay.
type ListMessagesReq struct {
	ProfileID      int64  `json:"profile_id"`
	ConversationID int64  `json:"conversation_id"`
	BeforeSeqNo    *int64 `json:"before_seq_no"`
	AfterSeqNo     *int64 `json:"after_seq_no"`
	Limit          int64  `json:"limit"`
}

type ListMessagesRes struct {
	Messages []*MessageResponse `json:"messages"`
}

type SendMessageReq struct {
	ProfileID      int64  `json:"profile_id"`
	ConversationID int64  `json:"conversation_id"`
	Content        string `json:"content"`
	// ClientMsgID makes a retry idempotent on a flaky mobile network: resending
	// the same value returns the message already stored instead of a duplicate.
	ClientMsgID      *string `json:"client_msg_id"`
	ReplyToMessageID *int64  `json:"reply_to_message_id"`
}

type SendMessageRes struct {
	Message *MessageResponse `json:"message"`
}

type MarkReadReq struct {
	ProfileID      int64 `json:"profile_id"`
	ConversationID int64 `json:"conversation_id"`
	// SeqNo is the highest sequence number the client has displayed. The
	// watermark only ever moves forward.
	SeqNo int64 `json:"seq_no"`
}

type MarkReadRes struct {
	ConversationID int64 `json:"conversation_id"`
	LastReadSeqNo  int64 `json:"last_read_seq_no"`
}

type UnreadCountReq struct {
	ProfileID int64 `json:"profile_id"`
}

type UnreadCountRes struct {
	UnreadCount int64 `json:"unread_count"`
}

// ---------- mappers ----------

func DomainToConversationResponse(c *domain.Conversation) *ConversationResponse {
	if c == nil {
		return nil
	}
	return &ConversationResponse{
		ConversationID:     c.ConversationId(),
		ConversationType:   c.ConversationType(),
		Title:              c.Title(),
		ParticipantCount:   c.ParticipantCount(),
		LastMessageSeqNo:   c.LastMessageSeqNo(),
		LastMessageType:    c.LastMessageType(),
		LastMessagePreview: c.LastMessagePreview(),
		LastMessageDt:      c.LastMessageDt(),
	}
}

func DomainToMessageResponse(m *domain.Message) *MessageResponse {
	if m == nil {
		return nil
	}
	return &MessageResponse{
		MessageID:        m.MessageId(),
		ConversationID:   m.ConversationId(),
		SeqNo:            m.SeqNo(),
		SenderProfileID:  m.SenderProfileId(),
		MessageType:      m.MessageType(),
		Content:          m.Content(),
		AttachmentCount:  m.AttachmentCount(),
		ReplyToMessageID: m.ReplyToMessageId(),
		ClientMsgID:      m.ClientMsgId(),
		SentDt:           m.SentDt(),
		IsRevoked:        !m.RevokedDt().Time.IsZero(),
	}
}

func DomainListToMessageResponse(list []*domain.Message) []*MessageResponse {
	out := make([]*MessageResponse, 0, len(list))
	for _, m := range list {
		out = append(out, DomainToMessageResponse(m))
	}
	return out
}
