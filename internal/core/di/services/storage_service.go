package di

import (
	"context"
	"time"

	"math-ai.com/math-ai/internal/applications/dto"
	"math-ai.com/math-ai/internal/shared/constant/status"
)

// IStorageService defines the interface for file storage operations
type IStorageService interface {
	// Upload uploads a file to storage and returns the URL and preview URL
	Upload(ctx context.Context, req *dto.UploadFileRequest) (status.Code, *dto.UploadFileResponse, error)

	// Delete removes a file from storage using its key
	Delete(ctx context.Context, req *dto.DeleteFileRequest) (status.Code, error)

	// GetPreviewURL generates a preview URL from an original storage URL
	GetPreviewURL(ctx context.Context, originalURL string) (string, error)

	// CreatePresignedUrl generates a temporary presigned URL for secure access
	CreatePresignedUrl(ctx context.Context, key string, expiration time.Duration) (string, error)

	// ValidateFileType checks if the file type is allowed
	ValidateFileType(filename string, contentType string) error
}
