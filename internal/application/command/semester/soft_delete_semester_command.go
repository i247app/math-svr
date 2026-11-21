package command

import (
	"context"

	"math-ai.com/math-ai/internal/application/transaction"
	"math-ai.com/math-ai/internal/domain/semester"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/errs"
)

// SoftDeleteSemesterCommand soft-deletes the parent row. The soft-deleted
// row stays in the table but is filtered out of every standard read.
type SoftDeleteSemesterCommand struct {
	SemesterID int64
}

type SoftDeleteSemesterCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewSoftDeleteSemesterCommandHandler(uow transaction.UnitOfWork) *SoftDeleteSemesterCommandHandler {
	return &SoftDeleteSemesterCommandHandler{uow: uow}
}

func (h *SoftDeleteSemesterCommandHandler) Handle(ctx context.Context, cmd SoftDeleteSemesterCommand) error {
	return h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		existing, err := repos.Semester.FindBySemesterId(ctx, cmd.SemesterID)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if existing == nil {
			return errs.NewError(ctx, semester.SemesterNotFound, nil,
				ErrSemesterNotFound)
		}

		if err := repos.Semester.SoftDeleteBySemesterId(ctx, cmd.SemesterID); err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		return nil
	})
}
