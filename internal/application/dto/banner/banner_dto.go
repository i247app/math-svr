package banner

import (
	"io"

	domain "math-ai.com/math-ai/internal/domain/banner"
	"math-ai.com/math-ai/internal/shared/pagination"
)

// BannerResponse is the banner surface exposed to the API. MediaUrl is
// the short-lived presigned URL derived from MediaURLKey at the service
// edge for IMAGE/VIDEO banners; clients should prefer MediaUrl for
// display and MediaURLKey only when they need to re-sign later.
// ButtonLinkURL is an absolute link and is returned verbatim.
type BannerResponse struct {
	ID            int64   `json:"id"`
	BannerID      int64   `json:"banner_id"`
	Title         *string `json:"title,omitempty"`
	ShortText     *string `json:"short_text,omitempty"`
	MediaType     string  `json:"media_type"`
	MediaURLKey   string  `json:"media_url_key"`
	MediaUrl      *string `json:"media_url"`
	ButtonText    *string `json:"button_text,omitempty"`
	ButtonLinkURL *string `json:"button_link_url,omitempty"`
	Note          *string `json:"note,omitempty"`
	BannerStatus  *string `json:"banner_status,omitempty"`
	CreateDt      string  `json:"create_dt"`
	ModifyDt      string  `json:"modify_dt"`
}

type CreateBannerReq struct {
	Title         *string `json:"title,omitempty"`
	ShortText     *string `json:"short_text,omitempty"`
	MediaType     string  `json:"media_type"`
	MediaURLKey   string  `json:"-"`
	ButtonText    *string `json:"button_text,omitempty"`
	ButtonLinkURL *string `json:"-,omitempty"`
	Note          *string `json:"note,omitempty"`
	// BannerStatus optionally overrides the default ACTIVE state on create
	// (e.g. staging a banner as INACTIVE until it's ready to show).
	BannerStatus *string `json:"banner_status,omitempty"`

	Media string `json:"avatar,omitempty"`

	MediaFile        io.Reader `json:"-"` // multipart file reader
	MediaFilename    string    `json:"-"` // original filename
	MediaContentType string    `json:"-"` // MIME type
}

type CreateBannerRes struct {
	Banner *BannerResponse `json:"banner"`
}

// UpdateBannerReq uses pointer fields for every patchable column so a
// client can omit fields they don't intend to change. A nil pointer
// means "leave alone"; a non-nil value overwrites.
type UpdateBannerReq struct {
	BannerID      int64   `json:"banner_id"`
	Title         *string `json:"title,omitempty"`
	ShortText     *string `json:"short_text,omitempty"`
	MediaType     *string `json:"media_type,omitempty"`
	MediaURLKey   *string `json:"media_url_key,omitempty"`
	ButtonText    *string `json:"button_text,omitempty"`
	ButtonLinkURL *string `json:"button_link_url,omitempty"`
	Note          *string `json:"note,omitempty"`
	BannerStatus  *string `json:"banner_status,omitempty"`

	Media string `json:"avatar,omitempty"`

	MediaFile        io.Reader `json:"-"` // multipart file reader
	MediaFilename    string    `json:"-"` // original filename
	MediaContentType string    `json:"-"` // MIME type
}

type UpdateBannerRes struct {
	Banner *BannerResponse `json:"banner"`
}

type DeleteBannerReq struct {
	BannerID int64 `json:"banner_id"`
}

type DeleteBannerRes struct{}

type GetBannerReq struct {
	BannerID int64 `json:"banner_id"`
}

type GetBannerRes struct {
	Banner *BannerResponse `json:"banner"`
}

type ListBannersReq struct {
	Search       *string `json:"search,omitempty"`
	MediaType    *string `json:"media_type,omitempty"`
	BannerStatus *string `json:"banner_status,omitempty"`
	BannerIDs    []int64 `json:"banner_ids,omitempty"`
	Page         int64   `json:"page"`
	Size         int64   `json:"size"`
}

type ListBannersRes struct {
	Banners    []*BannerResponse      `json:"banners"`
	Pagination *pagination.Pagination `json:"pagination"`
}

func DomainToResponse(b *domain.Banner) *BannerResponse {
	if b == nil {
		return nil
	}
	return &BannerResponse{
		ID:            b.Id(),
		BannerID:      b.BannerId(),
		Title:         b.Title(),
		ShortText:     b.ShortText(),
		MediaType:     b.MediaType(),
		MediaURLKey:   b.MediaURLKey(),
		ButtonText:    b.ButtonText(),
		ButtonLinkURL: b.ButtonLinkURL(),
		Note:          b.Note(),
		BannerStatus:  b.BannerStatus(),
		CreateDt:      b.CreateDt().String(),
		ModifyDt:      b.ModifyDt().String(),
	}
}

func DomainListToResponse(banners []*domain.Banner) []*BannerResponse {
	result := make([]*BannerResponse, len(banners))
	for i, b := range banners {
		result[i] = DomainToResponse(b)
	}
	return result
}
