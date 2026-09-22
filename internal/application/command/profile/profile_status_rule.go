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

// DeriveIdentity is the single source of both identity columns for a
// REGISTERED person's profile. profile_status and identity_code are
// locked to each other — OFFICIAL pairs with VERIFIED, anything else with
// USER — so the two can never drift apart. Nothing else may write
// VERIFIED.
//
// A guest profile never comes through here: guest creation sets GUEST
// explicitly, and a guest has no role to derive from.
func DeriveIdentity(role, idType, teacherId, studentId string) (enum.ProfileStatusType, enum.IdentityCodeType) {
	profileStatus := DeriveProfileStatus(role, idType, teacherId, studentId)
	if profileStatus == enum.ProfileStatusTypeOfficial {
		return profileStatus, enum.IdentityCodeVerified
	}
	return profileStatus, enum.IdentityCodeUser
}

// derefOrEmpty returns the pointee or "" when the pointer is nil. Used by
// the update flow to merge a *string patch over an existing column value.
func derefOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
