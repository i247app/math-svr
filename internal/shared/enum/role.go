package enum

type RoleType string

const (
	RoleTypeStudent RoleType = "STUDENT"
	RoleTypeTeacher RoleType = "TEACHER"
	RoleTypeParent  RoleType = "PARENT"
	// RoleTypeAdmin is an operator account. It is never self-assigned: the
	// role fields clients send (/users/create, /users/update,
	// /profiles/update) are checked with IsSelfAssignable, which refuses it.
	RoleTypeAdmin RoleType = "ADMIN"
)

func (s RoleType) String() string {
	return string(s)
}

func (s RoleType) Default() RoleType {
	return RoleTypeStudent
}

// IsValid reports whether s is a role the system knows, ADMIN included.
// Do not use it to check a role a client sent — see IsSelfAssignable.
func (s RoleType) IsValid() bool {
	switch s {
	case RoleTypeStudent, RoleTypeTeacher, RoleTypeParent, RoleTypeAdmin:
		return true
	default:
		return false
	}
}

// IsSelfAssignable reports whether a client may put s on its own account or
// profile. ADMIN is valid but not self-assignable: accepting it from a
// request body would let anyone who can sign up make themselves an admin.
func (s RoleType) IsSelfAssignable() bool {
	return s.IsValid() && s != RoleTypeAdmin
}

// ListRoles is every valid role, ADMIN included.
func ListRoles() []string {
	return append(ListSelfAssignableRoles(), RoleTypeAdmin.String())
}

// ListSelfAssignableRoles is the set of roles a client may send.
func ListSelfAssignableRoles() []string {
	return []string{
		RoleTypeStudent.String(),
		RoleTypeTeacher.String(),
		RoleTypeParent.String(),
	}
}
