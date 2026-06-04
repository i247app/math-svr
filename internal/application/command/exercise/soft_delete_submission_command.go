package command

import (
	"context"

	"math-ai.com/math-ai/internal/application/transaction"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/enum"
)

// SoftDeleteSubmissionCommand flips a submission to DELETED inside a
// UoW. Idempotent guard: an already-DELETED row is rejected so the
// command never silently no-ops.
type SoftDeleteSubmissionCommand struct {
	ActorID                       *int64
	ClassroomExerciseSubmissionID int64
}

type SoftDeleteSubmissionCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewSoftDeleteSubmissionCommandHandler(uow transaction.UnitOfWork) *SoftDeleteSubmissionCommandHandler {
	return &SoftDeleteSubmissionCommandHandler{uow: uow}
}

func (h *SoftDeleteSubmissionCommandHandler) Handle(ctx context.Context, cmd SoftDeleteSubmissionCommand) error {
	return h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		existing, err := repos.ExerciseSubmission.FindBySubmissionId(ctx, cmd.ClassroomExerciseSubmissionID)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if existing == nil {
			return errs.NewError(ctx, status.CLASSROOM_EXERCISE_SUBMISSION_NOT_FOUND, nil,
				ErrSubmissionNotFound)
		}
		if existing.SubmissionStatus() != nil &&
			*existing.SubmissionStatus() == string(enum.ClassroomExerciseSubmissionStatusDeleted) {
			return errs.NewError(ctx, status.CLASSROOM_EXERCISE_SUBMISSION_ALREADY_DELETED, nil,
				ErrSubmissionAlreadyDeleted)
		}
		if err := repos.ExerciseSubmission.SoftDelete(ctx, cmd.ClassroomExerciseSubmissionID, cmd.ActorID); err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		return nil
	})
}
