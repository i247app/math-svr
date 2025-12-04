package services

import (
	"context"
	"fmt"

	"math-ai.com/math-ai/internal/applications/dto"
	di "math-ai.com/math-ai/internal/core/di/services"
	"math-ai.com/math-ai/internal/shared/constant/status"
	"math-ai.com/math-ai/pkg/aws/s3"
)

// StorageService implements the storage service for S3
type StorageService struct {
	s3Client *s3.Client
}

// NewStorageService creates a new storage service instance
func NewStorageService(s3Client *s3.Client) di.IStorageService {
	return &StorageService{
		s3Client: s3Client,
	}
}

func (s *StorageService) HandleUpload(ctx context.Context, req *dto.UploadFileRequest) (status.Code, *dto.UploadFileResponse, error) {
	if req.File == nil || req.Filename == "" {
		return status.FAIL, nil, nil
	}

	res, err := s.s3Client.Upload(ctx, req.File, req.Filename, req.ContentType, req.Folder)
	if err != nil {
		////logger.Errorf("Failed to upload avatar: %v", err)
		return status.FAIL, nil, fmt.Errorf("failed to upload avatar: %w", err)
	}

	// Return preview URL for UI display
	return status.SUCCESS, res, nil
}

func (s *StorageService) HandleDelete(ctx context.Context, req *dto.DeleteFileRequest) (status.Code, error) {
	err := s.s3Client.Delete(ctx, req.Key)
	if err != nil {
		////logger.Errorf("Failed to delete file from S3: %v", err)
		return status.FAIL, fmt.Errorf("failed to delete file from storage: %w", err)
	}
	return status.SUCCESS, nil
}

func (s *StorageService) GetPreviewURL(ctx context.Context, req *dto.GetPreviewURLRequest) (status.Code, string, error) {
	previewURL, err := s.s3Client.GetPreviewURL(ctx, req.URL)
	if err != nil {
		////logger.Errorf("Failed to generate preview URL: %v", err)
		return status.FAIL, "", fmt.Errorf("failed to generate preview URL: %w", err)
	}
	return status.SUCCESS, previewURL, nil
}

func (s *StorageService) CreatePresignedUrl(ctx context.Context, req *dto.CreatePresignedUrlRequest) (status.Code, string, error) {
	presignedURL, err := s.s3Client.CreatePresignedUrl(ctx, req.Key, req.Expiration)
	if err != nil {
		////logger.Errorf("Failed to create presigned URL: %v", err)
		return status.FAIL, "", fmt.Errorf("failed to create presigned URL: %w", err)
	}
	return status.SUCCESS, presignedURL, nil
}

func (s *StorageService) ValidateFileType(ctx context.Context, req *dto.ValidateFileTypeRequest) (status.Code, error) {
	err := s.s3Client.ValidateFileType(req.Filename, req.ContentType)
	if err != nil {
		////logger.Errorf("Invalid file type for file %s: %v", req.Filename, err)
		return status.FAIL, fmt.Errorf("invalid file type: %w", err)
	}
	return status.SUCCESS, nil
}
