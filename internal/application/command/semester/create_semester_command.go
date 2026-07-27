package command

import (
	"context"

	"math-ai.com/math-ai/internal/application/command/shared/seqgen"
	"math-ai.com/math-ai/internal/application/transaction"
	"math-ai.com/math-ai/internal/domain/semester"
	"math-ai.com/math-ai/internal/domain/seq"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
)

// CreateSemesterCommand writes the parent ma_semesters row in a single
// transaction.
type CreateSemesterCommand struct {
	ActorID      *int64
	Name         string
	Description  string
	ImageKey     *string
	DisplayOrder int8
	Note         *string
}

type CreateSemesterCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewCreateSemesterCommandHandler(uow transaction.UnitOfWork) *CreateSemesterCommandHandler {
	return &CreateSemesterCommandHandler{uow: uow}
}

func (h *CreateSemesterCommandHandler) Handle(ctx context.Context, cmd CreateSemesterCommand) (*semester.Semester, error) {
	var created *semester.Semester

	err := h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		semesterID, err := seqgen.Next(ctx, repos.Seq, seq.NameSemester)
		if err != nil {
			return err
		}

		s := semester.NewSemester()
		s.SetSemesterId(semesterID)
		s.SetName(cmd.Name)
		s.SetDescription(cmd.Description)
		s.SetImageKey(cmd.ImageKey)
		s.SetDisplayOrder(cmd.DisplayOrder)
		s.SetNote(cmd.Note)
		s.SetCreateId(cmd.ActorID)

		saved, err := repos.Semester.Create(ctx, s)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if saved == nil {
			return errs.NewError(ctx, status.SEMESTER_NOT_FOUND, nil,
				ErrSemesterNotFoundAfterInsert)
		}
		created = saved
		return nil
	})
	if err != nil {
		return nil, err
	}
	return created, nil
}
