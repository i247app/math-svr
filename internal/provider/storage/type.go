package storage

import (
	"io"
	"time"
)

// UploadFileRequest represents a file upload request
type UploadFileRequest struct {
	File        io.Reader
	Filename    string
	ContentType string
	Folder      string // Optional folder/prefix in S3 bucket (e.g., "images", "videos")
}

// UploadFileResponse represents the result of a file upload
type UploadFileResponse struct {
	URL        string `json:"url"`         // Original S3 URL
	PreviewURL string `json:"preview_url"` // Preview URL for UI display
	Key        string `json:"key"`         // S3 object key
	Filename   string `json:"filename"`    // Original filename
	Size       int64  `json:"size"`        // File size in bytes
}

// DeleteFileRequest represents a file deletion request
type DeleteFileRequest struct {
	Key string `json:"key"` // S3 object key or full URL
}

// CreatePresignedUrlRequest represents a request to create a presigned URL
type CreatePresignedUrlRequest struct {
	Key        string        `json:"key"`        // S3 object key
	Expiration time.Duration `json:"expiration"` // URL expiration duration
}

// ValidateFileTypeRequest represents a request to validate file type
type ValidateFileTypeRequest struct {
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
}

// GetPreviewURLRequest represents a request to get preview URL
type GetPreviewURLRequest struct {
	URL string `json:"url"` // Original S3 URL
}

// GetPreviewURLResponse represents the preview URL response
type GetPreviewURLResponse struct {
	PreviewURL string `json:"preview_url"`
}

// DownloadFileRequest represents a request to download a file
type DownloadFileRequest struct {
	URL string `json:"url"` // S3 URL or key
}

// NormalizeKeyRequest asks the provider to coerce a client-supplied
// reference (either a full URL pointing at this provider's storage, or a
// bare object key) into a canonical key.
//
// AllowedPrefix is the folder the key MUST live under (e.g.
// "user-avatars"). Any other key is rejected — this stops a caller from
// claiming an object that belongs to another aggregate.
type NormalizeKeyRequest struct {
	Raw string
}
