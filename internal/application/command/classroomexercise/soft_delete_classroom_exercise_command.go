package command

import (
	"context"

	"math-ai.com/math-ai/internal/application/transaction"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/enum"
)

// SoftDeleteClassroomExerciseCommand transitions an exercise to DELETED
// + INACTIVE in one UoW. Idempotency: a row already in DELETED is
// rejected so the caller can distinguish a no-op from a fresh delete.
type SoftDeleteClassroomExerciseCommand struct {
	ActorID             *int64
	ClassroomExerciseID int64
}

type SoftDeleteClassroomExerciseCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewSoftDeleteClassroomExerciseCommandHandler(uow transaction.UnitOfWork) *SoftDeleteClassroomExerciseCommandHandler {
	return &SoftDeleteClassroomExerciseCommandHandler{uow: uow}
}

func (h *SoftDeleteClassroomExerciseCommandHandler) Handle(ctx context.Context, cmd SoftDeleteClassroomExerciseCommand) error {
	return h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		existing, err := repos.ClassroomExercise.FindByClassroomExerciseId(ctx, cmd.ClassroomExerciseID)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if existing == nil {
			return errs.NewError(ctx, status.CLASSROOM_EXERCISE_NOT_FOUND, nil,
				ErrExerciseNotFound)
		}
		if existing.ExerciseStatus() != nil &&
			*existing.ExerciseStatus() == string(enum.ClassroomExerciseStatusTypeDeleted) {
			return errs.NewError(ctx, status.CLASSROOM_EXERCISE_ALREADY_DELETED, nil, ErrExerciseAlreadyDeleted)
		}
		if err := repos.ClassroomExercise.SoftDelete(ctx, cmd.ClassroomExerciseID, cmd.ActorID); err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		return nil
	})
}
