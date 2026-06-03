package classroomexercise

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
	ModifyID       *int64
}

// IRepository owns all classroom_exercise persistence. The shape mirrors
// quiz.IRepository: a single Create that takes the fully-built domain
// object (id already minted), a patch-style Update, and a SoftDelete.
type IRepository interface {
	FindByClassroomExerciseId(ctx context.Context, id int64) (*Exercise, error)
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

// IRepository owns ma_classroom_exercise_submissions persistence. The
// shape mirrors classroomexercise.IRepository: one Create taking a
// fully-built domain entity, a focused grading patch, and SoftDelete.
type ISubmissionRepository interface {
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
	// ListByExerciseAndProfileIds fetches the submission row for each
	// (exerciseId, profileId in profileIds) pair in one round trip,
	// used by the teacher-side roster view to hydrate per-row
	// submission metadata onto the submitted-members page.
	ListByExerciseAndProfileIds(ctx context.Context, classroomExerciseId int64, profileIds []int64) ([]*Submission, error)
	Create(ctx context.Context, sub *Submission) (*Submission, error)
	UpdateGrading(ctx context.Context, submissionId int64, patch GradingPatch) error
	SoftDelete(ctx context.Context, submissionId int64, actorID *int64) error
}
