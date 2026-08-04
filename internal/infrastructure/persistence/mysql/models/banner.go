package models

import "time"

type BannerModel struct {
	Id            int64
	BannerId      int64
	Title         *string
	ShortText     *string
	MediaType     string
	MediaURLKey   string
	ButtonText    *string
	ButtonLinkURL *string
	Note          *string
	BannerStatus  *string
	Status        string
	CreateId      *int64
	CreateDt      time.Time
	ModifyId      *int64
	ModifyDt      time.Time
}
