package query

import (
	"context"

	"math-ai.com/math-ai/internal/domain/banner"
	"math-ai.com/math-ai/internal/shared/pagination"
)

// ListBannersQuery covers both the plain list and the search endpoint —
// Search is just a non-nil filter on the same handler so the repo only
// has one code path to maintain. MediaType and BannerStatus narrow the
// result; BannerIDs restricts it to the supplied external ids (IN-clause).
type ListBannersQuery struct {
	Search       *string
	MediaType    *string
	BannerStatus *string
	BannerIDs    []int64
	Page         int64
	Limit        int64
}

type ListBannersQueryHandler struct {
	bannerRepo banner.IRepository
}

func NewListBannersQueryHandler(bannerRepo banner.IRepository) *ListBannersQueryHandler {
	return &ListBannersQueryHandler{bannerRepo: bannerRepo}
}

func (h *ListBannersQueryHandler) Handle(ctx context.Context, q ListBannersQuery) ([]*banner.Banner, *pagination.Pagination, error) {
	return h.bannerRepo.ListBanners(ctx, &banner.ListBannersParams{
		Search:       q.Search,
		MediaType:    q.MediaType,
		BannerStatus: q.BannerStatus,
		BannerIds:    q.BannerIDs,
		Page:         q.Page,
		Limit:        q.Limit,
	})
}
