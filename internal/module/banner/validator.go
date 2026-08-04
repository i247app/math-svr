package banner

import (
	"context"
	"strings"

	dto "math-ai.com/math-ai/internal/application/dto/banner"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/enum"
)

// Column-width mirrors of migration 022. Keeping the bounds here lets the
// validator reject over-long input with a friendly status code instead of
// letting MySQL truncate.
const (
	titleMaxLen       = 255
	shortTextMaxLen   = 1000
	mediaURLKeyMaxLen = 1000
	buttonTextMaxLen  = 255
	buttonLinkMaxLen  = 1000
	noteMaxLen        = 500
)

func ValidateCreateBanner(ctx context.Context, req *dto.CreateBannerReq) error {
	mediaType := enum.BannerMediaType(strings.TrimSpace(req.MediaType))
	if !mediaType.IsValid() {
		return errs.NewError(ctx, status.BANNER_INVALID_MEDIA_TYPE, nil, ErrInvalidMediaType)
	}
	// IMAGE/VIDEO banners must reference a stored media object; TEXT banners
	// // may leave the key empty.
	// if mediaType.RequiresMediaKey() && strings.TrimSpace(req.MediaURLKey) == "" {
	// 	return errs.NewError(ctx, status.BANNER_MISSING_MEDIA_URL, nil, ErrMediaURLKeyRequired)
	// }

	if len(req.MediaURLKey) > mediaURLKeyMaxLen {
		return errs.NewError(ctx, status.BANNER_MEDIA_URL_TOO_LONG, nil, ErrMediaURLKeyTooLong)
	}
	if err := validateBannerCommonFields(ctx, req.Title, req.ShortText, req.ButtonText, req.ButtonLinkURL, req.Note); err != nil {
		return err
	}
	if req.BannerStatus != nil {
		if err := validateBannerStatus(ctx, *req.BannerStatus); err != nil {
			return err
		}
	}
	return nil
}

func ValidateUpdateBanner(ctx context.Context, req *dto.UpdateBannerReq) error {
	if req.BannerID == 0 {
		return errs.NewError(ctx, status.BANNER_MISSING_ID, nil, ErrBannerIDRequired)
	}
	if req.MediaType != nil {
		mediaType := enum.BannerMediaType(strings.TrimSpace(*req.MediaType))
		if !mediaType.IsValid() {
			return errs.NewError(ctx, status.BANNER_INVALID_MEDIA_TYPE, nil, ErrInvalidMediaType)
		}
	}
	if req.MediaURLKey != nil && len(*req.MediaURLKey) > mediaURLKeyMaxLen {
		return errs.NewError(ctx, status.BANNER_MEDIA_URL_TOO_LONG, nil, ErrMediaURLKeyTooLong)
	}
	if err := validateBannerCommonFields(ctx, req.Title, req.ShortText, req.ButtonText, req.ButtonLinkURL, req.Note); err != nil {
		return err
	}
	if req.BannerStatus != nil {
		if err := validateBannerStatus(ctx, *req.BannerStatus); err != nil {
			return err
		}
	}
	return nil
}

// validateBannerCommonFields enforces the length caps shared by create and
// update. All args are optional pointers; nil skips the check.
func validateBannerCommonFields(ctx context.Context, title, shortText, buttonText, buttonLink, note *string) error {
	if title != nil && len(*title) > titleMaxLen {
		return errs.NewError(ctx, status.BANNER_TITLE_TOO_LONG, nil, ErrTitleTooLong)
	}
	if shortText != nil && len(*shortText) > shortTextMaxLen {
		return errs.NewError(ctx, status.BANNER_SHORT_TEXT_TOO_LONG, nil, ErrShortTextTooLong)
	}
	if buttonText != nil && len(*buttonText) > buttonTextMaxLen {
		return errs.NewError(ctx, status.BANNER_BUTTON_TEXT_TOO_LONG, nil, ErrButtonTextTooLong)
	}
	if buttonLink != nil && len(*buttonLink) > buttonLinkMaxLen {
		return errs.NewError(ctx, status.BANNER_BUTTON_LINK_TOO_LONG, nil, ErrButtonLinkTooLong)
	}
	if note != nil && len(*note) > noteMaxLen {
		return errs.NewError(ctx, status.BANNER_NOTE_TOO_LONG, nil, ErrNoteTooLong)
	}
	return nil
}

func validateBannerStatus(ctx context.Context, v string) error {
	if !enum.BannerStatusType(strings.TrimSpace(v)).IsValid() {
		return errs.NewError(ctx, status.BANNER_INVALID_STATUS, nil, ErrInvalidBannerStatus)
	}
	return nil
}

func ValidateGetBanner(ctx context.Context, req *dto.GetBannerReq) error {
	if req.BannerID == 0 {
		return errs.NewError(ctx, status.BANNER_MISSING_ID, nil, ErrBannerIDRequired)
	}
	return nil
}

func ValidateDeleteBanner(ctx context.Context, req *dto.DeleteBannerReq) error {
	if req.BannerID == 0 {
		return errs.NewError(ctx, status.BANNER_MISSING_ID, nil, ErrBannerIDRequired)
	}
	return nil
}

func ValidateListBanners(ctx context.Context, req *dto.ListBannersReq) error {
	// Empty/blank filters collapse to nil so the repo skips the predicate.
	if req.Search != nil && strings.TrimSpace(*req.Search) == "" {
		req.Search = nil
	}
	if req.MediaType != nil {
		trimmed := strings.TrimSpace(*req.MediaType)
		if trimmed == "" {
			req.MediaType = nil
		} else if !enum.BannerMediaType(trimmed).IsValid() {
			return errs.NewError(ctx, status.BANNER_INVALID_MEDIA_TYPE, nil, ErrInvalidMediaType)
		}
	}
	if req.BannerStatus != nil {
		trimmed := strings.TrimSpace(*req.BannerStatus)
		if trimmed == "" {
			req.BannerStatus = nil
		} else if !enum.BannerStatusType(trimmed).IsValid() {
			return errs.NewError(ctx, status.BANNER_INVALID_STATUS, nil, ErrInvalidBannerStatus)
		}
	}
	req.BannerIDs = sanitizeBannerIDs(req.BannerIDs)
	return nil
}

// sanitizeBannerIDs drops non-positive ids and removes duplicates while
// preserving caller order. Returns nil for an all-invalid input so the repo
// treats it as no filter rather than an empty IN(...) that matches zero rows.
func sanitizeBannerIDs(ids []int64) []int64 {
	if len(ids) == 0 {
		return nil
	}
	seen := make(map[int64]struct{}, len(ids))
	out := make([]int64, 0, len(ids))
	for _, id := range ids {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
