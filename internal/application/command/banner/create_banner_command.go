package command

import (
	"context"

	"math-ai.com/math-ai/internal/application/command/shared/seqgen"
	"math-ai.com/math-ai/internal/application/transaction"
	"math-ai.com/math-ai/internal/domain/banner"
	"math-ai.com/math-ai/internal/domain/seq"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/enum"
)

type CreateBannerCommand struct {
	ActorID       *int64
	Title         *string
	ShortText     *string
	MediaType     string
	MediaURLKey   string
	ButtonText    *string
	ButtonLinkURL *string
	Note          *string
	BannerStatus  *string
}

type CreateBannerCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewCreateBannerCommandHandler(uow transaction.UnitOfWork) *CreateBannerCommandHandler {
	return &CreateBannerCommandHandler{uow: uow}
}

func (h *CreateBannerCommandHandler) Handle(ctx context.Context, cmd CreateBannerCommand) (*banner.Banner, error) {
	var created *banner.Banner

	err := h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		bannerID, err := seqgen.Next(ctx, repos.Seq, seq.NameBanner)
		if err != nil {
			return err
		}

		b := banner.NewBanner()
		b.SetBannerId(bannerID)
		b.SetTitle(cmd.Title)
		b.SetShortText(cmd.ShortText)
		b.SetMediaType(cmd.MediaType)
		b.SetMediaURLKey(cmd.MediaURLKey)
		b.SetButtonText(cmd.ButtonText)
		b.SetButtonLinkURL(cmd.ButtonLinkURL)
		b.SetNote(cmd.Note)
		// Default to ACTIVE so newly-created banners surface under the
		// client's `banner_status = ACTIVE` display filter instead of
		// landing as NULL.
		bannerStatus := cmd.BannerStatus
		if bannerStatus == nil {
			active := enum.BannerStatusTypeActive.String()
			bannerStatus = &active
		}
		b.SetBannerStatus(bannerStatus)
		b.SetCreateId(cmd.ActorID)

		saved, err := repos.Banner.Create(ctx, b)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if saved == nil {
			return errs.NewError(ctx, status.BANNER_NOT_FOUND, nil,
				ErrBannerNotFoundAfterInsert)
		}
		created = saved
		return nil
	})
	if err != nil {
		return nil, err
	}
	return created, nil
}
