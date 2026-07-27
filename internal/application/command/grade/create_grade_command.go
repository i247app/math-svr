package command

import (
	"context"

	"math-ai.com/math-ai/internal/application/command/shared/seqgen"
	"math-ai.com/math-ai/internal/application/transaction"
	"math-ai.com/math-ai/internal/domain/grade"
	"math-ai.com/math-ai/internal/domain/seq"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
)

// CreateGradeCommand writes the parent ma_grades row in a single
// transaction.
type CreateGradeCommand struct {
	ActorID      *int64
	Label        string
	Description  string
	ImageKey     *string
	DisplayOrder int8
	Note         *string
}

type CreateGradeCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewCreateGradeCommandHandler(uow transaction.UnitOfWork) *CreateGradeCommandHandler {
	return &CreateGradeCommandHandler{uow: uow}
}

func (h *CreateGradeCommandHandler) Handle(ctx context.Context, cmd CreateGradeCommand) (*grade.Grade, error) {
	var created *grade.Grade

	err := h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		gradeID, err := seqgen.Next(ctx, repos.Seq, seq.NameGrade)
		if err != nil {
			return err
		}

		g := grade.NewGrade()
		g.SetGradeId(gradeID)
		g.SetLabel(cmd.Label)
		g.SetDescription(cmd.Description)
		g.SetImageKey(cmd.ImageKey)
		g.SetDisplayOrder(cmd.DisplayOrder)
		g.SetNote(cmd.Note)
		g.SetCreateId(cmd.ActorID)

		saved, err := repos.Grade.Create(ctx, g)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if saved == nil {
			return errs.NewError(ctx, status.GRADE_NOT_FOUND, nil,
				ErrGradeNotFoundAfterInsert)
		}
		created = saved
		return nil
	})
	if err != nil {
		return nil, err
	}
	return created, nil
}
