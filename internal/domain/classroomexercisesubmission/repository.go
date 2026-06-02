package classroomexercisesubmission

import (
	"context"

	mtime "math-ai.com/math-ai/internal/domain/shared/time"
	"math-ai.com/math-ai/internal/shared/pagination"
)

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

// IRepository owns ma_classroom_exercise_submissions persistence. The
// shape mirrors classroomexercise.IRepository: one Create taking a
// fully-built domain entity, a focused grading patch, and SoftDelete.
type IRepository interface {
	FindBySubmissionId(ctx context.Context, submissionId int64) (*Submission, error)
	FindByExerciseAndProfile(ctx context.Context, classroomExerciseId, profileId int64) (*Submission, error)
	ListSubmissions(ctx context.Context, params ListSubmissionsParams) ([]*Submission, *pagination.Pagination, error)
	// ListSubmittedExerciseIdsByProfile returns the set of
	// classroom_exercise_id values for which the given profile already
	// has a non-DELETED submission, intersected with the supplied
	// exerciseIds. One IN-query per call — used by hydration paths that
	// stamp a "submission_status" hint on a page of exercises without
	// going N+1 against this table.
	ListSubmittedExerciseIdsByProfile(ctx context.Context, profileId int64, exerciseIds []int64) (map[int64]struct{}, error)
	Create(ctx context.Context, sub *Submission) (*Submission, error)
	UpdateGrading(ctx context.Context, submissionId int64, patch GradingPatch) error
	SoftDelete(ctx context.Context, submissionId int64, actorID *int64) error
}
