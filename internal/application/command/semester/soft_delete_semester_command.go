package command

import (
	"context"

	"math-ai.com/math-ai/internal/application/transaction"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/enum"
)

// SoftDeleteSemesterCommand soft-deletes the parent row AND every
// translation row in one transaction. Soft-deleted translations stay
// queryable for audit / history paths but are filtered out of every
// standard read.
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
		existing, err := repos.Semester.FindBySemesterId(ctx, cmd.SemesterID, enum.LanguageTypeVietnamese)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if existing == nil {
			return errs.NewError(ctx, status.SEMESTER_NOT_FOUND, nil,
				ErrSemesterNotFound)
		}

		if err := repos.SemesterTranslation.SoftDeleteBySemesterId(ctx, cmd.SemesterID); err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if err := repos.Semester.SoftDeleteBySemesterId(ctx, cmd.SemesterID); err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		return nil
	})
}
