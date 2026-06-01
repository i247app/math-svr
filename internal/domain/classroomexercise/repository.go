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
	ListExercises(ctx context.Context, params ListExercisesParams) ([]*Exercise, *pagination.Pagination, error)
	Create(ctx context.Context, e *Exercise) (*Exercise, error)
	Update(ctx context.Context, classroomExerciseId int64, patch UpdatePatch) error
	SoftDelete(ctx context.Context, classroomExerciseId int64, actorID *int64) error
}
