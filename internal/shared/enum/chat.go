package enum

// ChatConversationType discriminates the shape of a thread
// (ma_chat_conversations.conversation_type).
//
// DIRECT is a 1-1 thread and is the only type written today. GROUP and
// CLASSROOM exist so that adding group chat later is inserting rows rather
// than migrating a table — a 1-1 chat is simply a conversation with exactly
// two participants, so all three share the same read, write, unread and
// pagination logic.
type ChatConversationType string

const (
	ChatConversationTypeDirect    ChatConversationType = "DIRECT"
	ChatConversationTypeGroup     ChatConversationType = "GROUP"
	ChatConversationTypeClassroom ChatConversationType = "CLASSROOM"
)

func (t ChatConversationType) String() string { return string(t) }

func (t ChatConversationType) IsValid() bool {
	switch t {
	case ChatConversationTypeDirect, ChatConversationTypeGroup, ChatConversationTypeClassroom:
		return true
	default:
		return false
	}
}

// ChatConversationStatusType is the thread's business lifecycle
// (ma_chat_conversations.conversation_status).
type ChatConversationStatusType string

const (
	ChatConversationStatusActive   ChatConversationStatusType = "ACTIVE"
	ChatConversationStatusArchived ChatConversationStatusType = "ARCHIVED"
	ChatConversationStatusDeleted  ChatConversationStatusType = "DELETED"
)

func (s ChatConversationStatusType) String() string { return string(s) }

func (s ChatConversationStatusType) IsValid() bool {
	switch s {
	case ChatConversationStatusActive, ChatConversationStatusArchived, ChatConversationStatusDeleted:
		return true
	default:
		return false
	}
}

// ChatParticipantRoleType is a member's role inside one thread
// (ma_chat_participants.participant_role). Every participant of a DIRECT
// thread is a MEMBER; OWNER/ADMIN only become meaningful with group chat.
type ChatParticipantRoleType string

const (
	ChatParticipantRoleOwner  ChatParticipantRoleType = "OWNER"
	ChatParticipantRoleAdmin  ChatParticipantRoleType = "ADMIN"
	ChatParticipantRoleMember ChatParticipantRoleType = "MEMBER"
)

func (r ChatParticipantRoleType) String() string { return string(r) }

func (r ChatParticipantRoleType) IsValid() bool {
	switch r {
	case ChatParticipantRoleOwner, ChatParticipantRoleAdmin, ChatParticipantRoleMember:
		return true
	default:
		return false
	}
}

// ChatParticipantStatusType is the membership lifecycle of one participant
// row (ma_chat_participants.participant_status).
type ChatParticipantStatusType string

const (
	ChatParticipantStatusActive  ChatParticipantStatusType = "ACTIVE"
	ChatParticipantStatusLeft    ChatParticipantStatusType = "LEFT"
	ChatParticipantStatusRemoved ChatParticipantStatusType = "REMOVED"
	ChatParticipantStatusDeleted ChatParticipantStatusType = "DELETED"
)

func (s ChatParticipantStatusType) String() string { return string(s) }

func (s ChatParticipantStatusType) IsValid() bool {
	switch s {
	case ChatParticipantStatusActive, ChatParticipantStatusLeft,
		ChatParticipantStatusRemoved, ChatParticipantStatusDeleted:
		return true
	default:
		return false
	}
}

// ChatMessageType classifies a message's payload
// (ma_chat_messages.message_type). Only TEXT and SYSTEM are written today;
// the media types are reserved for the attachment phase, where the files
// themselves live in ma_chat_attachments and never on the message row.
type ChatMessageType string

const (
	ChatMessageTypeText   ChatMessageType = "TEXT"
	ChatMessageTypeImage  ChatMessageType = "IMAGE"
	ChatMessageTypeVideo  ChatMessageType = "VIDEO"
	ChatMessageTypeAudio  ChatMessageType = "AUDIO"
	ChatMessageTypeFile   ChatMessageType = "FILE"
	ChatMessageTypeSystem ChatMessageType = "SYSTEM"
)

func (t ChatMessageType) String() string { return string(t) }

func (t ChatMessageType) IsValid() bool {
	switch t {
	case ChatMessageTypeText, ChatMessageTypeImage, ChatMessageTypeVideo,
		ChatMessageTypeAudio, ChatMessageTypeFile, ChatMessageTypeSystem:
		return true
	default:
		return false
	}
}

// ChatMessageStatusType is a message's lifecycle
// (ma_chat_messages.message_status). REVOKED rows are still returned by reads
// so the client can render "tin nhắn đã được thu hồi" in place; only DELETED
// rows are filtered out.
type ChatMessageStatusType string

const (
	ChatMessageStatusSent    ChatMessageStatusType = "SENT"
	ChatMessageStatusEdited  ChatMessageStatusType = "EDITED"
	ChatMessageStatusRevoked ChatMessageStatusType = "REVOKED"
	ChatMessageStatusDeleted ChatMessageStatusType = "DELETED"
)

func (s ChatMessageStatusType) String() string { return string(s) }

func (s ChatMessageStatusType) IsValid() bool {
	switch s {
	case ChatMessageStatusSent, ChatMessageStatusEdited,
		ChatMessageStatusRevoked, ChatMessageStatusDeleted:
		return true
	default:
		return false
	}
}
