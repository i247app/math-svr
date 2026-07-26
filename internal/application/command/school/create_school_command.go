package command

import (
	"context"

	"math-ai.com/math-ai/internal/application/command/shared/seqgen"
	"math-ai.com/math-ai/internal/application/transaction"
	"math-ai.com/math-ai/internal/domain/school"
	"math-ai.com/math-ai/internal/domain/seq"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
)

type CreateSchoolCommand struct {
	ActorID     *int64
	Name        string
	Description *string
	ImageKey    *string
	District    *string
	Province    *string
	Note        *string
}

type CreateSchoolCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewCreateSchoolCommandHandler(uow transaction.UnitOfWork) *CreateSchoolCommandHandler {
	return &CreateSchoolCommandHandler{uow: uow}
}

func (h *CreateSchoolCommandHandler) Handle(ctx context.Context, cmd CreateSchoolCommand) (*school.School, error) {
	var created *school.School

	err := h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		schoolID, err := seqgen.Next(ctx, repos.Seq, seq.NameSchool)
		if err != nil {
			return err
		}

		s := school.NewSchool()
		s.SetSchoolId(schoolID)
		s.SetName(cmd.Name)
		s.SetDescription(cmd.Description)
		s.SetImageKey(cmd.ImageKey)
		s.SetDistrict(cmd.District)
		s.SetProvince(cmd.Province)
		s.SetNote(cmd.Note)
		s.SetCreateId(cmd.ActorID)

		saved, err := repos.School.Create(ctx, s)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if saved == nil {
			return errs.NewError(ctx, status.SCHOOL_NOT_FOUND, nil,
				ErrSchoolNotFoundAfterInsert)
		}
		created = saved
		return nil
	})
	if err != nil {
		return nil, err
	}
	return created, nil
}
