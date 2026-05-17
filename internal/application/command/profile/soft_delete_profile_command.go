package command

import (
	"context"

	"github.com/google/uuid"
	"math-ai.com/math-ai/internal/application/transaction"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/infrastructure/logger"
)

type SoftDeleteProfileCommand struct {
	ProfileID uuid.UUID
}

type SoftDeleteProfileCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewSoftDeleteProfileCommandHandler(uow transaction.UnitOfWork) *SoftDeleteProfileCommandHandler {
	return &SoftDeleteProfileCommandHandler{uow: uow}
}

func (h *SoftDeleteProfileCommandHandler) Handle(ctx context.Context, cmd SoftDeleteProfileCommand) error {
	log := logger.From(ctx)
	err := h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		if err := repos.Profile.SoftDelete(ctx, cmd.ProfileID); err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		return nil
	})
	if err != nil {
		log.Warnf("profile.soft_delete profile_id=%s outcome=error err=%v", cmd.ProfileID, err)
		return err
	}
	log.Info("profile.soft_delete", "profile_id", cmd.ProfileID, "outcome", "success")
	return nil
}

// entityDeletedStatus mirrors the "DELETED" literal written to every
// <entity>_status column on soft-delete (defined locally to the command
// package so it doesn't reach back into the repositories layer).
const entityDeletedStatus = "DELETED"
