package classroom

// import (
// 	"context"
// 	"time"

// 	dto "math-ai.com/math-ai/internal/application/dto/classroomprogress"
// 	errs "math-ai.com/math-ai/internal/domain/shared/error"
// 	"math-ai.com/math-ai/internal/domain/shared/status"
// 	"math-ai.com/math-ai/internal/shared/enum"
// )

// // maxProgressRangeDuration caps the chart/student window at ~2 years
// // (decision #12). Using 2*365 days rather than a leap-aware AddDate
// // keeps the check trivial and bounds the index scan; users won't
// // notice the ±1-day slack.
// const maxProgressRangeDuration = 2 * 365 * 24 * time.Hour

// // validateProgressCommon enforces the shared shape of both progress
// // endpoints: profile_id present, classroom_id present, bucket
// // recognised, date range valid + capped, tz recognised (or
// // defaulted). The request struct is mutated in place — tz is
// // normalised to DefaultProgressTz when empty so downstream handlers
// // always see a non-empty offset.
// func validateProgressCommon(
// 	ctx context.Context,
// 	profileID, classroomID int64,
// 	bucket string,
// 	fromDt, toDt time.Time,
// 	tz *string,
// ) error {
// 	if profileID == 0 {
// 		return errs.NewError(ctx, status.CLASSROOM_MISSING_OWNER_PROFILE_ID, nil,
// 			ErrProfileIDRequired)
// 	}
// 	if classroomID == 0 {
// 		return errs.NewError(ctx, status.CLASSROOM_MISSING_ID, nil,
// 			ErrClassroomIDRequired)
// 	}
// 	if !enum.ProgressBucket(bucket).IsValid() {
// 		return errs.NewError(ctx, status.CLASSROOM_PROGRESS_INVALID_BUCKET, nil,
// 			ErrProgressInvalidBucket)
// 	}
// 	if fromDt.IsZero() || toDt.IsZero() || !fromDt.Before(toDt) {
// 		return errs.NewError(ctx, status.CLASSROOM_PROGRESS_INVALID_DATE_RANGE, nil,
// 			ErrProgressInvalidDateRange)
// 	}
// 	if toDt.Sub(fromDt) > maxProgressRangeDuration {
// 		return errs.NewError(ctx, status.CLASSROOM_PROGRESS_INVALID_DATE_RANGE, nil,
// 			ErrProgressRangeTooLong)
// 	}
// 	if *tz == "" {
// 		*tz = enum.DefaultProgressTz
// 	} else if !enum.IsValidTzOffset(*tz) {
// 		return errs.NewError(ctx, status.CLASSROOM_PROGRESS_INVALID_TZ, nil,
// 			ErrProgressInvalidTz)
// 	}
// 	return nil
// }

// // ValidateScoresOverTime gates POST /classrooms/progress/scores-over-time.
// // All four req-side fields the handler needs are checked here; the
// // service layer then runs auth + repo calls without re-validating
// // shape.
// func ValidateScoresOverTime(ctx context.Context, req *dto.ScoresOverTimeReq) error {
// 	return validateProgressCommon(
// 		ctx,
// 		req.ProfileID, req.ClassroomID,
// 		req.Bucket,
// 		req.FromDt.Time, req.ToDt.Time,
// 		&req.Tz,
// 	)
// }

// // ValidateStudentsProgress gates POST /classrooms/progress/students.
// // Shares the common shape check + adds a page/size clamp. The clamp is
// // silent (no error) because pagination.NewPagination also clamps, and
// // surfacing a "page must be ≥ 1" error for a paging miss is more
// // friction than the API needs.
// func ValidateStudentsProgress(ctx context.Context, req *dto.StudentsProgressReq) error {
// 	if err := validateProgressCommon(
// 		ctx,
// 		req.ProfileID, req.ClassroomID,
// 		req.Bucket,
// 		req.FromDt.Time, req.ToDt.Time,
// 		&req.Tz,
// 	); err != nil {
// 		return err
// 	}
// 	if req.Page < 1 {
// 		req.Page = 1
// 	}
// 	// Size = 0 lets pagination.NewPagination fall through to its
// 	// default of 20 — leave it untouched so the contract stays in one
// 	// place.
// 	return nil
// }
