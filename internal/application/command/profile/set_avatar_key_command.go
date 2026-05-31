package command

import (
	"context"
	"errors"

	"math-ai.com/math-ai/internal/application/transaction"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
)

// SetAvatarKeyCommand persists a newly-uploaded avatar's S3 key onto a
// profile. The S3 upload itself runs in the module layer (see profile
// service) because it's external I/O, not DB state — this command owns
// only the transactional DB write that links the two.
type SetAvatarKeyCommand struct {
	ProfileID int64
	AvatarKey string
}

type SetAvatarKeyCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewSetAvatarKeyCommandHandler(uow transaction.UnitOfWork) *SetAvatarKeyCommandHandler {
	return &SetAvatarKeyCommandHandler{uow: uow}
}

func (h *SetAvatarKeyCommandHandler) Handle(ctx context.Context, cmd SetAvatarKeyCommand) error {
	return h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		existing, err := repos.Profile.FindByProfileId(ctx, cmd.ProfileID)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if existing == nil {
			return errs.NewError(ctx, status.PROFILE_NOT_FOUND, nil,
				errors.New("profile not found"))
		}
		if err := repos.Profile.UpdateAvatarKey(ctx, cmd.ProfileID, cmd.AvatarKey); err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		return nil
	})
}
