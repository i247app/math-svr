package command

import (
	"context"

	"math-ai.com/math-ai/internal/application/transaction"
	"math-ai.com/math-ai/internal/domain/profile"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	mtime "math-ai.com/math-ai/internal/domain/shared/time"
	"math-ai.com/math-ai/internal/shared/utils"
)

type CreateProfileCommand struct {
	UserID     string
	Name       string
	Dob        *mtime.MathTime
	ProgramID  *string
	GradeID    *string
	SemesterID *string
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
		p := BuildProfile(cmd)

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

func BuildProfile(cmd CreateProfileCommand) *profile.Profile {
	userUUID, err := utils.StringToUUID(cmd.UserID)
	if err != nil {
		return nil
	}

	programUUID, err := utils.PtrStringToUUID(cmd.ProgramID)
	if err != nil {
		return nil
	}
	gradeUUID, err := utils.PtrStringToUUID(cmd.GradeID)
	if err != nil {
		return nil
	}
	semesterUUID, err := utils.PtrStringToUUID(cmd.SemesterID)
	if err != nil {
		return nil
	}

	p := profile.NewProfile()
	p.SetProfileId(utils.GenerateUUID())
	p.SetUserId(userUUID)
	p.SetName(cmd.Name)
	p.SetProgramId(&programUUID)
	p.SetGradeId(&gradeUUID)
	p.SetSemesterId(&semesterUUID)
	p.SetNote(cmd.Note)
	if cmd.Dob != nil {
		p.SetDob(*cmd.Dob)
	}
	return p
}
