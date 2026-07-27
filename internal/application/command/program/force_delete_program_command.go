package command

import (
	"context"

	"math-ai.com/math-ai/internal/application/transaction"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
)

// ForceDeleteProgramCommand physically removes the program row.
type ForceDeleteProgramCommand struct {
	ProgramID int64
}

type ForceDeleteProgramCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewForceDeleteProgramCommandHandler(uow transaction.UnitOfWork) *ForceDeleteProgramCommandHandler {
	return &ForceDeleteProgramCommandHandler{uow: uow}
}

func (h *ForceDeleteProgramCommandHandler) Handle(ctx context.Context, cmd ForceDeleteProgramCommand) error {
	return h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		existing, err := repos.Program.FindByProgramId(ctx, cmd.ProgramID)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if existing == nil {
			return errs.NewError(ctx, status.PROGRAM_NOT_FOUND, nil,
				ErrProgramNotFound)
		}

		if err := repos.Program.ForceDeleteByProgramId(ctx, cmd.ProgramID); err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		return nil
	})
}
