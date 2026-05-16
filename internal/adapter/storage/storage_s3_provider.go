package storage

import (
	"context"
	"fmt"
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
