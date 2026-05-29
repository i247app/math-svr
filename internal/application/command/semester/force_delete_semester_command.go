package command

import (
	"context"
	"errors"

	"math-ai.com/math-ai/internal/application/transaction"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/enum"
)

// ForceDeleteSemesterCommand physically removes the semester and every
// translation in one transaction. Translations go first so an FK
// addition later won't break ordering.
type ForceDeleteSemesterCommand struct {
	SemesterID string
}

type ForceDeleteSemesterCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewForceDeleteSemesterCommandHandler(uow transaction.UnitOfWork) *ForceDeleteSemesterCommandHandler {
	return &ForceDeleteSemesterCommandHandler{uow: uow}
}

func (h *ForceDeleteSemesterCommandHandler) Handle(ctx context.Context, cmd ForceDeleteSemesterCommand) error {
	return h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		existing, err := repos.Semester.FindBySemesterId(ctx, cmd.SemesterID, enum.LanguageTypeVietnamese)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if existing == nil {
			return errs.NewError(ctx, status.SEMESTER_NOT_FOUND, nil,
				errors.New("semester not found"))
		}

		if err := repos.SemesterTranslation.ForceDeleteBySemesterId(ctx, cmd.SemesterID); err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if err := repos.Semester.ForceDeleteBySemesterId(ctx, cmd.SemesterID); err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		return nil
	})
}
