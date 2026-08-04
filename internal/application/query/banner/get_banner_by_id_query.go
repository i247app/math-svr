package query

import (
	"context"

	"math-ai.com/math-ai/internal/domain/banner"
)

type GetBannerByIdQuery struct {
	BannerID int64
}

type GetBannerByIdQueryHandler struct {
	bannerRepo banner.IRepository
}

func NewGetBannerByIdQueryHandler(bannerRepo banner.IRepository) *GetBannerByIdQueryHandler {
	return &GetBannerByIdQueryHandler{bannerRepo: bannerRepo}
}

func (h *GetBannerByIdQueryHandler) Handle(ctx context.Context, q GetBannerByIdQuery) (*banner.Banner, error) {
	return h.bannerRepo.FindByBannerId(ctx, q.BannerID)
}
