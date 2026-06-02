package user

import (
	"context"
	"time"

	"math-ai.com/math-ai/internal/adapter/storage"
	dto "math-ai.com/math-ai/internal/application/dto/user"
	userDTO "math-ai.com/math-ai/internal/application/dto/user"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/infrastructure/logger"
)

const imageUrlTTL = 1 * time.Hour

// uploadAvatarIfPresent ships the multipart avatar (if any) to S3 and returns
// the resulting key. Returns (nil, nil) when no avatar was submitted. The
// key lands on ma_users.avatar_key via BuildUser inside the create
// transaction.
func (s *Service) uploadAvatarIfPresent(ctx context.Context, req *dto.CreateUserReq) (*string, error) {
	if req.AvatarFile == nil || req.AvatarFilename == "" {
		return nil, nil
	}
	if s.storageProvider == nil {
		return nil, errs.NewError(ctx, status.STORAGE_CONFIG_INVALID, nil,
			ErrStorageAdapterNotConfigured)
	}

	if err := s.storageProvider.ValidateFileType(ctx, &storage.ValidateFileTypeRequest{
		Filename:    req.AvatarFilename,
		ContentType: req.AvatarContentType,
	}); err != nil {
		return nil, errs.NewError(ctx, status.USER_AVATAR_INVALID_FILE, nil, err)
	}

	uploaded, err := s.storageProvider.HandleUpload(ctx, &storage.UploadFileRequest{
		File:        req.AvatarFile,
		Filename:    req.AvatarFilename,
		ContentType: req.AvatarContentType,
		Folder:      avatarFolder,
	})
	if err != nil {
		return nil, errs.NewError(ctx, status.USER_AVATAR_UPLOAD_FAILED, nil, err)
	}
	if uploaded == nil || uploaded.Key == "" {
		return nil, errs.NewError(ctx, status.USER_AVATAR_UPLOAD_FAILED, nil,
			ErrUploadReturnedEmptyKey)
	}
	return &uploaded.Key, nil
}

func (s *Service) updateAvatarIfPresent(ctx context.Context, req *dto.UpdateUserReq) (*string, error) {
	if req.AvatarFile == nil || req.AvatarFilename == "" {
		return nil, nil
	}
	if s.storageProvider == nil {
		return nil, errs.NewError(ctx, status.STORAGE_CONFIG_INVALID, nil,
			ErrStorageAdapterNotConfigured)
	}

	if err := s.storageProvider.ValidateFileType(ctx, &storage.ValidateFileTypeRequest{
		Filename:    req.AvatarFilename,
		ContentType: req.AvatarContentType,
	}); err != nil {
		return nil, errs.NewError(ctx, status.USER_AVATAR_INVALID_FILE, nil, err)
	}

	uploaded, err := s.storageProvider.HandleUpload(ctx, &storage.UploadFileRequest{
		File:        req.AvatarFile,
		Filename:    req.AvatarFilename,
		ContentType: req.AvatarContentType,
		Folder:      avatarFolder,
	})
	if err != nil {
		return nil, errs.NewError(ctx, status.USER_AVATAR_UPLOAD_FAILED, nil, err)
	}
	if uploaded == nil || uploaded.Key == "" {
		return nil, errs.NewError(ctx, status.USER_AVATAR_UPLOAD_FAILED, nil,
			ErrUploadReturnedEmptyKey)
	}
	return &uploaded.Key, nil
}

// normalizeAvatarKey resolves a client-supplied avatar reference
// (either a bare S3 key or a URL pointing at our bucket) into the
// canonical key and validates that it lives under avatarFolder.
//
// The host allowlist + prefix enforcement live inside the storage
// provider so each backing vendor owns its own URL grammar. The
// service translates the resulting plain error into a typed
// MathError with the supplied invalidStatus code.
func (s *Service) normalizeAvatarKey(ctx context.Context, raw string, invalidStatus status.StatusCode) (string, error) {
	if s.storageProvider == nil {
		return "", errs.NewError(ctx, status.STORAGE_CONFIG_INVALID, nil,
			ErrStorageAdapterNotConfigured)
	}
	key, err := s.storageProvider.NormalizeKey(ctx, &storage.NormalizeKeyRequest{
		Raw: raw,
	})
	if err != nil {
		return "", errs.NewError(ctx, invalidStatus, nil, err)
	}
	if key == "" {
		return "", errs.NewError(ctx, invalidStatus, nil,
			ErrAvatarReferenceResolvedToEmptyKey)
	}
	return key, nil
}

func (s *Service) populateImageUrl(ctx context.Context, resp *userDTO.UserResponse) {
	if resp == nil || s.storageProvider == nil || resp.AvatarKey == nil || *resp.AvatarKey == "" {
		return
	}
	url, err := s.storageProvider.CreatePresignedUrl(ctx, &storage.CreatePresignedUrlRequest{
		Key:        *resp.AvatarKey,
		Expiration: imageUrlTTL,
	})
	if err != nil {
		logger.From(ctx).Warnf("user.avatar presign failed user_id=%s err=%v", resp.UserID, err)
		return
	}
	resp.AvatarUrl = &url
}
