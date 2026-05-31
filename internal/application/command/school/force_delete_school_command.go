package command

import (
	"context"
	"errors"

	"math-ai.com/math-ai/internal/application/transaction"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
)

// ForceDeleteSchoolCommand physically removes the school row. Any
// profile.school_id pointing here becomes a dangling reference; the
// profile service treats unresolved school_ids as nil at the API edge,
// so callers see school=null rather than an error.
type ForceDeleteSchoolCommand struct {
	SchoolID int64
}

type ForceDeleteSchoolCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewForceDeleteSchoolCommandHandler(uow transaction.UnitOfWork) *ForceDeleteSchoolCommandHandler {
	return &ForceDeleteSchoolCommandHandler{uow: uow}
}

func (h *ForceDeleteSchoolCommandHandler) Handle(ctx context.Context, cmd ForceDeleteSchoolCommand) error {
	return h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		existing, err := repos.School.FindBySchoolId(ctx, cmd.SchoolID)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if existing == nil {
			return errs.NewError(ctx, status.SCHOOL_NOT_FOUND, nil,
				errors.New("school not found"))
		}

		if err := repos.School.ForceDeleteBySchoolId(ctx, cmd.SchoolID); err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		return nil
	})
}
