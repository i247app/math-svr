package classroom

import (
	"context"
	"strings"

	dto "math-ai.com/math-ai/internal/application/dto/classroomprogress"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	mtime "math-ai.com/math-ai/internal/domain/shared/time"
	"math-ai.com/math-ai/internal/shared/enum"
)

// maxProgressRange caps the queryable window at ~2 years to keep the
// per-student range scan bounded. Mirrors FEATURE-SPEC §2 decision #11.
const maxProgressRangeHours = 2 * 365 * 24

// ValidateProfileProgress checks the request for the single-student
// progress detail and normalises the tz default in place. Returns a
// MathError on the first failure.
func (s *Service) ValidateProfileProgress(ctx context.Context, req *dto.ProfileProgressReq) error {
	if req.ClassroomID == 0 {
		return errs.NewError(ctx, status.CLASSROOM_MISSING_ID, nil, ErrClassroomIDRequired)
	}
	if req.ProfileID == 0 {
		return errs.NewError(ctx, status.PROFILE_NOT_FOUND, nil, ErrProfileIDRequired)
	}

	if req.FromDt == "" || req.ToDt == "" {
		return errs.NewError(ctx, status.CLASSROOM_PROGRESS_INVALID_DATE_RANGE, nil, ErrProgressInvalidDateRange)
	}
	from, err := mtime.ParseFromString(req.FromDt)
	if err != nil {
		return errs.NewError(ctx, status.CLASSROOM_PROGRESS_INVALID_DATE_RANGE, nil, ErrProgressInvalidDateRange)
	}
	to, err := mtime.ParseFromString(req.ToDt)
	if err != nil {
		return errs.NewError(ctx, status.CLASSROOM_PROGRESS_INVALID_DATE_RANGE, nil, ErrProgressInvalidDateRange)
	}

	if !from.Time.Before(to.Time) {
		return errs.NewError(ctx, status.CLASSROOM_PROGRESS_INVALID_DATE_RANGE, nil, ErrProgressInvalidDateRange)
	}
	if to.Time.Sub(from.Time).Hours() > maxProgressRangeHours {
		return errs.NewError(ctx, status.CLASSROOM_PROGRESS_INVALID_DATE_RANGE, nil, ErrProgressRangeTooLong)
	}

	// tz defaults to the Vietnam offset; otherwise must be a numeric offset.
	req.Tz = strings.TrimSpace(req.Tz)
	if req.Tz == "" {
		req.Tz = enum.DefaultProgressTz
	} else if !enum.IsValidTzOffset(req.Tz) {
		return errs.NewError(ctx, status.CLASSROOM_PROGRESS_INVALID_TZ, nil, ErrProgressInvalidTz)
	}

	// purpose is optional; when present it must be a known exercise intent.
	if req.Purpose != nil {
		p := strings.ToUpper(strings.TrimSpace(*req.Purpose))
		if p == "" {
			req.Purpose = nil
		} else if !enum.ClassroomExercisePurposeType(p).IsValid() {
			return errs.NewError(ctx, status.CLASSROOM_PROGRESS_INVALID_PURPOSE, nil, ErrProgressInvalidPurpose)
		} else {
			req.Purpose = &p
		}
	}

	return nil
}
