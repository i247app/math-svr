package command

import (
	"context"

	"math-ai.com/math-ai/internal/application/transaction"
	domain "math-ai.com/math-ai/internal/domain/classroomexercise"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	mtime "math-ai.com/math-ai/internal/domain/shared/time"
)

// UpdateClassroomExerciseCommand patches metadata only — title, chapter,
// lesson, dates, note. The AI-generated questions/answers columns are
// out of scope here; a regeneration would be a separate command.
//
// StartDate / EndDate use pointers to MathTime so the caller can
// distinguish "leave unchanged" (nil) from "clear the column" (a
// pointer to zero MathTime). The repo translates the latter into a
// typed SQL NULL.
type UpdateClassroomExerciseCommand struct {
	ActorID             *int64
	ClassroomExerciseID int64
	Title               *string
	Description         *string
	ChapterName         *string
	LessonName          *string
	StartDate           *mtime.MathTime
	EndDate             *mtime.MathTime
	Note                *string
	ExerciseStatus      *string
	Visibility          *string
}

type UpdateClassroomExerciseCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewUpdateClassroomExerciseCommandHandler(uow transaction.UnitOfWork) *UpdateClassroomExerciseCommandHandler {
	return &UpdateClassroomExerciseCommandHandler{uow: uow}
}

func (h *UpdateClassroomExerciseCommandHandler) Handle(ctx context.Context, cmd UpdateClassroomExerciseCommand) (*domain.Exercise, error) {
	var updated *domain.Exercise

	err := h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		existing, err := repos.ClassroomExercise.FindByClassroomExerciseId(ctx, cmd.ClassroomExerciseID)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if existing == nil {
			return errs.NewError(ctx, status.CLASSROOM_EXERCISE_NOT_FOUND, nil,
				ErrExerciseNotFound)
		}

		patch := domain.UpdatePatch{
			Title:          cmd.Title,
			Description:    cmd.Description,
			ChapterName:    cmd.ChapterName,
			LessonName:     cmd.LessonName,
			StartDate:      cmd.StartDate,
			EndDate:        cmd.EndDate,
			Note:           cmd.Note,
			ExerciseStatus: cmd.ExerciseStatus,
			Visibility:     cmd.Visibility,
			ModifyID:       cmd.ActorID,
		}
		if err := repos.ClassroomExercise.Update(ctx, cmd.ClassroomExerciseID, patch); err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}

		refreshed, err := repos.ClassroomExercise.FindByClassroomExerciseId(ctx, cmd.ClassroomExerciseID)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		updated = refreshed
		return nil
	})
	if err != nil {
		return nil, err
	}
	return updated, nil
}
