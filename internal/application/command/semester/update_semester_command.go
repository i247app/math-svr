package command

import (
	"context"

	"math-ai.com/math-ai/internal/application/transaction"
	"math-ai.com/math-ai/internal/domain/semester"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
)

// UpdateSemesterCommand patches the parent row. Each pointer field is
// "leave unchanged when nil"; only the non-nil ones reach the repo's
// COALESCE update.
type UpdateSemesterCommand struct {
	ActorID      *int64
	SemesterID   int64
	Name         *string
	Description  *string
	ImageKey     *string
	DisplayOrder *int8
	Note         *string
}

type UpdateSemesterCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewUpdateSemesterCommandHandler(uow transaction.UnitOfWork) *UpdateSemesterCommandHandler {
	return &UpdateSemesterCommandHandler{uow: uow}
}

func (h *UpdateSemesterCommandHandler) Handle(ctx context.Context, cmd UpdateSemesterCommand) (*semester.Semester, error) {
	var updated *semester.Semester

	err := h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		existing, err := repos.Semester.FindBySemesterId(ctx, cmd.SemesterID)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if existing == nil {
			return errs.NewError(ctx, status.SEMESTER_NOT_FOUND, nil,
				ErrSemesterNotFound)
		}

		patched := semester.NewSemester()
		patched.SetSemesterId(existing.SemesterId())
		// display_order is int8 — zero is a legal value; pass through the
		// existing one unless the caller explicitly set a new value.
		if cmd.DisplayOrder != nil {
			patched.SetDisplayOrder(*cmd.DisplayOrder)
		} else {
			patched.SetDisplayOrder(existing.DisplayOrder())
		}
		if cmd.Name != nil {
			patched.SetName(*cmd.Name)
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

		if err := repos.Semester.Update(ctx, patched); err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}

		refreshed, err := repos.Semester.FindBySemesterId(ctx, cmd.SemesterID)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if refreshed == nil {
			return errs.NewError(ctx, status.SEMESTER_NOT_FOUND, nil,
				ErrSemesterNotFoundAfterUpdate)
		}
		updated = refreshed
		return nil
	})
	if err != nil {
		return nil, err
	}
	return updated, nil
}
