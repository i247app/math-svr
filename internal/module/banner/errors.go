package banner

import "errors"

// Module-scoped sentinel errors. See conventions.md §Errors.
var (
	ErrBannerIDRequired      = errors.New("banner_id is required")
	ErrBannerNotFound        = errors.New("banner not found")
	ErrButtonLinkTooLong     = errors.New("button_link_url too long")
	ErrButtonTextTooLong     = errors.New("button_text too long")
	ErrInvalidMediaType      = errors.New("invalid media_type")
	ErrInvalidBannerStatus   = errors.New("invalid banner_status")
	ErrMediaURLKeyRequired   = errors.New("media_url_key is required for image/video banners")
	ErrMediaURLKeyTooLong    = errors.New("media_url_key too long")
	ErrNoteTooLong           = errors.New("note too long")
	ErrShortTextTooLong      = errors.New("short_text too long")
	ErrTitleTooLong          = errors.New("title too long")
	ErrMediaReferenceTooLong = errors.New("media_reference too long")

	ErrStorageAdapterNotConfigured       = errors.New("storage adapter is not configured")
	ErrUploadReturnedEmptyKey            = errors.New("upload returned an empty key")
	ErrBannerReferenceResolvedToEmptyKey = errors.New("avatar reference resolved to empty key")
)
