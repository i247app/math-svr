package chat

import "errors"

var (
	ErrProfileNotFound       = errors.New("acting profile not found")
	ErrProfileNotOwnedByUser = errors.New("acting profile does not belong to the session user")
	ErrNotClassroomMember    = errors.New("acting profile is not an active member of the classroom")
	ErrTargetNotMember       = errors.New("target profile is not an active member of the classroom")
	ErrTargetProfileNotFound = errors.New("target profile not found")
	ErrCannotMessageSelf     = errors.New("cannot open a conversation with yourself")
	ErrConversationNotFound  = errors.New("conversation not found")
)
