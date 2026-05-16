package middleware

import (
	"compress/gzip"
	"net/http"
	"strings"
)

// GzipMiddleware compresses HTTP responses for clients that support it.
func GzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip compression if the client doesn't accept gzip
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		// Create a gzip writer
		gz := gzip.NewWriter(w)
		defer gz.Close()

		// Set headers
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Add("Vary", "Accept-Encoding") // Helps caches understand the response varies by encoding

		// Wrap the response writer
		gzipResponseWriter := &gzipResponseWriter{
			Writer:         gz,
			ResponseWriter: w,
		}

		next.ServeHTTP(gzipResponseWriter, r)
	})
}

// gzipResponseWriter wraps http.ResponseWriter to transparently compress output.
type gzipResponseWriter struct {
	*gzip.Writer
	http.ResponseWriter
	statusCode int // Add field to store status code
}

func (w *gzipResponseWriter) Header() http.Header {
	return w.ResponseWriter.Header()
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func (w *gzipResponseWriter) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
	w.statusCode = statusCode
}
