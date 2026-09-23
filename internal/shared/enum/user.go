package enum

type UserStatusType string

const (
	UserStatusTypeActive   UserStatusType = "ACTIVE"
	UserStatusTypeInactive UserStatusType = "INACTIVE"
	UserStatusTypeDeleted  UserStatusType = "DELETED"
)

func (s UserStatusType) String() string {
	return string(s)
}

func (s UserStatusType) IsValid() bool {
	switch s {
	case UserStatusTypeActive, UserStatusTypeInactive:
		return true
	default:
		return false
	}
}
