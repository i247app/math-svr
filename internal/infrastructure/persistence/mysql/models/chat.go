package models

import "time"

type ChatConversationModel struct {
	Id                         int64
	ConversationId             int64
	ConversationType           string
	ClassroomId                *int64
	DmKey                      *string
	Title                      *string
	AvatarKey                  *string
	OwnerProfileId             *int64
	ParticipantCount           int64
	LastSeqNo                  int64
	MessageCount               int64
	LastMessageId              *int64
	LastMessageSeqNo           *int64
	LastMessageType            *string
	LastMessagePreview         *string
	LastMessageSenderProfileId *int64
	LastMessageDt              *time.Time
	Note                       *string
	ConversationStatus         *string
	Status                     string
	CreateId                   *int64
	CreateDt                   time.Time
	ModifyId                   *int64
	ModifyDt                   time.Time
}

type ChatParticipantModel struct {
	Id                 int64
	ParticipantId      int64
	ConversationId     int64
	ProfileId          int64
	UserId             int64
	ParticipantRole    string
	LastReadSeqNo      int64
	LastReadMessageId  *int64
	LastReadDt         *time.Time
	LastDeliveredSeqNo int64
	UnreadCount        int64
	IsMuted            bool
	MutedUntilDt       *time.Time
	IsPinned           bool
	ClearedBeforeSeqNo int64
	JoinedDt           *time.Time
	LeftDt             *time.Time
	InvitedByProfileId *int64
	Note               *string
	ParticipantStatus  *string
	Status             string
	CreateId           *int64
	CreateDt           time.Time
	ModifyId           *int64
	ModifyDt           time.Time
}

type ChatMessageModel struct {
	Id               int64
	MessageId        int64
	ConversationId   int64
	SeqNo            int64
	SenderProfileId  *int64
	SenderUserId     *int64
	MessageType      string
	Content          *string
	AttachmentCount  int64
	ReplyToMessageId *int64
	SystemEvent      *string
	SystemPayload    *string
	Metadata         *string
	ClientMsgId      *string
	SentDt           time.Time
	EditedDt         *time.Time
	RevokedDt        *time.Time
	Note             *string
	MessageStatus    *string
	Status           string
	CreateId         *int64
	CreateDt         time.Time
	ModifyId         *int64
	ModifyDt         time.Time
}
