package s3

import (
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/google/uuid"
)

// generateKey creates a unique S3 key for the file
func (s *Client) generateKey(filename string, folder string) string {
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
func (s *Client) generateS3URL(key string) string {
	// Standard S3 URL format
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s.Bucket, s.Region, key)
}

// extractKeyFromURL extracts the S3 object key from a full S3 URL
func (s *Client) extractKeyFromURL(urlOrKey string) string {
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
