package command

import "math-ai.com/math-ai/internal/shared/enum"

// memberRoleForProfileRole maps an ma_profiles.role to the member_role
// a profile receives when joining a classroom via the invite-code path.
// OWNER is intentionally never produced here — there is exactly one
// OWNER per classroom (the creator) and transfer-ownership is the only
// way to move that seat. A TEACHER profile joining via code lands as
// CO_TEACHER so they can manage the room without displacing the owner;
// STUDENT and PARENT (and any unknown role) land as STUDENT.
func memberRoleForProfileRole(profileRole string) string {
	switch profileRole {
	case string(enum.RoleTypeTeacher):
		return string(enum.ClassroomMemberRoleTypeCoTeacher)
	default:
		return string(enum.ClassroomMemberRoleTypeStudent)
	}
}
