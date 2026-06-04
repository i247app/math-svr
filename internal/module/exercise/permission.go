package exercise

import (
	"context"

	classroomDomain "math-ai.com/math-ai/internal/domain/classroom"
	exerciseDomain "math-ai.com/math-ai/internal/domain/exercise"
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
			ErrProfileNotFound)
	}
	if sessionUserID != 0 && sessionUserID != p.UserId() {
		return nil, errs.NewError(ctx, status.CLASSROOM_EXERCISE_PERMISSION_DENIED, nil,
			ErrProfileNotOwnedByUser)
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
			ErrNotClassroomMember)
	}
	if m.MemberStatus() == nil || *m.MemberStatus() != string(enum.ClassroomMemberStatusTypeActive) {
		return nil, errs.NewError(ctx, status.CLASSROOM_EXERCISE_PERMISSION_DENIED, nil,
			ErrMembershipNotActive)
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
			ErrManagerRoleRequired)
	}
}

// requirePrivateAccess enforces the visibility gate. PUBLIC exercises
// pass for any caller that already cleared the upstream member/manager
// check; PRIVATE exercises pass only when the caller is the creator —
// not even other classroom managers can touch them. Called after
// requireMember / requireManager so the upstream membership errors
// surface first.
func requirePrivateAccess(ctx context.Context, e *exerciseDomain.Exercise, callerProfileID int64) error {
	if e == nil {
		return nil
	}
	if e.Visibility() != string(enum.ClassroomExerciseVisibilityPrivate) {
		return nil
	}
	if e.CreatorProfileId() == callerProfileID {
		return nil
	}
	return errs.NewError(ctx, status.CLASSROOM_EXERCISE_PRIVATE_DENIED, nil,
		ErrPrivateExerciseOwnerOnly)
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
