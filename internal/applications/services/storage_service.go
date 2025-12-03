package services

import (
	"context"
	"fmt"
	"io"
	"path"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"

	"math-ai.com/math-ai/internal/applications/dto"
	di "math-ai.com/math-ai/internal/core/di/services"
	"math-ai.com/math-ai/internal/shared/config"
	"math-ai.com/math-ai/internal/shared/constant/status"
	"math-ai.com/math-ai/internal/shared/logger"
	"math-ai.com/math-ai/internal/shared/utils/filetype"
)

// StorageService implements the storage service for S3
type StorageService struct {
	s3Client *s3.Client
	config   *config.S3Config
}

// NewStorageService creates a new storage service instance
func NewStorageService(s3Config *config.S3Config) di.IStorageService {
	logger.Info("Initializing StorageService with S3 backend", s3Config.Region, s3Config.Bucket)

	// Create AWS config with static credentials
	awsConfig := aws.Config{
		Region: s3Config.Region,
		Credentials: credentials.NewStaticCredentialsProvider(
			s3Config.AccessKey,
			s3Config.SecretKey,
			"",
		),
	}

	// Create S3 client
	s3Client := s3.NewFromConfig(awsConfig)

	return &StorageService{
		s3Client: s3Client,
		config:   s3Config,
	}
}

// Upload uploads a file to S3 and returns URLs
func (s *StorageService) Upload(ctx context.Context, req *dto.UploadFileRequest) (status.Code, *dto.UploadFileResponse, error) {
	// Validate file type
	if err := s.ValidateFileType(req.Filename, req.ContentType); err != nil {
		logger.Errorf("File validation failed: %v", err)
		return status.BAD_REQUEST, nil, fmt.Errorf("invalid file type: %w", err)
	}

	// Generate unique key for the file
	key := s.generateKey(req.Filename, req.Folder)

	// Determine content type
	contentType := req.ContentType
	if contentType == "" {
		contentType = filetype.GetContentTypeFromFilename(req.Filename)
	}

	// Read the file into memory (for getting size)
	// Note: For large files, consider using multipart upload
	fileBytes, err := io.ReadAll(req.File)
	if err != nil {
		logger.Errorf("Failed to read file: %v", err)
		return status.FAIL, nil, fmt.Errorf("failed to read file: %w", err)
	}

	fileSize := int64(len(fileBytes))

	// Upload to S3
	_, err = s.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.config.Bucket),
		Key:         aws.String(key),
		Body:        strings.NewReader(string(fileBytes)),
		ContentType: aws.String(contentType),
	})

	if err != nil {
		logger.Errorf("Failed to upload file to S3: %v", err)
		return status.FAIL, nil, fmt.Errorf("failed to upload file: %w", err)
	}

	// Generate URLs
	originalURL := s.generateS3URL(key)
	previewURL := s.generatePreviewURL(key)

	logger.Infof("Successfully uploaded file: %s (key: %s, size: %d bytes)", req.Filename, key, fileSize)

	response := &dto.UploadFileResponse{
		URL:        originalURL,
		PreviewURL: previewURL,
		Key:        key,
		Filename:   req.Filename,
		Size:       fileSize,
	}

	return status.OK, response, nil
}

// Delete removes a file from S3
func (s *StorageService) Delete(ctx context.Context, req *dto.DeleteFileRequest) (status.Code, error) {
	// Extract key from URL if full URL was provided
	key := s.extractKeyFromURL(req.Key)

	if key == "" {
		return status.BAD_REQUEST, fmt.Errorf("invalid or empty file key")
	}

	// Delete from S3
	_, err := s.s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.config.Bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		logger.Errorf("Failed to delete file from S3 (key: %s): %v", key, err)
		return status.FAIL, fmt.Errorf("failed to delete file: %w", err)
	}

	logger.Infof("Successfully deleted file: %s", key)
	return status.OK, nil
}

// GetPreviewURL generates a preview URL from an original S3 URL
func (s *StorageService) GetPreviewURL(ctx context.Context, originalURL string) (string, error) {
	key := s.extractKeyFromURL(originalURL)
	if key == "" {
		return "", fmt.Errorf("invalid URL: cannot extract key")
	}

	return s.generatePreviewURL(key), nil
}

// ValidateFileType checks if the file type is allowed
func (s *StorageService) ValidateFileType(filename string, contentType string) error {
	return filetype.ValidateFile(filename, contentType)
}

// generateKey creates a unique S3 key for the file
func (s *StorageService) generateKey(filename string, folder string) string {
	// Generate unique ID
	uniqueID := uuid.New().String()

	// Get file extension
	ext := path.Ext(filename)

	// Build key with optional folder prefix
	if folder != "" {
		folder = strings.Trim(folder, "/")
		return fmt.Sprintf("%s/%s-%s%s", folder, time.Now().Format("20060102"), uniqueID, ext)
	}

	return fmt.Sprintf("%s-%s%s", time.Now().Format("20060102"), uniqueID, ext)
}

// generateS3URL creates the full S3 URL for a key
func (s *StorageService) generateS3URL(key string) string {
	// Standard S3 URL format
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s.config.Bucket, s.config.Region, key)
}

// generatePreviewURL creates a preview URL from a key
// This extracts the key and creates a clean preview URL for UI display
func (s *StorageService) generatePreviewURL(key string) string {
	// For preview, we use a cleaner format that can be easily parsed
	// Format: https://{bucket}.s3.{region}.amazonaws.com/preview/{key}
	// The "preview" path segment makes it clear this is for UI display
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/preview/%s", s.config.Bucket, s.config.Region, key)
}

// extractKeyFromURL extracts the S3 object key from a full S3 URL
func (s *StorageService) extractKeyFromURL(urlOrKey string) string {
	// If it doesn't look like a URL, assume it's already a key
	if !strings.Contains(urlOrKey, "://") {
		return urlOrKey
	}

	// Parse S3 URL formats:
	// 1. https://bucket.s3.region.amazonaws.com/key
	// 2. https://bucket.s3.region.amazonaws.com/preview/key
	// 3. https://s3.region.amazonaws.com/bucket/key

	// Remove protocol
	withoutProtocol := strings.TrimPrefix(urlOrKey, "https://")
	withoutProtocol = strings.TrimPrefix(withoutProtocol, "http://")

	// Split by /
	parts := strings.SplitN(withoutProtocol, "/", 2)
	if len(parts) < 2 {
		return ""
	}

	key := parts[1]

	// Remove "preview/" prefix if present
	key = strings.TrimPrefix(key, "preview/")

	return key
}
