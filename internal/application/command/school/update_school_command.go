package command

import (
	"context"
	"errors"

	"math-ai.com/math-ai/internal/application/transaction"
	"math-ai.com/math-ai/internal/domain/school"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
)

// UpdateSchoolCommand patches the parent row. Each pointer field is
// "leave unchanged when nil"; only the non-nil ones reach the repo's
// COALESCE update so partial PATCH-style payloads work.
type UpdateSchoolCommand struct {
	ActorID     *int64
	SchoolID    int64
	Name        *string
	Description *string
	ImageKey    *string
	District    *string
	Province    *string
	Note        *string
}

type UpdateSchoolCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewUpdateSchoolCommandHandler(uow transaction.UnitOfWork) *UpdateSchoolCommandHandler {
	return &UpdateSchoolCommandHandler{uow: uow}
}

func (h *UpdateSchoolCommandHandler) Handle(ctx context.Context, cmd UpdateSchoolCommand) (*school.School, error) {
	var updated *school.School

	err := h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		existing, err := repos.School.FindBySchoolId(ctx, cmd.SchoolID)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if existing == nil {
			return errs.NewError(ctx, status.SCHOOL_NOT_FOUND, nil,
				errors.New("school not found"))
		}

		patch := school.NewSchool()
		patch.SetSchoolId(existing.SchoolId())
		if cmd.Name != nil {
			patch.SetName(*cmd.Name)
		}
		if cmd.Description != nil {
			patch.SetDescription(cmd.Description)
		}
		if cmd.ImageKey != nil {
			patch.SetImageKey(cmd.ImageKey)
		}
		if cmd.District != nil {
			patch.SetDistrict(cmd.District)
		}
		if cmd.Province != nil {
			patch.SetProvince(cmd.Province)
		}
		if cmd.Note != nil {
			patch.SetNote(cmd.Note)
		}
		patch.SetModifyId(cmd.ActorID)

		if err := repos.School.Update(ctx, patch); err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}

		refreshed, err := repos.School.FindBySchoolId(ctx, cmd.SchoolID)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if refreshed == nil {
			return errs.NewError(ctx, status.SCHOOL_NOT_FOUND, nil,
				errors.New("school not found after update"))
		}
		updated = refreshed
		return nil
	})
	if err != nil {
		return nil, err
	}
	return updated, nil
}
