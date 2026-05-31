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
	ClassroomMemberStatusTypePending  ClassroomMemberStatusType = "PENDING"
	ClassroomMemberStatusTypeActive   ClassroomMemberStatusType = "ACTIVE"
	ClassroomMemberStatusTypeRejected ClassroomMemberStatusType = "REJECTED"
	ClassroomMemberStatusTypeLeft     ClassroomMemberStatusType = "LEFT"
	ClassroomMemberStatusTypeRemoved  ClassroomMemberStatusType = "REMOVED"
	ClassroomMemberStatusTypeDeleted  ClassroomMemberStatusType = "DELETED"
)

func (s ClassroomMemberStatusType) String() string {
	return string(s)
}

func (s ClassroomMemberStatusType) IsValid() bool {
	switch s {
	case ClassroomMemberStatusTypePending,
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
