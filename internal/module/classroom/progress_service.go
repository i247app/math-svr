package classroom

import (
	"context"

	"math-ai.com/math-ai/internal/adapter/storage"
	dto "math-ai.com/math-ai/internal/application/dto/classroomprogress"
	progressQuery "math-ai.com/math-ai/internal/application/query/classroomprogress"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/infrastructure/logger"
)

// requireProgressAccess implements the tri-state permission rule for
// the class-learning-progress endpoints (FEATURE-SPEC §5):
//
//   - filter_profile_id == callerProfileID → any ACTIVE member is fine
//     (the self-view exception from decision #15).
//   - otherwise → caller must be OWNER or CO_TEACHER (requireManager).
//
// Both branches funnel through the existing permission.go helpers, so
// the rejection codes (`CLASSROOM_PERMISSION_DENIED`,
// `CLASSROOM_NOT_FOUND` propagated by requireMember/requireManager)
// match the rest of the classroom module.
func (s *Service) requireProgressAccess(
	ctx context.Context,
	classroomID, callerProfileID int64,
	filterProfileID *int64,
) error {
	if filterProfileID != nil && *filterProfileID == callerProfileID {
		_, err := s.requireMember(ctx, classroomID, callerProfileID)
		return err
	}
	_, err := s.requireManager(ctx, classroomID, callerProfileID)
	return err
}

// GetScoresOverTime powers POST /classrooms/progress/scores-over-time.
// Flow mirrors the other teacher-side reads in this module:
// validate → resolve acting profile → permission gate → load
// classroom (404 if gone) → delegate to the application query →
// return DTO. No avatar hydration on this endpoint (chart points
// don't carry per-student avatars).
func (s *Service) GetScoresOverTime(
	ctx context.Context,
	req *dto.ScoresOverTimeReq,
	sessionUserID int64,
) (*dto.ScoresOverTimeRes, error) {
	if err := ValidateScoresOverTime(ctx, req); err != nil {
		return nil, err
	}
	caller, err := s.resolveActingProfile(ctx, req.ProfileID, sessionUserID)
	if err != nil {
		return nil, err
	}
	if err := s.requireProgressAccess(ctx, req.ClassroomID, caller.ProfileId(), req.FilterProfileID); err != nil {
		return nil, err
	}
	cr, err := s.classroomRepo.FindByClassroomId(ctx, req.ClassroomID)
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	if cr == nil {
		return nil, errs.NewError(ctx, status.CLASSROOM_NOT_FOUND, nil, ErrClassroomNotFound)
	}

	res, err := s.scoresOverTimeQuery.Handle(ctx, progressQuery.ScoresOverTimeQuery{
		ClassroomID:     req.ClassroomID,
		Bucket:          req.Bucket,
		From:            req.FromDt,
		To:              req.ToDt,
		TzOffset:        req.Tz,
		FilterProfileID: req.FilterProfileID,
	})
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	return res, nil
}

// GetStudentsProgress powers POST /classrooms/progress/students.
// Same flow as GetScoresOverTime, plus avatar URL hydration on the
// way out — the query layer returns raw avatar_key strings and the
// module signs them via the storage adapter (matches the existing
// classroom owner-summary hydration shape).
func (s *Service) GetStudentsProgress(
	ctx context.Context,
	req *dto.StudentsProgressReq,
	sessionUserID int64,
) (*dto.StudentsProgressRes, error) {
	if err := ValidateStudentsProgress(ctx, req); err != nil {
		return nil, err
	}
	caller, err := s.resolveActingProfile(ctx, req.ProfileID, sessionUserID)
	if err != nil {
		return nil, err
	}
	if err := s.requireProgressAccess(ctx, req.ClassroomID, caller.ProfileId(), req.FilterProfileID); err != nil {
		return nil, err
	}
	cr, err := s.classroomRepo.FindByClassroomId(ctx, req.ClassroomID)
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	if cr == nil {
		return nil, errs.NewError(ctx, status.CLASSROOM_NOT_FOUND, nil, ErrClassroomNotFound)
	}

	res, err := s.studentsProgressQuery.Handle(ctx, progressQuery.StudentsProgressQuery{
		ClassroomID:     req.ClassroomID,
		Bucket:          req.Bucket,
		From:            req.FromDt,
		To:              req.ToDt,
		TzOffset:        req.Tz,
		FilterProfileID: req.FilterProfileID,
		Page:            req.Page,
		Size:            req.Size,
	})
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}

	s.hydrateProgressAvatars(ctx, res.Students)
	return res, nil
}

// hydrateProgressAvatars mutates the student rows in place, filling
// AvatarURL via the storage adapter when an AvatarKey is present. No-op
// for rows without a key (the always-include-NO_DATA students from
// decision #18) and when the storage adapter is disabled
// (BOT_PROVIDER="" / "disabled"-style deploys). Failures are logged
// at WARN and swallowed — a missing avatar must not 500 the response.
//
// TTL matches coverUrlTTL (1h), keeping the URL lifetimes uniform
// across the classroom surface.
func (s *Service) hydrateProgressAvatars(ctx context.Context, students []dto.StudentProgressRow) {
	if s.storageProvider == nil {
		return
	}
	for i := range students {
		key := students[i].AvatarKey
		if key == nil || *key == "" {
			continue
		}
		url, err := s.storageProvider.CreatePresignedUrl(ctx, &storage.CreatePresignedUrlRequest{
			Key:        *key,
			Expiration: coverUrlTTL,
		})
		if err != nil {
			logger.From(ctx).Warnf("classroom.progress avatar presign failed profile_id=%d err=%v",
				students[i].ProfileID, err)
			continue
		}
		students[i].AvatarURL = &url
	}
}
