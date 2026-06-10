// Package classroomprogress holds the request/response shapes for the
// class-learning-progress endpoints. The shapes match the design
// recorded in docs/features/001-class-learning-progress/FEATURE-SPEC.md
// — every nullable score field rides as *float64 so empty buckets and
// students with no data can marshal as JSON null (decisions #8, #18 in
// the spec).
package classroomprogress

import (
	mtime "math-ai.com/math-ai/internal/domain/shared/time"
	"math-ai.com/math-ai/internal/shared/pagination"
)

// --- /classrooms/progress/scores-over-time ---

// ScoresOverTimeReq is the request body. ProfileID is the caller's
// acting profile (validated against the session UID upstream).
// FilterProfileID is optional — when set, the chart is narrowed to a
// single student. Bucket / Tz are validator-normalised (DAY|WEEK|MONTH;
// "+HH:MM" / "-HH:MM").
type ScoresOverTimeReq struct {
	ProfileID       int64          `json:"profile_id"`
	ClassroomID     int64          `json:"classroom_id"`
	Bucket          string         `json:"bucket"`
	FromDt          mtime.MathTime `json:"from_dt"`
	ToDt            mtime.MathTime `json:"to_dt"`
	Tz              string         `json:"tz"`
	FilterProfileID *int64         `json:"filter_profile_id,omitempty"`
}

// ScoresOverTimeRes returns one point per bucket in [FromDt, ToDt].
// Empty buckets are zero-filled (decision #8) so the chart x-axis stays
// continuous client-side.
type ScoresOverTimeRes struct {
	Bucket string       `json:"bucket"`
	FromDt string       `json:"from_dt"`
	ToDt   string       `json:"to_dt"`
	Tz     string       `json:"tz"`
	Points []ChartPoint `json:"points"`
}

// ChartPoint is one bucket's aggregate. The 0–100 *_pct fields ride
// alongside the 10-point fields (decision #13) so clients can pick the
// precision they want. All score fields are pointers — nil = empty
// bucket.
type ChartPoint struct {
	BucketLabel     string   `json:"bucket_label"`
	BucketStart     string   `json:"bucket_start"`
	SubmissionCount int64    `json:"submission_count"`
	AvgScore        *float64 `json:"avg_score"`
	HighestScore    *float64 `json:"highest_score"`
	LowestScore     *float64 `json:"lowest_score"`
	AvgScorePct     *float64 `json:"avg_score_pct"`
	HighestScorePct *float64 `json:"highest_score_pct"`
	LowestScorePct  *float64 `json:"lowest_score_pct"`
}

// --- /classrooms/progress/students ---

// StudentsProgressReq drives the per-student summary endpoint. Bucket
// affects the per-student trend_series density (decision #14) but NOT
// the slope/comment math, which always runs over the last 5 calendar
// days (decision #1).
type StudentsProgressReq struct {
	ProfileID       int64          `json:"profile_id"`
	ClassroomID     int64          `json:"classroom_id"`
	Bucket          string         `json:"bucket"`
	FromDt          mtime.MathTime `json:"from_dt"`
	ToDt            mtime.MathTime `json:"to_dt"`
	Tz              string         `json:"tz"`
	FilterProfileID *int64         `json:"filter_profile_id,omitempty"`
	Page            int64          `json:"page"`
	Size            int64          `json:"size"`
}

// StudentsProgressRes carries the three summary cards + the paginated
// student list. Summary is always class-level (denominator =
// ACTIVE STUDENT-role members only — decision #7); the Students slice
// respects FilterProfileID.
//
// CommentSource tells the client which of the two per-student series
// (trend_series_by_day vs trend_series_by_exercise) was used to derive
// the Comment chip. Default is "LAST_5_DAYS" — matches signed-off
// decision #1 and the screenshot footer. Exposed so a future flip to
// per-attempt classification doesn't require a wire-shape change.
type StudentsProgressRes struct {
	Bucket        string                 `json:"bucket"`
	FromDt        string                 `json:"from_dt"`
	ToDt          string                 `json:"to_dt"`
	Tz            string                 `json:"tz"`
	CommentSource string                 `json:"comment_source"`
	Summary       ProgressSummary        `json:"summary"`
	Students      []StudentProgressRow   `json:"students"`
	Pagination    *pagination.Pagination `json:"pagination"`
}

