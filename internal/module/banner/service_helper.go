package banner

import (
	"context"

	"math-ai.com/math-ai/internal/application/dto"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/provider/storage"
)

// normalizeBannerKey resolves a client-supplied media reference into a
// canonical S3 key. Host allowlist + prefix enforcement live in the
// storage provider; the service layer just translates a plain error into
// a typed MathError.
func (s *Service) normalizeBannerKey(ctx context.Context, raw string, invalidStatus status.StatusCode) (string, error) {
	if s.storageProvider == nil {
		return "", errs.NewError(ctx, status.STORAGE_CONFIG_INVALID, nil,
			ErrStorageProviderNotConfigured)
	}
	key, err := s.storageProvider.NormalizeKey(ctx, &storage.NormalizeKeyRequest{
		Raw: raw,
	})
	if err != nil {
		return "", errs.NewError(ctx, invalidStatus, nil, err)
	}
	if key == "" {
		return "", errs.NewError(ctx, invalidStatus, nil, ErrBannerReferenceResolvedToEmptyKey)
	}
	return key, nil
}

func (s *Service) uploadBannerIfPresent(ctx context.Context, req *dto.CreateBannerReq) (*string, error) {
	if req.MediaFile == nil || req.MediaFilename == "" {
		return nil, nil
	}
	if s.storageProvider == nil {
		return nil, errs.NewError(ctx, status.STORAGE_CONFIG_INVALID, nil,
			ErrStorageProviderNotConfigured)
	}

	if err := s.storageProvider.ValidateFileType(ctx, &storage.ValidateFileTypeRequest{
		Filename:    req.MediaFilename,
		ContentType: req.MediaContentType,
	}); err != nil {
		return nil, errs.NewError(ctx, status.BANNER_MEDIA_INVALID_FILE, nil, err)
	}

	uploaded, err := s.storageProvider.HandleUpload(ctx, &storage.UploadFileRequest{
		File:        req.MediaFile,
		Filename:    req.MediaFilename,
		ContentType: req.MediaContentType,
		Folder:      bannerFolder,
	})
	if err != nil {
		return nil, errs.NewError(ctx, status.BANNER_MEDIA_UPLOAD_FAILED, nil, err)
	}
	if uploaded == nil || uploaded.Key == "" {
		return nil, errs.NewError(ctx, status.BANNER_MEDIA_UPLOAD_FAILED, nil,
			ErrUploadReturnedEmptyKey)
	}
	return &uploaded.Key, nil
}

func (s *Service) updateBannerIfPresent(ctx context.Context, req *dto.UpdateBannerReq) (*string, error) {
	if req.MediaFile == nil || req.MediaFilename == "" {
		return nil, nil
	}
	if s.storageProvider == nil {
		return nil, errs.NewError(ctx, status.STORAGE_CONFIG_INVALID, nil,
			ErrStorageProviderNotConfigured)
	}

	if err := s.storageProvider.ValidateFileType(ctx, &storage.ValidateFileTypeRequest{
		Filename:    req.MediaFilename,
		ContentType: req.MediaContentType,
	}); err != nil {
		return nil, errs.NewError(ctx, status.BANNER_MEDIA_INVALID_FILE, nil, err)
	}

	uploaded, err := s.storageProvider.HandleUpload(ctx, &storage.UploadFileRequest{
		File:        req.MediaFile,
		Filename:    req.MediaFilename,
		ContentType: req.MediaContentType,
		Folder:      bannerFolder,
	})
	if err != nil {
		return nil, errs.NewError(ctx, status.BANNER_MEDIA_UPLOAD_FAILED, nil, err)
	}
	if uploaded == nil || uploaded.Key == "" {
		return nil, errs.NewError(ctx, status.BANNER_MEDIA_UPLOAD_FAILED, nil,
			ErrUploadReturnedEmptyKey)
	}
	return &uploaded.Key, nil
}
