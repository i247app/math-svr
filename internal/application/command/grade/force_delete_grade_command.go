package command

import (
	"context"
	"errors"

	"math-ai.com/math-ai/internal/application/transaction"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/enum"
)

// ForceDeleteGradeCommand physically removes the grade and every
// translation in one transaction. Translations go first so an FK addition
// later won't break ordering.
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
		existing, err := repos.Grade.FindByGradeId(ctx, cmd.GradeID, enum.LanguageTypeVietnamese)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if existing == nil {
			return errs.NewError(ctx, status.GRADE_NOT_FOUND, nil,
				errors.New("grade not found"))
		}

		if err := repos.GradeTranslation.ForceDeleteByGradeId(ctx, cmd.GradeID); err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if err := repos.Grade.ForceDeleteByGradeId(ctx, cmd.GradeID); err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		return nil
	})
}
