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

	// NormalizeKey resolves a client-supplied reference (URL or bare key)
	// to a canonical object key managed by this provider. Returns "" when
	// req.Raw is empty. Returns an error if the reference points outside
	// the provider's bucket/CDN, fails parsing, or lives outside
	// req.AllowedPrefix.
	NormalizeKey(ctx context.Context, req *NormalizeKeyRequest) (string, error)
}
