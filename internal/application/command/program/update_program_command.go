package command

import (
	"context"

	"math-ai.com/math-ai/internal/application/transaction"
	"math-ai.com/math-ai/internal/domain/program"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
)

// UpdateProgramCommand patches the parent row. Each pointer field is
// "leave unchanged when nil"; only the non-nil ones reach the repo's
// COALESCE update.
type UpdateProgramCommand struct {
	ActorID      *int64
	ProgramID    int64
	Label        *string
	Description  *string
	ImageKey     *string
	DisplayOrder *int8
	Note         *string
}

type UpdateProgramCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewUpdateProgramCommandHandler(uow transaction.UnitOfWork) *UpdateProgramCommandHandler {
	return &UpdateProgramCommandHandler{uow: uow}
}

func (h *UpdateProgramCommandHandler) Handle(ctx context.Context, cmd UpdateProgramCommand) (*program.Program, error) {
	var updated *program.Program

	err := h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		existing, err := repos.Program.FindByProgramId(ctx, cmd.ProgramID)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if existing == nil {
			return errs.NewError(ctx, status.PROGRAM_NOT_FOUND, nil,
				ErrProgramNotFound)
		}

		patched := program.NewProgram()
		patched.SetProgramId(existing.ProgramId())
		// display_order is int8 — zero is a legal value; pass through the
		// existing one unless the caller explicitly set a new value.
		if cmd.DisplayOrder != nil {
			patched.SetDisplayOrder(*cmd.DisplayOrder)
		} else {
			patched.SetDisplayOrder(existing.DisplayOrder())
		}
		if cmd.Label != nil {
			patched.SetLabel(*cmd.Label)
		}
		if cmd.Description != nil {
			patched.SetDescription(*cmd.Description)
		}
		if cmd.ImageKey != nil {
			patched.SetImageKey(cmd.ImageKey)
		}
		if cmd.Note != nil {
			patched.SetNote(cmd.Note)
		}
		patched.SetModifyId(cmd.ActorID)

		if err := repos.Program.Update(ctx, patched); err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}

		refreshed, err := repos.Program.FindByProgramId(ctx, cmd.ProgramID)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if refreshed == nil {
			return errs.NewError(ctx, status.PROGRAM_NOT_FOUND, nil,
				ErrProgramNotFoundAfterUpdate)
		}
		updated = refreshed
		return nil
	})
	if err != nil {
		return nil, err
	}
	return updated, nil
}
