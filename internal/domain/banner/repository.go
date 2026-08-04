package banner

import (
	"context"

	"math-ai.com/math-ai/internal/shared/pagination"
)

// ListBannersParams narrows the listing/search query. Search is matched
// case-insensitively against title + short_text. MediaType and
// BannerStatus narrow by exact match when set — BannerStatus lets the
// client fetch only ACTIVE banners for display while admins can list
// INACTIVE ones too. BannerIds restricts the result set to the supplied
// external ids via an IN clause; nil or empty leaves it unfiltered.
// TakeAll bypasses pagination for admin exports.
type ListBannersParams struct {
	Search       *string
	MediaType    *string
	BannerStatus *string
	BannerIds    []int64
	Page         int64
	Limit        int64
	TakeAll      bool
}

// IRepository owns ma_banners persistence. Banners are single-language
// display content; there is no translation surface.
type IRepository interface {
	FindByBannerId(ctx context.Context, bannerId int64) (*Banner, error)
	ListBanners(ctx context.Context, params *ListBannersParams) ([]*Banner, *pagination.Pagination, error)
	Create(ctx context.Context, b *Banner) (*Banner, error)
	Update(ctx context.Context, b *Banner) error
	SoftDeleteByBannerId(ctx context.Context, bannerId int64) error
	ForceDeleteByBannerId(ctx context.Context, bannerId int64) error
}
