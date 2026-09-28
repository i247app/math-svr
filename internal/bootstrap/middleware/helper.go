package middleware

import (
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

// tokenTail returns the last tokenTailLen characters of a session token —
// enough to correlate a session across log lines, too little to reuse it.
// Returns "" for a token shorter than that.
func tokenTail(token string) string {
	if len(token) < tokenTailLen {
		return ""
	}
	return token[len(token)-tokenTailLen:]
}
