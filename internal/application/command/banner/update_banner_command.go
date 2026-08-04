package command

import (
	"context"

	"math-ai.com/math-ai/internal/application/transaction"
	"math-ai.com/math-ai/internal/domain/banner"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
)

// UpdateBannerCommand patches the banner row. Each pointer field is
// "leave unchanged when nil"; only the non-nil ones reach the repo's
// COALESCE/non-empty update so partial PATCH-style payloads work.
type UpdateBannerCommand struct {
	ActorID       *int64
	BannerID      int64
	Title         *string
	ShortText     *string
	MediaType     *string
	MediaURLKey   *string
	ButtonText    *string
	ButtonLinkURL *string
	Note          *string
	BannerStatus  *string
}

type UpdateBannerCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewUpdateBannerCommandHandler(uow transaction.UnitOfWork) *UpdateBannerCommandHandler {
	return &UpdateBannerCommandHandler{uow: uow}
}

func (h *UpdateBannerCommandHandler) Handle(ctx context.Context, cmd UpdateBannerCommand) (*banner.Banner, error) {
	var updated *banner.Banner

	err := h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		existing, err := repos.Banner.FindByBannerId(ctx, cmd.BannerID)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if existing == nil {
			return errs.NewError(ctx, status.BANNER_NOT_FOUND, nil,
				ErrBannerNotFound)
		}

		patch := banner.NewBanner()
		patch.SetBannerId(existing.BannerId())
		patch.SetTitle(cmd.Title)
		patch.SetShortText(cmd.ShortText)
		if cmd.MediaType != nil {
			patch.SetMediaType(*cmd.MediaType)
		}
		if cmd.MediaURLKey != nil {
			patch.SetMediaURLKey(*cmd.MediaURLKey)
		}
		patch.SetButtonText(cmd.ButtonText)
		patch.SetButtonLinkURL(cmd.ButtonLinkURL)
		patch.SetNote(cmd.Note)
		patch.SetBannerStatus(cmd.BannerStatus)
		patch.SetModifyId(cmd.ActorID)

		if err := repos.Banner.Update(ctx, patch); err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}

		refreshed, err := repos.Banner.FindByBannerId(ctx, cmd.BannerID)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if refreshed == nil {
			return errs.NewError(ctx, status.BANNER_NOT_FOUND, nil,
				ErrBannerNotFoundAfterUpdate)
		}
		updated = refreshed
		return nil
	})
	if err != nil {
		return nil, err
	}
	return updated, nil
}
