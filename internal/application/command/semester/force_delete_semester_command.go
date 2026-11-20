package command

import (
	"context"

	"math-ai.com/math-ai/internal/application/transaction"
	"math-ai.com/math-ai/internal/domain/semester"
	"math-ai.com/math-ai/internal/domain/shared/status"
	errs "math-ai.com/math-ai/internal/shared/errs"
)

// ForceDeleteSemesterCommand physically removes the semester row.
type ForceDeleteSemesterCommand struct {
	SemesterID int64
}

type ForceDeleteSemesterCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewForceDeleteSemesterCommandHandler(uow transaction.UnitOfWork) *ForceDeleteSemesterCommandHandler {
	return &ForceDeleteSemesterCommandHandler{uow: uow}
}

func (h *ForceDeleteSemesterCommandHandler) Handle(ctx context.Context, cmd ForceDeleteSemesterCommand) error {
	return h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		existing, err := repos.Semester.FindBySemesterId(ctx, cmd.SemesterID)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if existing == nil {
			return errs.NewError(ctx, semester.SemesterNotFound, nil,
				ErrSemesterNotFound)
		}

		if err := repos.Semester.ForceDeleteBySemesterId(ctx, cmd.SemesterID); err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		return nil
	})
}
