package classroom

import (
	"context"

	classroomDomain "math-ai.com/math-ai/internal/domain/classroom"
	profileDomain "math-ai.com/math-ai/internal/domain/profile"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/enum"
)

// resolveActingProfile loads the profile named by profileID and confirms
// it belongs to sessionUserID. This is the §0 Q1 enforcement point: the
// body's profile_id can only act for profiles owned by the authenticated
// user. Handlers extract the session UID and pass it through; an empty
// sessionUserID skips the ownership check (intended for tests and
// internal callers, never for HTTP).
func (s *Service) resolveActingProfile(ctx context.Context, profileID, sessionUserID int64) (*profileDomain.Profile, error) {
	p, err := s.profileRepo.FindByProfileId(ctx, profileID)
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	if p == nil {
		return nil, errs.NewError(ctx, status.PROFILE_NOT_FOUND, nil, ErrProfileNotFound)
	}
	if sessionUserID != 0 && sessionUserID != p.UserId() {
		return nil, errs.NewError(ctx, status.CLASSROOM_PERMISSION_DENIED, nil, ErrProfileNotOwnedByUser)
	}
	return p, nil
}

// requireTeacherRole gates classroom ownership to TEACHER profiles only
// (§0 Q2). PARENT/STUDENT cannot create or own a classroom.
func (s *Service) requireTeacherRole(ctx context.Context, p *profileDomain.Profile) error {
	if p.Role() != string(enum.RoleTypeTeacher) {
		return errs.NewError(ctx, status.CLASSROOM_INVALID_OWNER_ROLE, nil, ErrTeacherOwnershipOnly)
	}
	return nil
}

// requireMember asserts profileID is an ACTIVE member of classroomID.
// Returns the loaded member row so callers can branch on role.
func (s *Service) requireMember(ctx context.Context, classroomID, profileID int64) (*classroomDomain.Member, error) {
	m, err := s.classroomMemberRepo.FindByClassroomAndProfile(ctx, classroomID, profileID)
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	if m == nil {
		return nil, errs.NewError(ctx, status.CLASSROOM_PERMISSION_DENIED, nil, ErrNotClassroomMember)
	}
	if m.MemberStatus() == nil || *m.MemberStatus() != string(enum.ClassroomMemberStatusTypeActive) {
		return nil, errs.NewError(ctx, status.CLASSROOM_PERMISSION_DENIED, nil, ErrMembershipNotActive)
	}
	return m, nil
}

// requireManager asserts the caller is OWNER or CO_TEACHER (active).
// Used for invite / remove / classroom-update operations.
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
		return nil, errs.NewError(ctx, status.CLASSROOM_PERMISSION_DENIED, nil, ErrManagerRoleRequired)
	}
}

// requireOwner asserts the caller is the OWNER (active). Used for
// transfer / delete / archive paths.
func (s *Service) requireOwner(ctx context.Context, classroomID, profileID int64) (*classroomDomain.Member, error) {
	m, err := s.requireMember(ctx, classroomID, profileID)
	if err != nil {
		return nil, err
	}
	if m.MemberRole() != string(enum.ClassroomMemberRoleTypeOwner) {
		return nil, errs.NewError(ctx, status.CLASSROOM_PERMISSION_DENIED, nil, ErrOwnerRoleRequired)
	}
	return m, nil
}

// guardClassroomMutable rejects writes against ARCHIVED rows. DELETED
// is already hidden by the activeWhere clause so the FindByXxx lookup
// returns nil; this guard catches the ARCHIVED case where the row is
// still surfaced to the caller for read.
func guardClassroomMutable(ctx context.Context, c *classroomDomain.Classroom) error {
	if c == nil {
		return errs.NewError(ctx, status.CLASSROOM_NOT_FOUND, nil, ErrClassroomNotFound)
	}
	if c.ClassroomStatus() != nil && *c.ClassroomStatus() == string(enum.ClassroomStatusTypeArchived) {
		return errs.NewError(ctx, status.CLASSROOM_ALREADY_ARCHIVED, nil, ErrClassroomArchived)
	}
	return nil
}
