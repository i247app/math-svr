package conversation

import (
	domain "math-ai.com/math-ai/internal/domain/conversation"
	"math-ai.com/math-ai/internal/shared/pagination"
)

// SendMessageReq is the per-turn request. ConversationID is nil to start a
// new conversation. UserID is injected from the session by the handler and
// is never read from the request body.
type SendMessageReq struct {
	ConversationID *int64 `json:"conversation_id"`
	Message        string `json:"message"`
	UserID         *int64 `json:"-"`
}

// SendMessageRes returns the assistant reply plus the conversation id the
// client must echo on the next turn.
type SendMessageRes struct {
	ConversationID int64           `json:"conversation_id"`
	Reply          string          `json:"reply"`
	Message        MessageResponse `json:"message"`
}

// ListConversationsReq lists the caller's conversations. UserID is
// session-injected.
type ListConversationsReq struct {
	Page   int    `json:"page"`
	Size   int    `json:"size"`
	UserID *int64 `json:"-"`
}

type ListConversationsRes struct {
	Conversations []ConversationResponse `json:"conversations"`
	Pagination    *pagination.Pagination `json:"pagination"`
}

// GetConversationByIdReq carries the path id + session uid for ownership.
type GetConversationByIdReq struct {
	ConversationID int64  `json:"conversation_id"`
	UserID         *int64 `json:"-"`
}

type GetConversationByIdRes struct {
	Conversation ConversationResponse `json:"conversation"`
	Messages     []MessageResponse    `json:"messages"`
}

// DeleteConversationReq soft-deletes an owned conversation.
type DeleteConversationReq struct {
	ConversationID int64  `json:"conversation_id"`
	UserID         *int64 `json:"-"`
}

type DeleteConversationRes struct{}

// ConversationResponse is the wire shape for a conversation row (no message
// bodies). Title/Purpose are nil-omitted (unused in v1).
type ConversationResponse struct {
	ConversationID int64   `json:"conversation_id"`
	Title          *string `json:"title,omitempty"`
	Purpose        *string `json:"purpose,omitempty"`
	MessageCount   int64   `json:"message_count"`
	CreateDt       string  `json:"create_dt"`
	ModifyDt       string  `json:"modify_dt"`
}

// MessageResponse is the wire shape for one message turn.
type MessageResponse struct {
	MessageID int64  `json:"message_id"`
	Role      string `json:"role"`
	Content   string `json:"content"`
	SeqNo     int64  `json:"seq_no"`
	CreateDt  string `json:"create_dt"`
}

func ConversationDomainToResponse(c *domain.Conversation) ConversationResponse {
	return ConversationResponse{
		ConversationID: c.ConversationId(),
		Title:          c.Title(),
		Purpose:        c.Purpose(),
		MessageCount:   c.MessageCount(),
		CreateDt:       c.CreateDt().String(),
		ModifyDt:       c.ModifyDt().String(),
	}
}

func ConversationListToResponse(cs []*domain.Conversation) []ConversationResponse {
	out := make([]ConversationResponse, 0, len(cs))
	for _, c := range cs {
		out = append(out, ConversationDomainToResponse(c))
	}
	return out
}

func MessageDomainToResponse(m *domain.Message) MessageResponse {
	return MessageResponse{
		MessageID: m.MessageId(),
		Role:      m.Role(),
		Content:   m.Content(),
		SeqNo:     m.SeqNo(),
		CreateDt:  m.CreateDt().String(),
	}
}

func MessageListToResponse(ms []*domain.Message) []MessageResponse {
	out := make([]MessageResponse, 0, len(ms))
	for _, m := range ms {
		out = append(out, MessageDomainToResponse(m))
	}
	return out
}
