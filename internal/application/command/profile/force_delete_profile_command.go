package command

import (
	"context"

	"math-ai.com/math-ai/internal/application/transaction"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/infrastructure/logger"
)

type ForceDeleteProfileCommand struct {
	ProfileID int64
}

type ForceDeleteProfileCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewForceDeleteProfileCommandHandler(uow transaction.UnitOfWork) *ForceDeleteProfileCommandHandler {
	return &ForceDeleteProfileCommandHandler{uow: uow}
}

func (h *ForceDeleteProfileCommandHandler) Handle(ctx context.Context, cmd ForceDeleteProfileCommand) error {
	log := logger.From(ctx)
	err := h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		if err := repos.Profile.ForceDelete(ctx, cmd.ProfileID); err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		return nil
	})
	if err != nil {
		log.Warnf("profile.soft_delete profile_id=%d outcome=error err=%v", cmd.ProfileID, err)
		return err
	}
	log.Info("profile.soft_delete", "profile_id", cmd.ProfileID, "outcome", "success")
	return nil
}
