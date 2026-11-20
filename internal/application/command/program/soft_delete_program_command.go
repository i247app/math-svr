package command

import (
	"context"

	"math-ai.com/math-ai/internal/application/transaction"
	"math-ai.com/math-ai/internal/domain/program"
	"math-ai.com/math-ai/internal/domain/shared/status"
	errs "math-ai.com/math-ai/internal/shared/errs"
)

// SoftDeleteProgramCommand soft-deletes the parent row. The soft-deleted
// row stays in the table but is filtered out of every standard read.
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
		existing, err := repos.Program.FindByProgramId(ctx, cmd.ProgramID)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if existing == nil {
			return errs.NewError(ctx, program.ProgramNotFound, nil,
				ErrProgramNotFound)
		}

		if err := repos.Program.SoftDeleteByProgramId(ctx, cmd.ProgramID); err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		return nil
	})
}
