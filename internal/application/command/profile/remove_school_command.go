package command

import (
	"context"

	"math-ai.com/math-ai/internal/application/transaction"
	"math-ai.com/math-ai/internal/domain/profile"
	"math-ai.com/math-ai/internal/domain/school"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/errs"
)

// RemoveSchoolCommand clears profile.school_id. Distinct from Update —
// which uses COALESCE and cannot express "set this column to NULL" — so
// the caller can drop the school link without having to know the
// previous value or pass any sentinel.
type RemoveSchoolCommand struct {
	ProfileID int64
}

type RemoveSchoolCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewRemoveSchoolCommandHandler(uow transaction.UnitOfWork) *RemoveSchoolCommandHandler {
	return &RemoveSchoolCommandHandler{uow: uow}
}

func (h *RemoveSchoolCommandHandler) Handle(ctx context.Context, cmd RemoveSchoolCommand) (*profile.Profile, error) {
	var updated *profile.Profile

	err := h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		existing, err := repos.Profile.FindByProfileId(ctx, cmd.ProfileID)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if existing == nil {
			return errs.NewError(ctx, profile.ProfileNotFound, nil,
				ErrProfileNotFound)
		}

		if err := repos.Profile.SetSchoolId(ctx, cmd.ProfileID, nil); err != nil {
			return errs.NewError(ctx, school.SchoolProfileLinkFailed, nil, err)
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
