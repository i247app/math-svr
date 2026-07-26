package enum

type RoleType string

const (
	RoleTypeStudent RoleType = "STUDENT"
	RoleTypeTeacher RoleType = "TEACHER"
	RoleTypeParent  RoleType = "PARENT"
)

func (s RoleType) String() string {
	return string(s)
}

func (s RoleType) Default() RoleType {
	return RoleTypeStudent
}

func (s RoleType) IsValid() bool {
	switch s {
	case RoleTypeStudent, RoleTypeTeacher, RoleTypeParent:
		return true
	default:
		return false
	}
}

func ListRoles() []string {
	return []string{
		RoleTypeStudent.String(),
		RoleTypeTeacher.String(),
		RoleTypeParent.String(),
	}
}
