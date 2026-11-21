package permission

import (
	"context"
	"io"

	"math-ai.com/math-ai/internal/application/dto"
	roledom "math-ai.com/math-ai/internal/domain/role"
	"math-ai.com/math-ai/internal/infrastructure/logger"
	"math-ai.com/math-ai/internal/provider/storage"
	"math-ai.com/math-ai/internal/shared/errs"
)

// roleImageFolder is the S3 prefix role images are uploaded under.
const roleImageFolder = "role-images"

// uploadRoleImage ships a multipart image to S3 and returns its key, or
// (nil, nil) when no file was sent. It runs before the UoW opens so a slow
// upload never holds a transaction; callers delete the object again if the
// write that should reference it fails (deleteRoleImage).
func (s *Service) uploadRoleImage(ctx context.Context, file roleImageFile) (*string, error) {
	if file.reader == nil || file.filename == "" {
		return nil, nil
	}
	if s.storageProvider == nil {
		return nil, errs.NewError(ctx, storage.StorageConfigInvalid, nil, ErrStorageProviderNotConfigured)
	}

	if err := s.storageProvider.ValidateFileType(ctx, &storage.ValidateFileTypeRequest{
		Filename:    file.filename,
		ContentType: file.contentType,
	}); err != nil {
		return nil, errs.NewError(ctx, roledom.RoleImageInvalidFile, nil, err)
	}

	uploaded, err := s.storageProvider.HandleUpload(ctx, &storage.UploadFileRequest{
		File:        file.reader,
		Filename:    file.filename,
		ContentType: file.contentType,
		Folder:      roleImageFolder,
	})
	if err != nil {
		return nil, errs.NewError(ctx, roledom.RoleImageUploadFailed, nil, err)
	}
	if uploaded == nil || uploaded.Key == "" {
		return nil, errs.NewError(ctx, roledom.RoleImageUploadFailed, nil, ErrUploadReturnedEmptyKey)
	}
	return &uploaded.Key, nil
}

// deleteRoleImage removes an object best-effort: the request outcome is
// already decided, so a failure only leaves an orphan, which is logged.
func (s *Service) deleteRoleImage(ctx context.Context, key *string, reason string) {
	if key == nil || *key == "" || s.storageProvider == nil {
		return
	}
	if err := s.storageProvider.HandleDelete(ctx, &storage.DeleteFileRequest{Key: *key}); err != nil {
		logger.From(ctx).Warnf("role.image %s cleanup failed key=%s err=%v", reason, *key, err)
	}
}

// populateRoleImageUrl adds a short-lived presigned URL when the role has an
// image. No-op when storage is disabled; a signing failure is logged and the
// URL omitted rather than failing the read.
func (s *Service) populateRoleImageUrl(ctx context.Context, resp *dto.RoleResponse) {
	if resp == nil || s.storageProvider == nil || resp.RoleImageKey == nil || *resp.RoleImageKey == "" {
		return
	}
	url, err := s.storageProvider.CreatePresignedUrl(ctx, &storage.CreatePresignedUrlRequest{
		Key:        *resp.RoleImageKey,
		Expiration: roleImageUrlTTL,
	})
	if err != nil {
		logger.From(ctx).Warnf("role.image presign failed role_id=%d err=%v", resp.RoleID, err)
		return
	}
	resp.RoleImageUrl = &url
}

// roleImageFile is the multipart file part as the create and update requests
// carry it.
type roleImageFile struct {
	reader      io.Reader
	filename    string
	contentType string
}
