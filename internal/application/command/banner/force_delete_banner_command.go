package command

import (
	"context"

	"math-ai.com/math-ai/internal/application/transaction"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
)

// ForceDeleteBannerCommand physically removes the banner row.
type ForceDeleteBannerCommand struct {
	BannerID int64
}

type ForceDeleteBannerCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewForceDeleteBannerCommandHandler(uow transaction.UnitOfWork) *ForceDeleteBannerCommandHandler {
	return &ForceDeleteBannerCommandHandler{uow: uow}
}

func (h *ForceDeleteBannerCommandHandler) Handle(ctx context.Context, cmd ForceDeleteBannerCommand) error {
	return h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		existing, err := repos.Banner.FindByBannerId(ctx, cmd.BannerID)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if existing == nil {
			return errs.NewError(ctx, status.BANNER_NOT_FOUND, nil,
				ErrBannerNotFound)
		}

		if err := repos.Banner.ForceDeleteByBannerId(ctx, cmd.BannerID); err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		return nil
	})
}
