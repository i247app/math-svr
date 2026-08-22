package chat

import "errors"

var (
	// ErrNotParticipant is the base error behind CHAT_NOT_PARTICIPANT. It is
	// returned both when the caller has no participant row and when that row
	// is no longer ACTIVE — the client is told the same thing either way, so a
	// probe cannot distinguish "no such thread" from "a thread you left".
	ErrNotParticipant = errors.New("profile is not an active participant of the conversation")

	// ErrTargetNotInClassroom backs CHAT_TARGET_NOT_IN_CLASSROOM.
	ErrTargetNotInClassroom = errors.New("target profile is not an active member of the classroom")
)
