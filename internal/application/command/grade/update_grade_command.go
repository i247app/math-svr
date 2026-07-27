package command

import (
	"context"

	"math-ai.com/math-ai/internal/application/transaction"
	"math-ai.com/math-ai/internal/domain/grade"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
)

// UpdateGradeCommand patches the parent row. Each pointer field is
// "leave unchanged when nil"; only the non-nil ones reach the repo's
// COALESCE update.
type UpdateGradeCommand struct {
	ActorID      *int64
	GradeID      int64
	Label        *string
	Description  *string
	ImageKey     *string
	DisplayOrder *int8
	Note         *string
}

type UpdateGradeCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewUpdateGradeCommandHandler(uow transaction.UnitOfWork) *UpdateGradeCommandHandler {
	return &UpdateGradeCommandHandler{uow: uow}
}

func (h *UpdateGradeCommandHandler) Handle(ctx context.Context, cmd UpdateGradeCommand) (*grade.Grade, error) {
	var updated *grade.Grade

	err := h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		existing, err := repos.Grade.FindByGradeId(ctx, cmd.GradeID)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if existing == nil {
			return errs.NewError(ctx, status.GRADE_NOT_FOUND, nil,
				ErrGradeNotFound)
		}

		patched := grade.NewGrade()
		patched.SetGradeId(existing.GradeId())
		// display_order is int8 — zero is a legal value; pass through the
		// existing one unless the caller explicitly set a new value.
		if cmd.DisplayOrder != nil {
			patched.SetDisplayOrder(*cmd.DisplayOrder)
		} else {
			patched.SetDisplayOrder(existing.DisplayOrder())
		}
		if cmd.Label != nil {
			patched.SetLabel(*cmd.Label)
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

		if err := repos.Grade.Update(ctx, patched); err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}

		refreshed, err := repos.Grade.FindByGradeId(ctx, cmd.GradeID)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if refreshed == nil {
			return errs.NewError(ctx, status.GRADE_NOT_FOUND, nil,
				ErrGradeNotFoundAfterUpdate)
		}
		updated = refreshed
		return nil
	})
	if err != nil {
		return nil, err
	}
	return updated, nil
}
