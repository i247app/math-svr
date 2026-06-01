package enum

type ClassroomStatusType string

const (
	ClassroomStatusTypeActive   ClassroomStatusType = "ACTIVE"
	ClassroomStatusTypeArchived ClassroomStatusType = "ARCHIVED"
	ClassroomStatusTypeDeleted  ClassroomStatusType = "DELETED"
)

func (s ClassroomStatusType) String() string {
	return string(s)
}

func (s ClassroomStatusType) IsValid() bool {
	switch s {
	case ClassroomStatusTypeActive, ClassroomStatusTypeArchived, ClassroomStatusTypeDeleted:
		return true
	default:
		return false
	}
}

type ClassroomMemberRoleType string

const (
	ClassroomMemberRoleTypeOwner     ClassroomMemberRoleType = "OWNER"
	ClassroomMemberRoleTypeCoTeacher ClassroomMemberRoleType = "CO_TEACHER"
	ClassroomMemberRoleTypeTeacher   ClassroomMemberRoleType = "TEACHER"
	ClassroomMemberRoleTypeStudent   ClassroomMemberRoleType = "STUDENT"
)

func (s ClassroomMemberRoleType) String() string {
	return string(s)
}

func (s ClassroomMemberRoleType) IsValid() bool {
	switch s {
	case ClassroomMemberRoleTypeOwner, ClassroomMemberRoleTypeCoTeacher, ClassroomMemberRoleTypeStudent:
		return true
	default:
		return false
	}
}

// ClassroomMemberStatusType is the source of truth for the membership
// lifecycle. Invitations live on ma_classroom_members: a freshly-sent
// invitation is a row in PENDING; acceptance flips it to ACTIVE;
// rejection / manager-cancellation / member-leave / member-removal all
// land at their respective terminal states without dropping the row,
// so the (classroom_id, profile_id) UNIQUE key stays meaningful when
// the same target is re-invited or rejoins later.
type ClassroomMemberStatusType string

const (
	// PENDING_INVITATION → teacher invited; awaiting invitee accept/reject.
	ClassroomMemberStatusTypePendingInvitation ClassroomMemberStatusType = "PENDING_INVITATION"
	// PENDING_REQUEST → user requested; awaiting owner approve/reject.
	ClassroomMemberStatusTypePendingRequest ClassroomMemberStatusType = "PENDING_REQUEST"
	ClassroomMemberStatusTypeActive         ClassroomMemberStatusType = "ACTIVE"
	ClassroomMemberStatusTypeRejected       ClassroomMemberStatusType = "REJECTED"
	ClassroomMemberStatusTypeLeft           ClassroomMemberStatusType = "LEFT"
	ClassroomMemberStatusTypeRemoved        ClassroomMemberStatusType = "REMOVED"
	ClassroomMemberStatusTypeDeleted        ClassroomMemberStatusType = "DELETED"
)

func (s ClassroomMemberStatusType) String() string {
	return string(s)
}

func (s ClassroomMemberStatusType) IsValid() bool {
	switch s {
	case ClassroomMemberStatusTypePendingInvitation,
		ClassroomMemberStatusTypePendingRequest,
		ClassroomMemberStatusTypeActive,
		ClassroomMemberStatusTypeRejected,
		ClassroomMemberStatusTypeLeft,
		ClassroomMemberStatusTypeRemoved,
		ClassroomMemberStatusTypeDeleted:
		return true
	default:
		return false
	}
}

// ClassroomRelationshipType encodes the calling profile's relation to a
// classroom on a list response. NONE means the profile has no row in
// ma_classroom_members for that classroom (or the row sits in a
// terminal state — REJECTED/LEFT/REMOVED — which the UI should treat
// the same as "not participating"). MEMBER, PENDING_INVITATION, and
// PENDING_REQUEST map 1:1 onto ClassroomMemberStatusType so the
// frontend can branch on a single field instead of inspecting raw
// member_status values.
type ClassroomRelationshipType string

const (
	ClassroomRelationshipTypeMember             ClassroomRelationshipType = "MEMBER"
	ClassroomRelationshipTypePendingInvitation  ClassroomRelationshipType = "PENDING_INVITATION"
	ClassroomRelationshipTypePendingRequest     ClassroomRelationshipType = "PENDING_REQUEST"
	ClassroomRelationshipTypeNone               ClassroomRelationshipType = "NONE"
)

func (s ClassroomRelationshipType) String() string {
	return string(s)
}

// ClassroomInviteeIdentifierType identifies how a send-invitation
// caller addressed the target — EMAIL/PHONE need alias lookup to
// resolve a profile_id, PROFILE_ID is a direct id.
type ClassroomInviteeIdentifierType string

const (
	ClassroomInviteeIdentifierTypeEmail     ClassroomInviteeIdentifierType = "EMAIL"
	ClassroomInviteeIdentifierTypePhone     ClassroomInviteeIdentifierType = "PHONE"
	ClassroomInviteeIdentifierTypeProfileId ClassroomInviteeIdentifierType = "PROFILE_ID"
)

func (s ClassroomInviteeIdentifierType) String() string {
	return string(s)
}

func (s ClassroomInviteeIdentifierType) IsValid() bool {
	switch s {
	case ClassroomInviteeIdentifierTypeEmail,
		ClassroomInviteeIdentifierTypePhone,
		ClassroomInviteeIdentifierTypeProfileId:
		return true
	default:
		return false
	}
}

// ClassroomExerciseStatusType is the lifecycle on ma_classroom_exercises.
// Mirrors ClassroomStatusType but kept separate because the legal
// transitions are independent (an exercise can be archived without
// touching the classroom).
type ClassroomExerciseStatusType string

const (
	ClassroomExerciseStatusTypeActive   ClassroomExerciseStatusType = "ACTIVE"
	ClassroomExerciseStatusTypeArchived ClassroomExerciseStatusType = "ARCHIVED"
	ClassroomExerciseStatusTypeDeleted  ClassroomExerciseStatusType = "DELETED"
)

func (s ClassroomExerciseStatusType) String() string {
	return string(s)
}

func (s ClassroomExerciseStatusType) IsValid() bool {
	switch s {
	case ClassroomExerciseStatusTypeActive,
		ClassroomExerciseStatusTypeArchived,
		ClassroomExerciseStatusTypeDeleted:
		return true
	default:
		return false
	}
}
