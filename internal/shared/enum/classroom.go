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

type ClassroomMemberStatusType string

const (
	ClassroomMemberStatusTypeInvited ClassroomMemberStatusType = "INVITED"
	ClassroomMemberStatusTypeActive  ClassroomMemberStatusType = "ACTIVE"
	ClassroomMemberStatusTypeLeft    ClassroomMemberStatusType = "LEFT"
	ClassroomMemberStatusTypeRemoved ClassroomMemberStatusType = "REMOVED"
	ClassroomMemberStatusTypeDeleted ClassroomMemberStatusType = "DELETED"
)

func (s ClassroomMemberStatusType) String() string {
	return string(s)
}

func (s ClassroomMemberStatusType) IsValid() bool {
	switch s {
	case ClassroomMemberStatusTypeInvited,
		ClassroomMemberStatusTypeActive,
		ClassroomMemberStatusTypeLeft,
		ClassroomMemberStatusTypeRemoved,
		ClassroomMemberStatusTypeDeleted:
		return true
	default:
		return false
	}
}

type ClassroomInvitationStatusType string

const (
	ClassroomInvitationStatusTypePending   ClassroomInvitationStatusType = "PENDING"
	ClassroomInvitationStatusTypeAccepted  ClassroomInvitationStatusType = "ACCEPTED"
	ClassroomInvitationStatusTypeRejected  ClassroomInvitationStatusType = "REJECTED"
	ClassroomInvitationStatusTypeExpired   ClassroomInvitationStatusType = "EXPIRED"
	ClassroomInvitationStatusTypeCancelled ClassroomInvitationStatusType = "CANCELLED"
	ClassroomInvitationStatusTypeRevoked   ClassroomInvitationStatusType = "REVOKED"
)

func (s ClassroomInvitationStatusType) String() string {
	return string(s)
}

func (s ClassroomInvitationStatusType) IsValid() bool {
	switch s {
	case ClassroomInvitationStatusTypePending,
		ClassroomInvitationStatusTypeAccepted,
		ClassroomInvitationStatusTypeRejected,
		ClassroomInvitationStatusTypeExpired,
		ClassroomInvitationStatusTypeCancelled,
		ClassroomInvitationStatusTypeRevoked:
		return true
	default:
		return false
	}
}

// ClassroomInviteeIdentifierType identifies how `invitee_identifier` on
// ma_classroom_invitations should be interpreted at acceptance time:
// EMAIL/PHONE rows are claimable by any caller whose alias matches;
// PROFILE_ID rows are direct targeted invites.
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
