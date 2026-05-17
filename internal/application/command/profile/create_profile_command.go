package command

import (
	"context"
	"time"

	"github.com/google/uuid"
	"math-ai.com/math-ai/internal/application/transaction"
	"math-ai.com/math-ai/internal/domain/profile"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	mtime "math-ai.com/math-ai/internal/domain/shared/time"
	"math-ai.com/math-ai/internal/shared/utils"
)

type CreateProfileCommand struct {
	UserID     uuid.UUID
	Name       string
	Dob        *time.Time
	ProgramID  uuid.UUID
	GradeID    uuid.UUID
	SemesterID uuid.UUID
	Note       *string
}

type CreateProfileCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewCreateProfileCommandHandler(uow transaction.UnitOfWork) *CreateProfileCommandHandler {
	return &CreateProfileCommandHandler{uow: uow}
}

func (h *CreateProfileCommandHandler) Handle(ctx context.Context, cmd CreateProfileCommand) (*profile.Profile, error) {
	var created *profile.Profile

	handler := func(ctx context.Context, repos transaction.Repositories) error {
		p := profile.NewProfile()
		p.SetProfileId(utils.GenerateUUID())
		p.SetUserId(cmd.UserID)
		p.SetName(cmd.Name)
		p.SetProgramId(cmd.ProgramID)
		p.SetGradeId(cmd.GradeID)
		p.SetSemesterId(cmd.SemesterID)
		p.SetNote(cmd.Note)
		if cmd.Dob != nil {
			p.SetDob(mtime.MathTime{Time: *cmd.Dob})
		}

		saved, err := repos.Profile.Create(ctx, p)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		created = saved
		return nil
	}

	if err := h.uow.Do(ctx, handler); err != nil {
		return nil, err
	}
	return created, nil
}
