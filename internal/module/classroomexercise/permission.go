package classroomexercise

import (
	"context"
	"errors"

	classroomDomain "math-ai.com/math-ai/internal/domain/classroom"
	profileDomain "math-ai.com/math-ai/internal/domain/profile"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/enum"
)

// resolveActingProfile loads the caller's profile and confirms it
// belongs to the authenticated user. Mirrors the classroom module's
// gate so an exercise endpoint cannot act on a profile the session
// doesn't own.
func (s *Service) resolveActingProfile(ctx context.Context, profileID, sessionUserID int64) (*profileDomain.Profile, error) {
	p, err := s.profileRepo.FindByProfileId(ctx, profileID)
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	if p == nil {
		return nil, errs.NewError(ctx, status.PROFILE_NOT_FOUND, nil,
			errors.New("profile not found"))
	}
	if sessionUserID != 0 && sessionUserID != p.UserId() {
		return nil, errs.NewError(ctx, status.CLASSROOM_EXERCISE_PERMISSION_DENIED, nil,
			errors.New("profile does not belong to the authenticated user"))
	}
	return p, nil
}

func (s *Service) requireMember(ctx context.Context, classroomID, profileID int64) (*classroomDomain.Member, error) {
	m, err := s.classroomMemberRepo.FindByClassroomAndProfile(ctx, classroomID, profileID)
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	if m == nil {
		return nil, errs.NewError(ctx, status.CLASSROOM_EXERCISE_PERMISSION_DENIED, nil,
			errors.New("not a member of this classroom"))
	}
	if m.MemberStatus() == nil || *m.MemberStatus() != string(enum.ClassroomMemberStatusTypeActive) {
		return nil, errs.NewError(ctx, status.CLASSROOM_EXERCISE_PERMISSION_DENIED, nil,
			errors.New("membership is not active"))
	}
	return m, nil
}

// requireManager gates exercise mutations to OWNER or CO_TEACHER —
// students cannot create / update / delete an exercise.
func (s *Service) requireManager(ctx context.Context, classroomID, profileID int64) (*classroomDomain.Member, error) {
	m, err := s.requireMember(ctx, classroomID, profileID)
	if err != nil {
		return nil, err
	}
	switch m.MemberRole() {
	case string(enum.ClassroomMemberRoleTypeOwner),
		string(enum.ClassroomMemberRoleTypeCoTeacher):
		return m, nil
	default:
		return nil, errs.NewError(ctx, status.CLASSROOM_EXERCISE_PERMISSION_DENIED, nil,
			errors.New("manager role required"))
	}
}

// isManager returns true when the caller is OWNER or CO_TEACHER, used
// to decide whether to include right_answer in the response payload.
func isManagerRole(m *classroomDomain.Member) bool {
	if m == nil {
		return false
	}
	switch m.MemberRole() {
	case string(enum.ClassroomMemberRoleTypeOwner),
		string(enum.ClassroomMemberRoleTypeCoTeacher):
		return true
	default:
		return false
	}
}
