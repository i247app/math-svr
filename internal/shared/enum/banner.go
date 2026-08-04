package enum

// BannerStatusType is the business lifecycle state of a banner row
// (ma_banners.banner_status). ACTIVE banners are shown to clients;
// INACTIVE ones are hidden but recoverable; DELETED marks a soft-delete.
type BannerStatusType string

const (
	BannerStatusTypeActive   BannerStatusType = "ACTIVE"
	BannerStatusTypeInactive BannerStatusType = "INACTIVE"
	BannerStatusTypeDeleted  BannerStatusType = "DELETED"
)

func (s BannerStatusType) String() string {
	return string(s)
}

func (s BannerStatusType) IsValid() bool {
	switch s {
	case BannerStatusTypeActive, BannerStatusTypeInactive:
		return true
	default:
		return false
	}
}

// BannerMediaType classifies the banner's primary payload
// (ma_banners.media_type). IMAGE/VIDEO reference an S3 object via
// media_url_key (presigned at the module edge); TEXT is title +
// short_text + button only and carries no media key.
type BannerMediaType string

const (
	BannerMediaTypeText  BannerMediaType = "TEXT"
	BannerMediaTypeImage BannerMediaType = "IMAGE"
	BannerMediaTypeVideo BannerMediaType = "VIDEO"
)

func (m BannerMediaType) String() string {
	return string(m)
}

func (m BannerMediaType) IsValid() bool {
	switch m {
	case BannerMediaTypeText, BannerMediaTypeImage, BannerMediaTypeVideo:
		return true
	default:
		return false
	}
}

// RequiresMediaKey reports whether the media type must be backed by a
// non-empty media_url_key. TEXT banners do not.
func (m BannerMediaType) RequiresMediaKey() bool {
	return m == BannerMediaTypeImage || m == BannerMediaTypeVideo
}
