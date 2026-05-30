package storage

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"math-ai.com/math-ai/internal/libs/s3"
)

type S3Provider struct {
	client *s3.Client
}

func (p *S3Provider) Name() StorageProviderName {
	return ProviderS3
}

func NewS3Provider(client *s3.Client) *S3Provider {
	return &S3Provider{
		client: client,
	}
}

func (s *S3Provider) HandleUpload(ctx context.Context, req *UploadFileRequest) (*UploadFileResponse, error) {
	if req.File == nil || req.Filename == "" {
		return nil, nil
	}

	resp, err := s.client.Upload(ctx, req.File, req.Filename, req.ContentType, req.Folder)
	if err != nil {
		////logger.Errorf("Failed to upload avatar: %v", err)
		return nil, fmt.Errorf("failed to upload avatar: %w", err)
	}

	res := &UploadFileResponse{
		URL:      resp.URL,
		Key:      resp.Key,
		Filename: resp.Filename,
	}

	return res, nil
}

func (s *S3Provider) HandleDelete(ctx context.Context, req *DeleteFileRequest) error {
	err := s.client.Delete(ctx, req.Key)
	if err != nil {
		////logger.Errorf("Failed to delete file from S3: %v", err)
		return fmt.Errorf("failed to delete file from storage: %w", err)
	}
	return nil
}

func (s *S3Provider) GetPreviewURL(ctx context.Context, req *GetPreviewURLRequest, duration time.Duration) (string, error) {
	previewURL, err := s.client.GetPreviewURL(ctx, req.URL, duration)
	if err != nil {
		////logger.Errorf("Failed to generate preview URL: %v", err)
		return "", fmt.Errorf("failed to generate preview URL: %w", err)
	}
	return previewURL, nil
}

func (s *S3Provider) CreatePresignedUrl(ctx context.Context, req *CreatePresignedUrlRequest) (string, error) {
	presignedURL, err := s.client.CreatePresignedUrl(ctx, req.Key, req.Expiration)
	if err != nil {
		////logger.Errorf("Failed to create presigned URL: %v", err)
		return "", fmt.Errorf("failed to create presigned URL: %w", err)
	}
	return presignedURL, nil
}

func (s *S3Provider) ValidateFileType(ctx context.Context, req *ValidateFileTypeRequest) error {
	err := s.client.ValidateFileType(req.Filename, req.ContentType)
	if err != nil {
		////logger.Errorf("Invalid file type for file %s: %v", req.Filename, err)
		return fmt.Errorf("invalid file type: %w", err)
	}
	return nil
}

func (s *S3Provider) Download(ctx context.Context, req *DownloadFileRequest) ([]byte, error) {
	data, err := s.client.Download(ctx, req.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to download file from storage: %w", err)
	}
	return data, nil
}

// NormalizeKey accepts either a bare S3 object key or a full URL that
// points at the configured bucket and returns the canonical key.
//
// Host allowlist (the only forms we recognize as "ours"):
//
//	<bucket>.s3.<region>.amazonaws.com  (virtual-hosted, what UploadFileResponse.URL uses)
//	<bucket>.s3.amazonaws.com           (virtual-hosted, region-less legacy)
//	s3.<region>.amazonaws.com           (path-style)
//	s3.amazonaws.com                    (path-style, legacy)
//
// Any other host — including arbitrary CDNs, presigned URLs hosted off
// our bucket, third-party image URLs — is rejected. We do this on
// purpose: storing arbitrary URLs in avatar_key changes the column's
// contract and creates an SSRF/link-rot footgun. Add a CDN host here
// (and to the StorageConfig) the day we put CloudFront in front of the
// bucket.
//
// Prefix policy: the returned key MUST live under req.AllowedPrefix
// (e.g. "user-avatars"). This prevents a caller from claiming an
// object belonging to another aggregate as its own avatar.
func (s *S3Provider) NormalizeKey(ctx context.Context, req *NormalizeKeyRequest) (string, error) {
	if req == nil {
		return "", errors.New("storage: normalize key: nil request")
	}
	raw := strings.TrimSpace(req.Raw)
	if raw == "" {
		return "", nil
	}

	bucket := s.client.Bucket()
	region := s.client.Region()
	if bucket == "" {
		return "", errors.New("storage: bucket is not configured")
	}

	var key string
	if strings.Contains(raw, "://") {
		u, err := url.Parse(raw)
		if err != nil {
			return "", fmt.Errorf("storage: invalid avatar URL: %w", err)
		}
		if u.Scheme != "http" && u.Scheme != "https" {
			return "", fmt.Errorf("storage: unsupported avatar URL scheme %q", u.Scheme)
		}
		host := strings.ToLower(u.Host)
		path := strings.TrimPrefix(u.Path, "/")
		switch host {
		case bucket + ".s3." + region + ".amazonaws.com",
			bucket + ".s3.amazonaws.com":
			key = path
		case "s3." + region + ".amazonaws.com",
			"s3.amazonaws.com":
			if !strings.HasPrefix(path, bucket+"/") {
				return "", fmt.Errorf("storage: avatar URL bucket mismatch")
			}
			key = strings.TrimPrefix(path, bucket+"/")
		default:
			return "", fmt.Errorf("storage: avatar URL host %q not allowed", u.Host)
		}
	} else {
		key = strings.TrimPrefix(raw, "/")
	}

	if key == "" {
		return "", errors.New("storage: avatar key is empty")
	}

	if strings.Contains(key, "..") {
		return "", errors.New("storage: avatar key contains traversal segments")
	}
	// ma_users.avatar_key / ma_profiles.avatar_key column width.
	if len(key) > 256 {
		return "", errors.New("storage: avatar key exceeds 256 characters")
	}
	return key, nil
}
