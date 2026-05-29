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

type RoleProfileType string

const (
	RoleProfileTypeStudent RoleProfileType = "STUDENT"
	RoleProfileTypeTeacher RoleProfileType = "TEACHER"
	RoleProfileTypeParent  RoleProfileType = "PARENT"
)

func (s RoleProfileType) String() string {
	return string(s)
}

func (s RoleProfileType) Default() RoleProfileType {
	return RoleProfileTypeStudent
}

func (s RoleProfileType) IsValid() bool {
	switch s {
	case RoleProfileTypeStudent, RoleProfileTypeTeacher, RoleProfileTypeParent:
		return true
	default:
		return false
	}
}
