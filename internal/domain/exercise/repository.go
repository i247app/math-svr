package exercise

import (
	"context"

	mtime "math-ai.com/math-ai/internal/domain/shared/time"
	"math-ai.com/math-ai/internal/shared/pagination"
)

// ListExercisesParams narrows the listing query. ClassroomID is required
// for the public endpoint (members only see their own classroom's
// exercises); Status filters on exercise_status when set.
//
// CallerProfileID drives the visibility filter:
//
//	WHERE visibility = 'PUBLIC' OR creator_profile_id = CallerProfileID
//
// The service layer is expected to supply the caller's acting profile_id
// so PRIVATE rows owned by other profiles are invisible at the SQL level.
// Zero (when no caller is known) keeps PUBLIC-only behavior — a safe
// default for any code path that forgets to set it.
type ListExercisesParams struct {
	ClassroomID     int64
	CallerProfileID int64

	// Optional filters. Nil/zero values mean "skip this predicate".
	Status           *string
	Visibility       *string
	CreatorProfileID *int64
	ProgramID        *int64
	ChapterName      *string
	LessonName       *string
	Search           *string
	Purpose          *string

	// Sort hints. Both default to the repo's preferred ordering
	// (created DESC). Validator-normalized tokens only — see
	// classroomexercise.validExerciseSortBy / validExerciseSortOrder.
	SortBy    *string
	SortOrder *string

	Page  int64
	Limit int64
}

// UpdatePatch is the COALESCE-style update payload. Any nil field is
// skipped server-side so the caller can mutate metadata (title, dates,
// note) without re-supplying the AI-generated questions/answers.
//
// StartDate / EndDate are pointers to MathTime so a zero-value (clear)
// can be expressed by passing a zero MathTime, distinct from "leave
// unchanged" (nil). Visibility is patchable so a creator can flip an
// exercise between PUBLIC and PRIVATE without re-supplying the rest of
// the row; creator_profile_id is intentionally absent — ownership is
// set once at create time and never changes.
type UpdatePatch struct {
	Title          *string
	Description    *string
	ChapterName    *string
	LessonName     *string
	StartDate      *mtime.MathTime
	EndDate        *mtime.MathTime
	Note           *string
	ExerciseStatus *string
	Visibility     *string
	// Purpose flips the exercise intent (HOMEWORK / EXAM). Nil leaves
	// the column untouched; the repo's COALESCE preserves the prior
	// value. Validator must normalise + whitelist before this reaches
	// the repo since the column is NOT NULL.
	Purpose  *string
	ModifyID *int64
}

// ListByClassroomIdsParams drives the batched home/layout exercise read.
// ClassroomIDs is required (empty → no rows). The optional filters mirror
// the per-classroom ListExercises gate but operate across a set of
// classrooms in a single round trip so the home dashboard never goes N+1.
//
//	CallerProfileID  visibility gate — PUBLIC OR creator = caller. Zero
//	                 keeps PUBLIC-only (safe default).
//	CreatorProfileID exact creator filter (teacher's "exercises I assigned").
//	OnlyActive       drop ARCHIVED rows (keep exercise_status NULL/ACTIVE).
//	NotExpiredAsOf   end_date IS NULL OR end_date >= asOf — "still open".
//	Limit            hard cap on the returned rows (defaults applied by repo).
type ListByClassroomIdsParams struct {
	ClassroomIDs     []int64
	CallerProfileID  int64
	CreatorProfileID *int64
	OnlyActive       bool
	NotExpiredAsOf   *mtime.MathTime
	Limit            int64
}

// IRepository owns all classroom_exercise persistence. The shape mirrors
// quiz.IRepository: a single Create that takes the fully-built domain
// object (id already minted), a patch-style Update, and a SoftDelete.
type IRepository interface {
	FindByClassroomExerciseId(ctx context.Context, id int64) (*Exercise, error)
	// ListByClassroomIds returns active exercises across a set of
	// classrooms for the home/layout dashboard, applying the optional
	// visibility / creator / lifecycle / expiry filters in one query.
	// Ordered create_dt DESC so the most recently assigned surface first.
	ListByClassroomIds(ctx context.Context, params ListByClassroomIdsParams) ([]*Exercise, error)
	// ListByClassroomExerciseIds batches the per-id lookup for hydration
	// flows (e.g. the submission list embedding exercise summaries).
	// One round trip per page regardless of size; rows missing from the
	// result are simply absent from the returned slice — callers should
	// treat missing ids as "exercise deleted or not visible".
	ListByClassroomExerciseIds(ctx context.Context, ids []int64) ([]*Exercise, error)
	ListExercises(ctx context.Context, params ListExercisesParams) ([]*Exercise, *pagination.Pagination, error)
	Create(ctx context.Context, e *Exercise) (*Exercise, error)
	Update(ctx context.Context, classroomExerciseId int64, patch UpdatePatch) error
	SoftDelete(ctx context.Context, classroomExerciseId int64, actorID *int64) error
}

// ListSubmissionsParams narrows the listing query. The repo composes
// optional filters (each nil/zero value is skipped) on top of the
// active-where filter (excludes DELETED + system-inactive rows).
//
// SortBy / SortOrder are normalised by the module-layer validator
// against a whitelist; the repo trusts them and maps to a real column.
type ListSubmissionsParams struct {
	ClassroomID         int64
	ClassroomExerciseID int64
	ProfileID           int64

	Status *string

	SortBy    *string
	SortOrder *string

	Page  int64
	Limit int64
}

