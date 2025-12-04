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
	"math-ai.com/math-ai/internal/applications/dto"
	"math-ai.com/math-ai/internal/shared/config"
	"math-ai.com/math-ai/internal/shared/logger"
	"math-ai.com/math-ai/internal/shared/utils/filetype"
)

type Client struct {
	s3Client  *s3.Client
	accessKey string
	secretKey string
	Region    string
	Bucket    string
}

// NewS3Client creates a new storage service instance
func NewClient(s3Config *config.S3Config) *Client {
	logger.Info("Initializing S3Client with S3 backend", s3Config.Region, s3Config.Bucket)

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

	return &Client{
		s3Client:  s3Client,
		accessKey: s3Config.AccessKey,
		secretKey: s3Config.SecretKey,
		Region:    s3Config.Region,
		Bucket:    s3Config.Bucket,
	}
}

// Upload uploads a file to S3 and returns URLs
func (s *Client) Upload(ctx context.Context, file io.Reader, filename, contentType, folder string) (*dto.UploadFileResponse, error) {
	// Validate file type
	if err := s.ValidateFileType(filename, contentType); err != nil {
		logger.Errorf("File validation failed: %v", err)
		return nil, fmt.Errorf("invalid file type: %w", err)
	}

	// Generate unique key for the file
	key := s.generateKey(filename, folder)

	// Determine content type
	if contentType == "" {
		contentType = filetype.GetContentTypeFromFilename(filename)
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
		Bucket:      aws.String(s.Bucket),
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

	response := &dto.UploadFileResponse{
		URL:      originalURL,
		Key:      key,
		Filename: filename,
		Size:     fileSize,
	}

	return response, nil
}

// Delete removes a file from S3
func (s *Client) Delete(ctx context.Context, keyUrl string) error {
	// Extract key from URL if full URL was provided
	key := s.extractKeyFromURL(keyUrl)

	if key == "" {
		return fmt.Errorf("invalid or empty file key")
	}

	// Delete from S3
	_, err := s.s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.Bucket),
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
func (s *Client) GetPreviewURL(ctx context.Context, originalURL string) (string, error) {
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
	return filetype.ValidateFile(filename, contentType)
}

// CreatePresignedUrl generates a temporary presigned URL for secure S3 object access
func (s *Client) CreatePresignedUrl(ctx context.Context, key string, expiration time.Duration) (string, error) {
	// Extract key from URL if full URL was provided
	objectKey := s.extractKeyFromURL(key)
	if objectKey == "" {
		return "", fmt.Errorf("invalid or empty key")
	}

	// Create presign client
	presignClient := s3.NewPresignClient(s.s3Client)

	// Create GetObject request
	getObjectInput := &s3.GetObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(objectKey),
	}

	// Generate presigned URL
	presignedReq, err := presignClient.PresignGetObject(ctx, getObjectInput, func(opts *s3.PresignOptions) {
		opts.Expires = expiration
	})

	if err != nil {
		logger.Errorf("Failed to create presigned URL for key %s: %v", objectKey, err)
		return "", fmt.Errorf("failed to create presigned URL: %w", err)
	}

	logger.Infof("Created presigned URL for key %s (expires in %v)", objectKey, expiration)
	return presignedReq.URL, nil
}
