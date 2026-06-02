package profile

import (
	"context"

	"math-ai.com/math-ai/internal/adapter/storage"
	dto "math-ai.com/math-ai/internal/application/dto/profile"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
)

// UploadAvatar streams the request body to S3 then persists the resulting
// key against the profile in a single transaction. Returns the key plus a
// short-lived presigned URL for immediate display.
func (s *Service) UploadAvatarStatic(ctx context.Context, req *dto.UploadFileRequest) (*dto.UploadFileResponse, error) {
	if err := s.storageProvider.ValidateFileType(ctx, &storage.ValidateFileTypeRequest{
		Filename:    req.Filename,
		ContentType: req.ContentType,
	}); err != nil {
		return nil, errs.NewError(ctx, status.PROFILE_AVATAR_INVALID_FILE, nil, err)
	}

	uploaded, err := s.storageProvider.HandleUpload(ctx, &storage.UploadFileRequest{
		File:        req.File,
		Filename:    req.Filename,
		ContentType: req.ContentType,
		Folder:      avatarFolder,
	})
	if err != nil {
		return nil, errs.NewError(ctx, status.PROFILE_AVATAR_UPLOAD_FAILED, nil, err)
	}
	if uploaded == nil || uploaded.Key == "" {
		return nil, errs.NewError(ctx, status.PROFILE_AVATAR_UPLOAD_FAILED, nil,
			ErrUploadReturnedEmptyKey)
	}

	return &dto.UploadFileResponse{
		URL:        uploaded.URL,
		PreviewURL: uploaded.PreviewURL,
		Key:        uploaded.Key,
		Filename:   uploaded.Filename,
		Size:       uploaded.Size,
	}, nil
}

func (s *Service) DeleteFile(ctx context.Context, req *dto.DeleteFileRequest) (*dto.DeleteFileResponse, error) {
	if err := s.storageProvider.HandleDelete(ctx, &storage.DeleteFileRequest{
		Key: req.Key,
	}); err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	return &dto.DeleteFileResponse{
		Message: "File deleted successfully",
	}, nil
}
