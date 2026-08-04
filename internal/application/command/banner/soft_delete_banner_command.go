package command

import (
	"context"

	"math-ai.com/math-ai/internal/application/transaction"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
)

// SoftDeleteBannerCommand flips the banner's status fields to the deleted
// markers so subsequent reads filter it out. The row stays in the table
// for auditing/recovery.
type SoftDeleteBannerCommand struct {
	BannerID int64
}

type SoftDeleteBannerCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewSoftDeleteBannerCommandHandler(uow transaction.UnitOfWork) *SoftDeleteBannerCommandHandler {
	return &SoftDeleteBannerCommandHandler{uow: uow}
}

func (h *SoftDeleteBannerCommandHandler) Handle(ctx context.Context, cmd SoftDeleteBannerCommand) error {
	return h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		existing, err := repos.Banner.FindByBannerId(ctx, cmd.BannerID)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if existing == nil {
			return errs.NewError(ctx, status.BANNER_NOT_FOUND, nil,
				ErrBannerNotFound)
		}

		if err := repos.Banner.SoftDeleteByBannerId(ctx, cmd.BannerID); err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		return nil
	})
}
