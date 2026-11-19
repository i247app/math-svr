package models

type BannerModel struct {
	Audit

	BannerId      int64
	Title         *string
	ShortText     *string
	MediaType     string
	MediaURLKey   string
	ButtonText    *string
	ButtonLinkURL *string
	BannerStatus  *string
}
