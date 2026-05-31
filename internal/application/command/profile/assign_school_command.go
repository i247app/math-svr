package command

import (
	"context"
	"errors"

	"math-ai.com/math-ai/internal/application/transaction"
	"math-ai.com/math-ai/internal/domain/profile"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
)

// AssignSchoolCommand sets profile.school_id to the supplied schoolId.
// Verifies the school exists and is not soft-deleted; verifies the
// profile exists. Both checks run inside the same UoW so a race that
// soft-deletes either row mid-flight fails the whole assignment.
type AssignSchoolCommand struct {
	ProfileID int64
	SchoolID  int64
}

type AssignSchoolCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewAssignSchoolCommandHandler(uow transaction.UnitOfWork) *AssignSchoolCommandHandler {
	return &AssignSchoolCommandHandler{uow: uow}
}

func (h *AssignSchoolCommandHandler) Handle(ctx context.Context, cmd AssignSchoolCommand) (*profile.Profile, error) {
	var updated *profile.Profile

	err := h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		existing, err := repos.Profile.FindByProfileId(ctx, cmd.ProfileID)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if existing == nil {
			return errs.NewError(ctx, status.PROFILE_NOT_FOUND, nil,
				errors.New("profile not found"))
		}

		school, err := repos.School.FindBySchoolId(ctx, cmd.SchoolID)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if school == nil {
			return errs.NewError(ctx, status.SCHOOL_NOT_FOUND, nil,
				errors.New("school not found"))
		}

		schoolID := cmd.SchoolID
		if err := repos.Profile.SetSchoolId(ctx, cmd.ProfileID, &schoolID); err != nil {
			return errs.NewError(ctx, status.SCHOOL_PROFILE_LINK_FAILED, nil, err)
		}

		refreshed, err := repos.Profile.FindByProfileId(ctx, cmd.ProfileID)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		updated = refreshed
		return nil
	})
	if err != nil {
		return nil, err
	}
	return updated, nil
}
