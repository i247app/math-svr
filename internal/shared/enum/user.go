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

type UserAliasStatusType string

const (
	UserAliasStatusTypeActive   UserAliasStatusType = "ACTIVE"
	UserAliasStatusTypeInactive UserAliasStatusType = "INACTIVE"
	UserAliasStatusTypeDeleted  UserAliasStatusType = "DELETED"
)

func (s UserAliasStatusType) String() string {
	return string(s)
}

func (s UserAliasStatusType) IsValid() bool {
	switch s {
	case UserAliasStatusTypeActive, UserAliasStatusTypeInactive:
		return true
	default:
		return false
	}
}
