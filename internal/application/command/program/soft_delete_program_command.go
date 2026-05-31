package command

import (
	"context"
	"errors"

	"math-ai.com/math-ai/internal/application/transaction"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/enum"
)

// SoftDeleteProgramCommand soft-deletes the parent row AND every
// translation row in one transaction. Soft-deleted translations stay
// queryable for audit / history paths but are filtered out of every
// standard read.
type SoftDeleteProgramCommand struct {
	ProgramID int64
}

type SoftDeleteProgramCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewSoftDeleteProgramCommandHandler(uow transaction.UnitOfWork) *SoftDeleteProgramCommandHandler {
	return &SoftDeleteProgramCommandHandler{uow: uow}
}

func (h *SoftDeleteProgramCommandHandler) Handle(ctx context.Context, cmd SoftDeleteProgramCommand) error {
	return h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		existing, err := repos.Program.FindByProgramId(ctx, cmd.ProgramID, enum.LanguageTypeVietnamese)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if existing == nil {
			return errs.NewError(ctx, status.PROGRAM_NOT_FOUND, nil,
				errors.New("program not found"))
		}

		if err := repos.ProgramTranslation.SoftDeleteByProgramId(ctx, cmd.ProgramID); err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if err := repos.Program.SoftDeleteByProgramId(ctx, cmd.ProgramID); err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		return nil
	})
}
