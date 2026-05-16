package s3

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"math-ai.com/math-ai/internal/infrastructure/logger"
	"math-ai.com/math-ai/internal/shared/utils"
)

type Config struct {
	AccessKey string
	SecretKey string
	Region    string
	Bucket    string
}

type Client struct {
	s3Client *s3.Client
	cfg      Config
}

// NewS3Client creates a new storage service instance
func NewClient(cfg Config) *Client {
	awsConfig := aws.Config{
		Region: cfg.Region,
		Credentials: credentials.NewStaticCredentialsProvider(
			cfg.AccessKey,
			cfg.SecretKey,
			"",
		),
	}

	// Create S3 client
	s3Client := s3.NewFromConfig(awsConfig)

	return &Client{
		s3Client: s3Client,
		cfg:      cfg,
	}
}

type UploadFileResponse struct {
	URL        string `json:"url"`         // Original S3 URL
	PreviewURL string `json:"preview_url"` // Preview URL for UI display
	Key        string `json:"key"`         // S3 object key
	Filename   string `json:"filename"`    // Original filename
	Size       int64  `json:"size"`        // File size in bytes
}

// Upload uploads a file to S3 and returns URLs
func (s *Client) Upload(ctx context.Context, file io.Reader, filename, contentType, folder string) (*UploadFileResponse, error) {
	logger := logger.From(ctx)

	// Validate file type
	if err := s.ValidateFileType(filename, contentType); err != nil {
		logger.Errorf("File validation failed: %v", err)
		return nil, fmt.Errorf("invalid file type: %w", err)
	}

	// Generate unique key for the file
	key := s.generateKey(filename, folder)

	// Determine content type
	if contentType == "" {
		contentType = utils.GetContentTypeFromFilename(filename)
	}

	// Read the file into memory (for getting size)
	// Note: For large files, consider using multipart upload
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		logger.Errorf("Failed to read file: %v", err)
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	fileSize := int64(len(fileBytes))

	// Upload to S3
	_, err = s.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.cfg.Bucket),
		Key:         aws.String(key),
		Body:        strings.NewReader(string(fileBytes)),
		ContentType: aws.String(contentType),
	})

	if err != nil {
		logger.Errorf("Failed to upload file to S3: %v", err)
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	// Generate URLs
	originalURL := s.generateS3URL(key)

	logger.Infof("Successfully uploaded file: %s (key: %s, size: %d bytes)", filename, key, fileSize)

	response := &UploadFileResponse{
		URL:      originalURL,
		Key:      key,
		Filename: filename,
		Size:     fileSize,
	}

	return response, nil
}

// Delete removes a file from S3
func (s *Client) Delete(ctx context.Context, originalURL string) error {
	logger := logger.From(ctx)

	// Extract key from URL if full URL was provided
	key := s.extractKeyFromURL(originalURL)

	if key == "" {
		return fmt.Errorf("invalid or empty file key")
	}

	// Delete from S3
	_, err := s.s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.cfg.Bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		logger.Errorf("Failed to delete file from S3 (key: %s): %v", key, err)
		return fmt.Errorf("failed to delete file: %w", err)
	}

	logger.Infof("Successfully deleted file: %s", key)
	return nil
}

// GetPreviewURL generates a preview URL from an original S3 URL
func (s *Client) GetPreviewURL(ctx context.Context, originalURL string, duration time.Duration) (string, error) {
	logger := logger.From(ctx)

	key := s.extractKeyFromURL(originalURL)
	if key == "" {
		return "", fmt.Errorf("invalid URL: cannot extract key")
	}

	previewUrl, err := s.CreatePresignedUrl(ctx, key, time.Hour)
	if err != nil {
		logger.Errorf("Failed to generate preview URL for key %s: %v", key, err)
		return "", fmt.Errorf("failed to generate preview URL: %w", err)
	}

	return previewUrl, nil
}

// ValidateFileType checks if the file type is allowed
func (s *Client) ValidateFileType(filename string, contentType string) error {
	return utils.ValidateFile(filename, contentType)
}

// CreatePresignedUrl generates a temporary presigned URL for secure S3 object access
func (s *Client) CreatePresignedUrl(ctx context.Context, key string, expiration time.Duration) (string, error) {
	logger := logger.From(ctx)

	// // Extract key from URL if full URL was provided
	// objectKey := s.extractKeyFromURL(key)
	// if objectKey == "" {
	// 	return "", fmt.Errorf("invalid or empty key")
	// }

	// Create presign client
	presignClient := s3.NewPresignClient(s.s3Client)

	// Create GetObject request
	getObjectInput := &s3.GetObjectInput{
		Bucket: aws.String(s.cfg.Bucket),
		Key:    aws.String(key),
	}

	// Generate presigned URL
	presignedReq, err := presignClient.PresignGetObject(ctx, getObjectInput, func(opts *s3.PresignOptions) {
		opts.Expires = expiration
	})

	if err != nil {
		logger.Errorf("Failed to create presigned URL for key %s: %v", key, err)
		return "", fmt.Errorf("failed to create presigned URL: %w", err)
	}

	logger.Infof("Created presigned URL for key %s (expires in %v)", key, expiration)
	return presignedReq.URL, nil
}

// Download fetches file content from S3 and returns the bytes
func (s *Client) Download(ctx context.Context, urlOrKey string) ([]byte, error) {
	logger := logger.From(ctx)

	// Extract key from URL if full URL was provided
	key := s.extractKeyFromURL(urlOrKey)
	if key == "" {
		return nil, fmt.Errorf("invalid or empty file key")
	}

	// Get object from S3
	result, err := s.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.cfg.Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		logger.Errorf("Failed to download file from S3 (key: %s): %v", key, err)
		return nil, fmt.Errorf("failed to download file: %w", err)
	}
	defer result.Body.Close()

	// Read the file content
	data, err := io.ReadAll(result.Body)
	if err != nil {
		logger.Errorf("Failed to read file content (key: %s): %v", key, err)
		return nil, fmt.Errorf("failed to read file content: %w", err)
	}

	logger.Infof("Successfully downloaded file: %s (%d bytes)", key, len(data))
	return data, nil
}
