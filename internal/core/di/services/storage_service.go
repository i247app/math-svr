package di

import (
	"context"

	"math-ai.com/math-ai/internal/applications/dto"
	"math-ai.com/math-ai/internal/shared/constant/status"
)

// IStorageService defines the interface for file storage operations
type IStorageService interface {
	HandleUpload(ctx context.Context, req *dto.UploadFileRequest) (status.Code, *dto.UploadFileResponse, error)

	// Delete removes a file from storage using its key
	HandleDelete(ctx context.Context, req *dto.DeleteFileRequest) (status.Code, error)

	// GetPreviewURL generates a preview URL from an original storage URL
	GetPreviewURL(ctx context.Context, req *dto.GetPreviewURLRequest) (status.Code, string, error)

	// CreatePresignedUrl generates a temporary presigned URL for secure access
	CreatePresignedUrl(ctx context.Context, req *dto.CreatePresignedUrlRequest) (status.Code, string, error)

	// ValidateFileType checks if the file type is allowed
	ValidateFileType(ctx context.Context, req *dto.ValidateFileTypeRequest) (status.Code, error)
}
