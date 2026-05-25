package user

import (
	"context"
	"errors"
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
			errors.New("storage adapter is not configured"))
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
			errors.New("upload returned an empty key"))
	}
	return &uploaded.Key, nil
}

func (s *Service) updateAvatarIfPresent(ctx context.Context, req *dto.UpdateUserReq) (*string, error) {
	if req.AvatarFile == nil || req.AvatarFilename == "" {
		return nil, nil
	}
	if s.storageProvider == nil {
		return nil, errs.NewError(ctx, status.STORAGE_CONFIG_INVALID, nil,
			errors.New("storage adapter is not configured"))
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
			errors.New("upload returned an empty key"))
	}
	return &uploaded.Key, nil
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
