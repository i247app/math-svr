package command

import (
	"context"

	"math-ai.com/math-ai/internal/application/command/shared/seqgen"
	"math-ai.com/math-ai/internal/application/transaction"
	"math-ai.com/math-ai/internal/domain/program"
	"math-ai.com/math-ai/internal/domain/seq"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
)

// CreateProgramCommand writes the parent ma_programs row in a single
// transaction.
type CreateProgramCommand struct {
	ActorID      *int64
	Label        string
	Description  string
	ImageKey     *string
	DisplayOrder int8
	Note         *string
}

type CreateProgramCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewCreateProgramCommandHandler(uow transaction.UnitOfWork) *CreateProgramCommandHandler {
	return &CreateProgramCommandHandler{uow: uow}
}

func (h *CreateProgramCommandHandler) Handle(ctx context.Context, cmd CreateProgramCommand) (*program.Program, error) {
	var created *program.Program

	err := h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		programID, err := seqgen.Next(ctx, repos.Seq, seq.NameProgram)
		if err != nil {
			return err
		}

		p := program.NewProgram()
		p.SetProgramId(programID)
		p.SetLabel(cmd.Label)
		p.SetDescription(cmd.Description)
		p.SetImageKey(cmd.ImageKey)
		p.SetDisplayOrder(cmd.DisplayOrder)
		p.SetNote(cmd.Note)
		p.SetCreateId(cmd.ActorID)

		saved, err := repos.Program.Create(ctx, p)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if saved == nil {
			return errs.NewError(ctx, status.PROGRAM_NOT_FOUND, nil,
				ErrProgramNotFoundAfterInsert)
		}
		created = saved
		return nil
	})
	if err != nil {
		return nil, err
	}
	return created, nil
}
