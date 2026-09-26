package grade

import (
	"context"
	"time"

	"math-ai.com/math-ai/internal/adapter/storage"
	dto "math-ai.com/math-ai/internal/application/dto/grade"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
)

const (
	// avatarFolder is the S3 prefix new avatar uploads land under.
	avatarFolder = "grades"
	// avatarUrlTTL bounds how long a generated avatar preview URL is valid.
	// Short enough that a stale link in the wild expires quickly; long
	// enough that a single screen render doesn't have to refresh mid-view.
	avatarUrlTTL = 1 * time.Hour
	// refImageUrlTTL bounds presigned URLs we hand back for embedded
	// program/grade/semester image_keys.
	refImageUrlTTL = 1 * time.Hour
)

// uploadAvatarIfPresent ships the multipart avatar (if any) to S3 and returns
// the resulting key. Returns (nil, nil) when no avatar was submitted. The
// key lands on ma_users.avatar_key via BuildUser inside the create
// transaction.
func (s *Service) uploadImageIfPresent(ctx context.Context, req *dto.UpdateGradeReq) (*string, error) {
	if req.ImageFile == nil || req.ImageFilename == "" {
		return nil, nil
	}
	if s.storageProvider == nil {
		return nil, errs.NewError(ctx, status.STORAGE_CONFIG_INVALID, nil, ErrStorageAdapterNotConfigured)
	}

	if err := s.storageProvider.ValidateFileType(ctx, &storage.ValidateFileTypeRequest{
		Filename:    req.ImageFilename,
		ContentType: req.ImageContentType,
	}); err != nil {
		return nil, errs.NewError(ctx, status.PROFILE_AVATAR_INVALID_FILE, nil, err)
	}

	uploaded, err := s.storageProvider.HandleUpload(ctx, &storage.UploadFileRequest{
		File:        req.ImageFile,
		Filename:    req.ImageFilename,
		ContentType: req.ImageContentType,
		Folder:      avatarFolder,
	})
	if err != nil {
		return nil, errs.NewError(ctx, status.PROFILE_AVATAR_UPLOAD_FAILED, nil, err)
	}
	if uploaded == nil || uploaded.Key == "" {
		return nil, errs.NewError(ctx, status.PROFILE_AVATAR_UPLOAD_FAILED, nil,
			ErrUploadReturnedEmptyKey)
	}
	return &uploaded.Key, nil
}
