package classroomexercisesubmission

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
// belongs to the authenticated user. Mirrors the classroom + exercise
// modules' identical gate so a submission endpoint cannot act on a
// profile the session doesn't own.
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
		return nil, errs.NewError(ctx, status.CLASSROOM_EXERCISE_SUBMISSION_PERMISSION_DENIED, nil,
			errors.New("profile does not belong to the authenticated user"))
	}
	return p, nil
}

// resolveCaller resolves the acting profile with a fallback: when the
// request omits ProfileID and the session user has exactly one profile,
// that profile is used. Otherwise the caller must disambiguate.
func (s *Service) resolveCaller(ctx context.Context, profileID *int64, sessionUserID int64) (*profileDomain.Profile, error) {
	if profileID != nil && *profileID != 0 {
		return s.resolveActingProfile(ctx, *profileID, sessionUserID)
	}
	if sessionUserID == 0 {
		return nil, errs.NewError(ctx, status.CLASSROOM_EXERCISE_SUBMISSION_PERMISSION_DENIED, nil,
			errors.New("profile_id is required"))
	}
	profiles, err := s.profileRepo.ListByUserId(ctx, sessionUserID)
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	if len(profiles) == 1 {
		return profiles[0], nil
	}
	return nil, errs.NewError(ctx, status.CLASSROOM_EXERCISE_SUBMISSION_PERMISSION_DENIED, nil,
		errors.New("profile_id is required when the user owns multiple profiles"))
}

func (s *Service) requireMember(ctx context.Context, classroomID, profileID int64) (*classroomDomain.Member, error) {
	m, err := s.classroomMemberRepo.FindByClassroomAndProfile(ctx, classroomID, profileID)
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	if m == nil {
		return nil, errs.NewError(ctx, status.CLASSROOM_EXERCISE_SUBMISSION_PERMISSION_DENIED, nil,
			errors.New("not a member of this classroom"))
	}
	if m.MemberStatus() == nil || *m.MemberStatus() != string(enum.ClassroomMemberStatusTypeActive) {
		return nil, errs.NewError(ctx, status.CLASSROOM_EXERCISE_SUBMISSION_PERMISSION_DENIED, nil,
			errors.New("membership is not active"))
	}
	return m, nil
}

// requireManager gates teacher-side reads + the soft-delete path.
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
		return nil, errs.NewError(ctx, status.CLASSROOM_EXERCISE_SUBMISSION_PERMISSION_DENIED, nil,
			errors.New("manager role required"))
	}
}

// requireManagerForUser is the user-scoped counterpart to requireManager:
// the caller is identified only by sessionUserID (no specific profile
// has been claimed), so we walk every profile owned by the user and
// allow if ANY of them is an active OWNER / CO_TEACHER of the target
// classroom. Used by the flexible list endpoint where the caller may
// omit profile_id entirely and still need a manager gate. Returns the
// matching member row.
func (s *Service) requireManagerForUser(ctx context.Context, classroomID, sessionUserID int64) (*classroomDomain.Member, error) {
	if sessionUserID == 0 {
		return nil, errs.NewError(ctx, status.CLASSROOM_EXERCISE_SUBMISSION_PERMISSION_DENIED, nil,
			errors.New("session is required"))
	}
	profiles, err := s.profileRepo.ListByUserId(ctx, sessionUserID)
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	for _, p := range profiles {
		m, err := s.classroomMemberRepo.FindByClassroomAndProfile(ctx, classroomID, p.ProfileId())
		if err != nil {
			return nil, errs.NewError(ctx, status.FAIL, nil, err)
		}
		if m == nil || m.MemberStatus() == nil ||
			*m.MemberStatus() != string(enum.ClassroomMemberStatusTypeActive) {
			continue
		}
		if isManagerRole(m) {
			return m, nil
		}
	}
	return nil, errs.NewError(ctx, status.CLASSROOM_EXERCISE_SUBMISSION_PERMISSION_DENIED, nil,
		errors.New("manager role required"))
}

// isManagerRole returns true when the caller is OWNER or CO_TEACHER —
// used to decide whether the manager-only read path is allowed for a
// given submission.
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
