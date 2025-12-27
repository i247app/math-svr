package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"math-ai.com/math-ai/internal/shared/logger"
	"math-ai.com/math-ai/internal/shared/metadata"
)

// MetadataMiddleware extracts __metadata from request body and injects it into context
// This middleware should be applied to all POST/PUT/PATCH endpoints that accept JSON
func MetadataMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// // Only process JSON requests with a body
			// if r.Body == nil || r.Method == http.MethodGet || r.Method == http.MethodDelete {
			// 	next.ServeHTTP(w, r)
			// 	return
			// }

			contentType := r.Header.Get("Content-Type")
			// Only process application/json content type
			if contentType != "application/json" {
				// For multipart/form-data or other types, skip metadata extraction
				next.ServeHTTP(w, r)
				return
			}

			// Read the entire request body
			bodyBytes, err := io.ReadAll(r.Body)
			if err != nil {
				logger.Error("Failed to read request body: %v", err)
				next.ServeHTTP(w, r)
				return
			}
			defer r.Body.Close()

			// Try to parse as JSON to extract __metadata
			var rawBody map[string]interface{}
			if err := json.Unmarshal(bodyBytes, &rawBody); err != nil {
				// If not valid JSON, just restore the body and continue
				r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
				next.ServeHTTP(w, r)
				return
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
				delete(rawBody, "__metadata")

				// Re-serialize the body without __metadata
				cleanBodyBytes, err := json.Marshal(rawBody)
				if err != nil {
					logger.Error("Failed to re-serialize request body: %v", err)
					r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
					next.ServeHTTP(w, r)
					return
				}
				bodyBytes = cleanBodyBytes
			} else {
				// No __metadata found, create empty metadata
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

			// Create new request with updated context and clean body
			r = r.WithContext(ctx)
			r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

			// Continue to next handler
			next.ServeHTTP(w, r)
		})
	}
}
