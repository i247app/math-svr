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
	case string(enum.RoleProfileTypeTeacher):
		return string(enum.ClassroomMemberRoleTypeCoTeacher)
	default:
		return string(enum.ClassroomMemberRoleTypeStudent)
	}
}

// roleCountDeltas maps a member role to the student/teacher counter
// deltas for a single-row mutation. delta is the sign of the operation
// (+1 for join/promote, -1 for leave/remove/demote). OWNER and
// CO_TEACHER both land in the teacher bucket — they are the two
// "manager" roles that get to act as teachers in the classroom.
func roleCountDeltas(role string, delta int64) (studentDelta, teacherDelta int64) {
	switch role {
	case string(enum.ClassroomMemberRoleTypeStudent):
		return delta, 0
	case string(enum.ClassroomMemberRoleTypeCoTeacher),
		string(enum.ClassroomMemberRoleTypeOwner),
		string(enum.ClassroomMemberRoleTypeTeacher):
		return 0, delta
	default:
		return 0, 0
	}
}
