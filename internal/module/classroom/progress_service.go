package classroom

import (
	"context"

	dto "math-ai.com/math-ai/internal/application/dto/classroomprogress"
	progressQuery "math-ai.com/math-ai/internal/application/query/classroomprogress"
	profileDomain "math-ai.com/math-ai/internal/domain/profile"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
)

// requireProfileProgressAccess enforces FEATURE-SPEC §5: a caller may
// view their own progress as long as they are an ACTIVE member, but may
// only view another profile's progress when they manage the classroom
// (OWNER or CO_TEACHER).
func (s *Service) requireProfileProgressAccess(
	ctx context.Context, classroomID int64, caller *profileDomain.Profile, target int64,
) error {
	if target == caller.ProfileId() {
		_, err := s.requireMember(ctx, classroomID, caller.ProfileId())
		return err
	}
	_, err := s.requireManager(ctx, classroomID, caller.ProfileId())
	return err
}

// GetProfileProgress powers POST /classrooms/progress/profile — the
// single-student progress detail (chart series + summary cards).
func (s *Service) GetProfileProgress(ctx context.Context, req dto.ProfileProgressReq, sessionUserID int64) (*dto.ProfileProgressRes, error) {
	if err := s.ValidateProfileProgress(ctx, &req); err != nil {
		return nil, err
	}

	caller, err := s.resolveActingProfile(ctx, req.ProfileID, sessionUserID)
	if err != nil {
		return nil, err
	}

	var target int64
	if req.TargetProfileID != nil && *req.TargetProfileID != 0 {
		target = *req.TargetProfileID
	}

	if err := s.requireProfileProgressAccess(ctx, req.ClassroomID, caller, target); err != nil {
		return nil, err
	}

	classroom, err := s.classroomRepo.FindByClassroomId(ctx, req.ClassroomID)
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	if classroom == nil {
		return nil, errs.NewError(ctx, status.CLASSROOM_NOT_FOUND, nil, ErrClassroomNotFound)
	}

	result, err := s.profileProgressQuery.Handle(ctx, progressQuery.ProfileProgressQuery{
		ClassroomID:     req.ClassroomID,
		TargetProfileID: target,
		From:            req.FromDt,
		To:              req.ToDt,
		Purpose:         req.Purpose,
	})
	if err != nil {
		return nil, err
	}

	return &dto.ProfileProgressRes{
		TargetProfileID: target,
		FromDt:          req.FromDt,
		ToDt:            req.ToDt,
		Tz:              req.Tz,
		Purpose:         req.Purpose,
		Series:          result.Series,
		Summary:         result.Summary,
	}, nil
}
