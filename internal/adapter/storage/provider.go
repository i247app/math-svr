package storage

import (
	"context"
	"time"
)

type StorageProvider interface {
	Name() StorageProviderName

	HandleUpload(ctx context.Context, req *UploadFileRequest) (*UploadFileResponse, error)

	// Delete removes a file from storage using its key
	HandleDelete(ctx context.Context, req *DeleteFileRequest) error

	// GetPreviewURL generates a preview URL from an original storage URL
	GetPreviewURL(ctx context.Context, req *GetPreviewURLRequest, duration time.Duration) (string, error)

	// CreatePresignedUrl generates a temporary presigned URL for secure access
	CreatePresignedUrl(ctx context.Context, req *CreatePresignedUrlRequest) (string, error)

	// ValidateFileType checks if the file type is allowed
	ValidateFileType(ctx context.Context, req *ValidateFileTypeRequest) error

	// Download fetches file content from storage and returns the bytes
	Download(ctx context.Context, req *DownloadFileRequest) ([]byte, error)
}
