package command

import (
	"context"

	"math-ai.com/math-ai/internal/application/transaction"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
)

// ForceDeleteGradeCommand physically removes the grade row.
type ForceDeleteGradeCommand struct {
	GradeID int64
}

type ForceDeleteGradeCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewForceDeleteGradeCommandHandler(uow transaction.UnitOfWork) *ForceDeleteGradeCommandHandler {
	return &ForceDeleteGradeCommandHandler{uow: uow}
}

func (h *ForceDeleteGradeCommandHandler) Handle(ctx context.Context, cmd ForceDeleteGradeCommand) error {
	return h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		existing, err := repos.Grade.FindByGradeId(ctx, cmd.GradeID)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if existing == nil {
			return errs.NewError(ctx, status.GRADE_NOT_FOUND, nil,
				ErrGradeNotFound)
		}

		if err := repos.Grade.ForceDeleteByGradeId(ctx, cmd.GradeID); err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		return nil
	})
}
