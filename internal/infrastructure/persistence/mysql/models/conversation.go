package models

import "time"

// ConversationModel mirrors ma_ai_conversations. Pure data; the repository
// maps it to the domain Conversation via ModelToDomainConversation.
type ConversationModel struct {
	Id                 int64
	ConversationId     int64
	UserId             int64
	ProfileId          *int64
	Title              *string
	Purpose            *string
	MessageCount       int64
	Note               *string
	ConversationStatus *string
	Status             string
	CreateId           *int64
	CreateDt           time.Time
	ModifyId           *int64
	ModifyDt           time.Time
}

// MessageModel mirrors ma_ai_messages.
type MessageModel struct {
	Id             int64
	MessageId      int64
	ConversationId int64
	Role           string
	Content        string
	SeqNo          int64
	Note           *string
	Status         string
	CreateId       *int64
	CreateDt       time.Time
	ModifyId       *int64
	ModifyDt       time.Time
}
