package command

import (
	"context"
	"errors"

	"math-ai.com/math-ai/internal/application/transaction"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
)

// SoftDeleteSchoolCommand flips the school's status fields to the deleted
// markers so subsequent reads filter it out. Profile rows that referenced
// this school keep their school_id column intact — surfacing it as nil
// in responses is the composing service's responsibility.
type SoftDeleteSchoolCommand struct {
	SchoolID string
}

type SoftDeleteSchoolCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewSoftDeleteSchoolCommandHandler(uow transaction.UnitOfWork) *SoftDeleteSchoolCommandHandler {
	return &SoftDeleteSchoolCommandHandler{uow: uow}
}

func (h *SoftDeleteSchoolCommandHandler) Handle(ctx context.Context, cmd SoftDeleteSchoolCommand) error {
	return h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		existing, err := repos.School.FindBySchoolId(ctx, cmd.SchoolID)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if existing == nil {
			return errs.NewError(ctx, status.SCHOOL_NOT_FOUND, nil,
				errors.New("school not found"))
		}

		if err := repos.School.SoftDeleteBySchoolId(ctx, cmd.SchoolID); err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		return nil
	})
}
