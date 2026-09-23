package enum

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
