package command

import "math-ai.com/math-ai/internal/shared/enum"

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
		string(enum.ClassroomMemberRoleTypeOwner):
		return 0, delta
	default:
		return 0, 0
	}
}
