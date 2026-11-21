package command

import (
	"context"

	"math-ai.com/math-ai/internal/application/transaction"
	"math-ai.com/math-ai/internal/domain/grade"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/errs"
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
			return errs.NewError(ctx, grade.GradeNotFound, nil,
				ErrGradeNotFound)
		}

		if err := repos.Grade.ForceDeleteByGradeId(ctx, cmd.GradeID); err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		return nil
	})
}
