package command

import (
	"strings"

	"math-ai.com/math-ai/internal/shared/enum"
)

// DeriveProfileStatus implements the project's profile_status business rule:
//
//   - TEACHER → OFFICIAL only when BOTH id_type and teacher_id are present
//     (non-empty after trimming).
//   - STUDENT → OFFICIAL only when student_id is present.
//   - Anything else → INCOMPLETE.
//
// Inputs are taken as plain strings so the caller can pre-merge a patch
// over an existing row (see Update) before applying the rule. Pass "" for
// any field that is absent or NULL.
func DeriveProfileStatus(role, idType, teacherId, studentId string) enum.ProfileStatusType {
	switch role {
	case enum.RoleTypeTeacher.String():
		if strings.TrimSpace(idType) != "" && strings.TrimSpace(teacherId) != "" {
			return enum.ProfileStatusTypeOfficial
		}
	case enum.RoleTypeStudent.String():
		if strings.TrimSpace(studentId) != "" {
			return enum.ProfileStatusTypeOfficial
		}
	}
	return enum.ProfileStatusTypeIncomplete
}

// derefOrEmpty returns the pointee or "" when the pointer is nil. Used by
// the update flow to merge a *string patch over an existing column value.
func derefOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
