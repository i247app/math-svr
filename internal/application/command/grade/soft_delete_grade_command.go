package command

import (
	"context"

	"math-ai.com/math-ai/internal/application/transaction"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/enum"
)

// SoftDeleteGradeCommand soft-deletes the parent row AND every translation
// row in one transaction. Soft-deleted translations stay queryable for
// audit / history paths but are filtered out of every standard read.
type SoftDeleteGradeCommand struct {
	GradeID int64
}

type SoftDeleteGradeCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewSoftDeleteGradeCommandHandler(uow transaction.UnitOfWork) *SoftDeleteGradeCommandHandler {
	return &SoftDeleteGradeCommandHandler{uow: uow}
}

func (h *SoftDeleteGradeCommandHandler) Handle(ctx context.Context, cmd SoftDeleteGradeCommand) error {
	return h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		existing, err := repos.Grade.FindByGradeId(ctx, cmd.GradeID, enum.LanguageTypeVietnamese)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if existing == nil {
			return errs.NewError(ctx, status.GRADE_NOT_FOUND, nil,
				ErrGradeNotFound)
		}

		if err := repos.GradeTranslation.SoftDeleteByGradeId(ctx, cmd.GradeID); err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if err := repos.Grade.SoftDeleteByGradeId(ctx, cmd.GradeID); err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		return nil
	})
}
