package command

import (
	"context"
	"errors"

	"math-ai.com/math-ai/internal/application/transaction"
	"math-ai.com/math-ai/internal/domain/profile"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	mtime "math-ai.com/math-ai/internal/domain/shared/time"
)

type UpdateProfileCommand struct {
	ProfileID  string
	Name       *string
	Dob        *mtime.MathTime
	ProgramID  *string
	GradeID    *string
	SemesterID *string
	Note       *string
	AvatarKey  *string
}

type UpdateProfileCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewUpdateProfileCommandHandler(uow transaction.UnitOfWork) *UpdateProfileCommandHandler {
	return &UpdateProfileCommandHandler{uow: uow}
}

// Handle mutates the existing profile row. Per the project rule, advancing
// a child's (grade, semester) NEVER inserts a new row.
func (h *UpdateProfileCommandHandler) Handle(ctx context.Context, cmd UpdateProfileCommand) (*profile.Profile, error) {
	var updated *profile.Profile

	handler := func(ctx context.Context, repos transaction.Repositories) error {
		existing, err := repos.Profile.FindByProfileId(ctx, cmd.ProfileID)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if existing == nil {
			return errs.NewError(ctx, status.PROFILE_NOT_FOUND, nil,
				errors.New("profile not found"))
		}

		patch := BuildUpdateProfile(cmd)

		if err := repos.Profile.Update(ctx, patch); err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}

		updated, err = repos.Profile.FindByProfileId(ctx, cmd.ProfileID)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		return nil
	}

	if err := h.uow.Do(ctx, handler); err != nil {
		return nil, err
	}
	return updated, nil
}

func BuildUpdateProfile(cmd UpdateProfileCommand) *profile.Profile {
	patch := profile.NewProfile()
	patch.SetProfileId(cmd.ProfileID)
	if cmd.Name != nil {
		patch.SetName(*cmd.Name)
	}
	if cmd.Dob != nil {
		patch.SetDob(*cmd.Dob)
	}
	if cmd.ProgramID != nil {
		patch.SetProgramId(cmd.ProgramID)
	}
	if cmd.GradeID != nil {
		patch.SetGradeId(cmd.GradeID)
	}
	if cmd.SemesterID != nil {
		patch.SetSemesterId(cmd.SemesterID)
	}
	if cmd.Note != nil {
		patch.SetNote(cmd.Note)
	}
	if cmd.AvatarKey != nil && *cmd.AvatarKey != "" {
		patch.SetAvatarKey(cmd.AvatarKey)
	}
	return patch
}
