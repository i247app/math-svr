package enum

type ProfileStatusType string

const (
	ProfileStatusTypeActive   ProfileStatusType = "ACTIVE"
	ProfileStatusTypeInactive ProfileStatusType = "INACTIVE"
	ProfileStatusTypeDeleted  ProfileStatusType = "DELETED"
)

func (s ProfileStatusType) String() string {
	return string(s)
}

func (s ProfileStatusType) IsValid() bool {
	switch s {
	case ProfileStatusTypeActive, ProfileStatusTypeInactive:
		return true
	default:
		return false
	}
}
