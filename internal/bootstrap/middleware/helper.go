package middleware

import (
	"bytes"
	"net/http"
	"strings"
)

// isWebSocketUpgrade reports whether r is a WebSocket handshake. Middleware that
// wraps or buffers the ResponseWriter must bypass such requests: Accept hijacks
// the connection, so any wrapper lacking Hijack/Unwrap breaks the upgrade, and
// response buffering would try to capture an unbounded, long-lived stream.
func isWebSocketUpgrade(r *http.Request) bool {
	return strings.EqualFold(r.Header.Get("Upgrade"), "websocket") &&
		strings.Contains(strings.ToLower(r.Header.Get("Connection")), "upgrade")
}

// responseWriterWrapper wraps the http.ResponseWriter to capture the response body.
type responseWriterWrapper struct {
	http.ResponseWriter
	body       *bytes.Buffer
	statusCode int // Add field to store status code
}

func (w *responseWriterWrapper) Write(b []byte) (int, error) {
	return w.body.Write(b)
}

func (w *responseWriterWrapper) Header() http.Header {
	return w.ResponseWriter.Header()
}

// Implement WriteHeader to capture the status code
func (w *responseWriterWrapper) WriteHeader(statusCode int) {
	w.statusCode = statusCode
}
