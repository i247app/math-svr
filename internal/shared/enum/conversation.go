package enum

// ConversationRoleType is the author role of a single chat message in an
// AI conversation. It maps onto the `role` column of ma_ai_messages and,
// at the adapter edge, onto the bot adapter's message roles. Named with a
// Conversation prefix to avoid colliding with RoleType (user/teacher
// roles) in this package.
type ConversationRoleType string

const (
	ConversationRoleTypeSystem    ConversationRoleType = "system"
	ConversationRoleTypeUser      ConversationRoleType = "user"
	ConversationRoleTypeAssistant ConversationRoleType = "assistant"
)

func (r ConversationRoleType) String() string {
	return string(r)
}

func (r ConversationRoleType) IsValid() bool {
	switch r {
	case ConversationRoleTypeSystem, ConversationRoleTypeUser, ConversationRoleTypeAssistant:
		return true
	default:
		return false
	}
}

// ConversationStatusType is the business lifecycle of a conversation row
// (the `conversation_status` column), mirroring the dual-status pattern
// used by classroom/school. ACTIVE is the live state; DELETED is the
// soft-delete tombstone hidden by *ActiveWhere filters.
type ConversationStatusType string

const (
	ConversationStatusTypeActive  ConversationStatusType = "ACTIVE"
	ConversationStatusTypeDeleted ConversationStatusType = "DELETED"
)

func (s ConversationStatusType) String() string {
	return string(s)
}

func (s ConversationStatusType) IsValid() bool {
	switch s {
	case ConversationStatusTypeActive, ConversationStatusTypeDeleted:
		return true
	default:
		return false
	}
}