// ProgressSummary backs the three coloured cards on the screenshot:
// total student count, count of improving students (decisions #5:
// PROGRESS+GOOD_PROGRESS), count of need-support students (decision
// #6: NEED_TO_TRY only). The *_delta_pct fields compare against the
// immediately-prior same-length period (decision #4); they're nil
// when the prior period had zero matching students (avoid div-by-zero).
type ProgressSummary struct {
	TotalStudents       int64    `json:"total_students"`
	ParticipatingCount  int64    `json:"participating_count"`
	ParticipationRate   float64  `json:"participation_rate"`
	ImprovingCount      int64    `json:"improving_count"`
	ImprovingDeltaPct   *float64 `json:"improving_delta_pct"`
	NeedSupportCount    int64    `json:"need_support_count"`
	NeedSupportDeltaPct *float64 `json:"need_support_delta_pct"`
}

// StudentProgressRow is one row of the per-student table. AvatarKey is
// the raw S3 key from ma_profiles; AvatarURL is the presigned URL the
// module layer fills in after the query runs (mirrors the existing
// classroom owner-summary hydration). Comment is the
// enum.ProgressComment value as a string — see FEATURE-SPEC §4 for the
// rule table.
//
// Per the 2026-06-10 refinement the row carries two parallel sparkline
// series:
//
//   - TrendSeriesByDay      → fixed 5-calendar-day window, daily averages
//   - TrendSeriesByExercise → last 5 graded submissions, one point each
//
// The chart's `bucket` request field no longer affects either series —
// the day window is always daily, the exercise window is always
// per-attempt. `Slope` is driven by whichever source matches the
// response-level `comment_source`.
type StudentProgressRow struct {
	ProfileID             int64                `json:"profile_id"`
	ProfileName           string               `json:"profile_name"`
	AvatarKey             *string              `json:"avatar_key,omitempty"`
	AvatarURL             *string              `json:"avatar_url,omitempty"`
	SubmissionCount       int64                `json:"submission_count"`
	AvgScore              *float64             `json:"avg_score"`
	HighestScore          *float64             `json:"highest_score"`
	LowestScore           *float64             `json:"lowest_score"`
	AvgScorePct           *float64             `json:"avg_score_pct"`
	HighestScorePct       *float64             `json:"highest_score_pct"`
	LowestScorePct        *float64             `json:"lowest_score_pct"`
	FirstSubmittedDt      string               `json:"first_submitted_dt"`
	LastSubmittedDt       string               `json:"last_submitted_dt"`
	TrendSeriesByDay      []TrendPoint         `json:"trend_series_by_day"`
	TrendSeriesByExercise []ExerciseTrendPoint `json:"trend_series_by_exercise"`
	Slope                 *float64             `json:"slope"`
	Comment               string               `json:"comment"`
}

// TrendPoint is one daily-bucket sparkline point. The 2026-06-10
// refinement fixes this series at daily density (the chart's `bucket`
// request field no longer flows through).
type TrendPoint struct {
	BucketLabel string   `json:"bucket_label"`
	AvgScore    *float64 `json:"avg_score"`
	AvgScorePct *float64 `json:"avg_score_pct"`
}

// ExerciseTrendPoint is one per-attempt sparkline point: a single
// submission's score, tagged with the parent exercise's title and id.
// BucketLabel carries the exercise title (looked up via
// exercise.IRepository.ListByClassroomExerciseIds) — when the title
// can't be resolved, it falls back to "#<exercise_id>". ExerciseID is
// the stable join key the client should use for dedup / linking.
type ExerciseTrendPoint struct {
	BucketLabel string   `json:"bucket_label"`
	ExerciseID  int64    `json:"exercise_id"`
	AvgScore    *float64 `json:"avg_score"`
	AvgScorePct *float64 `json:"avg_score_pct"`
}
