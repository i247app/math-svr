package classroomexercise

import (
	"context"

	mtime "math-ai.com/math-ai/internal/domain/shared/time"
	"math-ai.com/math-ai/internal/shared/pagination"
)

// ListExercisesParams narrows the listing query. ClassroomID is required
// for the public endpoint (members only see their own classroom's
// exercises); Status filters on exercise_status when set.
type ListExercisesParams struct {
	ClassroomID int64
	Status      *string
	Page        int64
	Limit       int64
}

// UpdatePatch is the COALESCE-style update payload. Any nil field is
// skipped server-side so the caller can mutate metadata (title, dates,
// note) without re-supplying the AI-generated questions/answers.
//
// StartDate / EndDate are pointers to MathTime so a zero-value (clear)
// can be expressed by passing a zero MathTime, distinct from "leave
// unchanged" (nil).
type UpdatePatch struct {
	Title          *string
	ChapterName    *string
	LessonName     *string
	StartDate      *mtime.MathTime
	EndDate        *mtime.MathTime
	Note           *string
	ExerciseStatus *string
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
