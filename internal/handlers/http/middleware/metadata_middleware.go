package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"math-ai.com/math-ai/internal/shared/logger"
	"math-ai.com/math-ai/internal/shared/metadata"
)

const (
	MaxMetadataSize = 10 << 20 // 10 MB for multipart form parsing
)

// MetadataMiddleware extracts __metadata from request and injects it into context
// Supports both application/json and multipart/form-data content types
func MetadataMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			contentType := r.Header.Get("Content-Type")

			var requestMetadata *metadata.RequestMetadata

			// Determine content type and extract metadata accordingly
			if strings.HasPrefix(contentType, "application/json") {
				// Handle JSON requests
				requestMetadata = extractMetadataFromJSON(r)
			} else if strings.HasPrefix(contentType, "multipart/form-data") {
				// Handle multipart/form-data requests (file uploads)
				requestMetadata = extractMetadataFromMultipart(r)
			} else {
				// For other content types, create empty metadata
				requestMetadata = metadata.NewRequestMetadata()
			}

			// Log metadata extraction (for debugging)
			if requestMetadata.TraceID != "" {
				logger.Debug("Extracted metadata - TraceID: %s, Platform: %s, Locale: %s",
					requestMetadata.TraceID,
					requestMetadata.ClientInfo.Platform,
					requestMetadata.UserContext.Locale)
			}

			// Inject metadata into context
			ctx := metadata.WithMetadata(r.Context(), requestMetadata)

			// Create new request with updated context
			r = r.WithContext(ctx)

			// Continue to next handler
			next.ServeHTTP(w, r)
		})
	}
}

// extractMetadataFromJSON extracts __metadata from JSON request body
func extractMetadataFromJSON(r *http.Request) *metadata.RequestMetadata {
	// Read the entire request body
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Error("Failed to read request body: %v", err)
		return metadata.NewRequestMetadata()
	}
	defer r.Body.Close()

	// Try to parse as JSON to extract __metadata
	var rawBody map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &rawBody); err != nil {
		// If not valid JSON, just restore the body and return empty metadata
		r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		return metadata.NewRequestMetadata()
	}

	// Extract __metadata if present
	var requestMetadata *metadata.RequestMetadata
	if metadataRaw, exists := rawBody["__metadata"]; exists {
		// Parse __metadata into struct
		metadataBytes, err := json.Marshal(metadataRaw)
		if err == nil {
			requestMetadata = metadata.NewRequestMetadata()
			if err := json.Unmarshal(metadataBytes, requestMetadata); err != nil {
				logger.Warn("Failed to parse __metadata: %v", err)
				requestMetadata = metadata.NewRequestMetadata()
			}
		}

		// Remove __metadata from the body so DTOs can be parsed normally
		// delete(rawBody, "__metadata")

		// Re-serialize the body without __metadata
		cleanBodyBytes, err := json.Marshal(rawBody)
		if err != nil {
			logger.Error("Failed to re-serialize request body: %v", err)
			r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			return requestMetadata
		}
		bodyBytes = cleanBodyBytes
	} else {
		// No __metadata found, create empty metadata
		requestMetadata = metadata.NewRequestMetadata()
	}

	// Restore the body (cleaned of __metadata)
	r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	return requestMetadata
}

// extractMetadataFromMultipart extracts __metadata from multipart/form-data request
// The client should send metadata as a form field named "__metadata" with JSON string value
func extractMetadataFromMultipart(r *http.Request) *metadata.RequestMetadata {
	// Parse multipart form (with size limit)
	err := r.ParseMultipartForm(MaxMetadataSize)
	if err != nil {
		logger.Warn("Failed to parse multipart form: %v", err)
		return metadata.NewRequestMetadata()
	}

	// Look for __metadata form field
	metadataJSON := r.FormValue("__metadata")
	if metadataJSON == "" {
		// No metadata provided, return empty
		return metadata.NewRequestMetadata()
	}

	// Parse the JSON string into RequestMetadata
	requestMetadata := metadata.NewRequestMetadata()
	if err := json.Unmarshal([]byte(metadataJSON), requestMetadata); err != nil {
		logger.Warn("Failed to parse __metadata from multipart form: %v", err)
		return metadata.NewRequestMetadata()
	}

	// Note: We don't remove the __metadata field from the form
	// because controllers will parse the form themselves and won't be affected by it
	// (they only read specific fields they need)

	return requestMetadata
}