// GradingPatch holds the bot-supplied grading result. AIReview is the
// qualitative review string; the other fields are numeric scores. Any
// field left nil means "leave the existing value alone" — the repo
// applies them via COALESCE so a retry can refine partial results
// without nulling out prior progress.
type GradingPatch struct {
	AIReview        *string
	TotalQuestions  *int64
	CorrectNumber   *int64
	ScorePercentage *int64
	// GradedDt is the timestamp the row should record as the grading
	// time. A zero value leaves graded_dt untouched.
	GradedDt mtime.MathTime
	// SubmissionStatus is typically advanced to GRADED here. nil leaves
	// the row's status untouched.
	SubmissionStatus *string
	ModifyID         *int64
}

// IRepository owns ma_exercise_submissions persistence. The
// shape mirrors classroomexercise.IRepository: one Create taking a
// fully-built domain entity, a focused grading patch, and SoftDelete.
type ISubmissionRepository interface {
	FindBySubmissionId(ctx context.Context, submissionId int64) (*Submission, error)
	FindByExerciseAndProfile(ctx context.Context, classroomExerciseId, profileId int64) (*Submission, error)
	// ListRecentByProfileIds returns the most recent non-DELETED
	// submissions across the given set of profiles, ordered by
	// submitted_dt DESC. Used by the parent home/layout to surface
	// "exercises my children just completed" in one round trip. Limit
	// caps the result (repo applies a default when zero).
	ListRecentByProfileIds(ctx context.Context, profileIds []int64, limit int64) ([]*Submission, error)
	ListSubmissions(ctx context.Context, params ListSubmissionsParams) ([]*Submission, *pagination.Pagination, error)
	// ListSubmittedExerciseIdsByProfile returns the set of
	// classroom_exercise_id values for which the given profile already
	// has a non-DELETED submission, intersected with the supplied
	// exerciseIds. One IN-query per call — used by hydration paths that
	// stamp a "submission_status" hint on a page of exercises without
	// going N+1 against this table.
	ListSubmittedExerciseIdsByProfile(ctx context.Context, profileId int64, exerciseIds []int64) (map[int64]struct{}, error)
	// ListByExerciseAndProfileIds fetches the submission row for each
	// (exerciseId, profileId in profileIds) pair in one round trip,
	// used by the teacher-side roster view to hydrate per-row
	// submission metadata onto the submitted-members page.
	ListByExerciseAndProfileIds(ctx context.Context, classroomExerciseId int64, profileIds []int64) ([]*Submission, error)
	// ListBucketedScores aggregates submission scores into chart-shaped
	// rows for the /classrooms/progress/scores-over-time endpoint —
	// one row per (bucket_label) inside [From, To], with COUNT plus
	// AVG/MIN/MAX of score_percentage. The bucket label format is
	// driven by params.Bucket (DAY|WEEK|MONTH); see the SQL builder
	// in the mysql impl for the exact DATE_FORMAT/CONVERT_TZ shape.
	// Caller is responsible for zero-filling missing buckets.
	ListBucketedScores(ctx context.Context, params BucketedScoresParams) ([]*BucketedScoreRow, error)
	// ListForProgress streams the (profile_id, score_percentage,
	// submitted_dt) tuples for a classroom inside [From, To],
	// ordered by (profile_id, submitted_dt ASC). One round trip; the
	// caller groups by profile_id and runs the trend / slope /
	// comment math in-process. FilterProfileID is optional — when
	// non-nil, the result is narrowed to that profile only.
	ListForProgress(ctx context.Context, params ProgressRangeParams) ([]*ProgressRow, error)
	Create(ctx context.Context, sub *Submission) (*Submission, error)
	UpdateGrading(ctx context.Context, submissionId int64, patch GradingPatch) error
	SoftDelete(ctx context.Context, submissionId int64, actorID *int64) error
}

// BucketedScoresParams drives ListBucketedScores. Bucket is one of
// "DAY", "WEEK", "MONTH" — validator-normalised at the module layer
// against enum.ProgressBucket. TzOffset is a CONVERT_TZ-friendly
// offset string ("+07:00", "-05:30"); the mysql impl substitutes it
// verbatim into CONVERT_TZ(submitted_dt, '+00:00', ?). FilterProfileID
// narrows the aggregation to a single student when non-nil.
type BucketedScoresParams struct {
	ClassroomID     int64
	Bucket          string
	TzOffset        string
	From            mtime.MathTime
	To              mtime.MathTime
	FilterProfileID *int64
}

// BucketedScoreRow is one chart point's raw aggregation, on the DB's
// native 0–100 score_percentage scale. The application layer converts
// to the 10-point scale for the UI; both forms ride the wire (decision
// #13 in the feature spec).
type BucketedScoreRow struct {
	BucketLabel     string
	SubmissionCount int64
	AvgPct          float64
	MinPct          int64
	MaxPct          int64
}

// ProgressRangeParams drives ListForProgress. The query returns the
// minimum tuple needed to run the per-student trend math — no joins,
// no expensive columns. FilterProfileID narrows the result to a single
// student when non-nil (the §15 self-exception path).
type ProgressRangeParams struct {
	ClassroomID     int64
	From            mtime.MathTime
	To              mtime.MathTime
	FilterProfileID *int64
}

// ProgressRow is one submission's score and timestamp, scoped to a
// classroom. Ordered (profile_id ASC, submitted_dt ASC) by the repo
// so the caller can stream-group by profile_id without an extra sort.
//
// ClassroomExerciseID is included so the application layer can build
// the per-exercise trend series (trend_series_by_exercise on the
// /classrooms/progress/students endpoint) and look up the exercise
// title via exercise.IRepository.ListByClassroomExerciseIds.
type ProgressRow struct {
	ProfileID           int64
	ClassroomExerciseID int64
	ScorePercentage     int64
	SubmittedDt         mtime.MathTime
}
