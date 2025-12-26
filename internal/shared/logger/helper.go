package logger

import (
	"context"
	"fmt"
	"net/http"

	"math-ai.com/math-ai/internal/session"
)

// extractSessionInfoFromRequest extracts token and userid from the request
func extractSessionInfoFromRequest(r *http.Request) (token string, userid string, route string) {
	token = "anon"
	userid = "anon"
	route = "GET /unknown"

	if r == nil {
		return
	}

	// Get session from request context
	sess := session.GetRequestSession(r)
	if sess == nil {
		return
	}

	// Extract token (key) from session
	if key, ok := sess.Get("key"); ok {
		if keyStr, ok := key.(string); ok && len(keyStr) >= 6 {
			if len(keyStr) >= 6 {
				token = keyStr[len(keyStr)-6:]
			} else {
				token = keyStr
			}
		}
	}

	// Extract userid from session
	if uid, ok := sess.UID(); ok {
		userid = uid
	}

	// Extract route from context if available
	route = fmt.Sprintf("%s %s", r.Method, r.URL.Path)

	return
}

// withSessionInfo adds session info (token, userid) to context for logging
func withSessionInfo(ctx context.Context, token, userid, route string) context.Context {
	ctx = context.WithValue(ctx, tokenKey, token)
	ctx = context.WithValue(ctx, useridKey, userid)
	ctx = context.WithValue(ctx, routeKey, route)
	return ctx
}

// extractSessionInfo retrieves token and userid from context
func extractSessionInfo(ctx context.Context) (token, userid, route string) {
	token = "anon"
	userid = "anon"

	if t, ok := ctx.Value(tokenKey).(string); ok && t != "" {
		token = t
	}
	if u, ok := ctx.Value(useridKey).(string); ok && u != "" {
		userid = u
	}
	if r, ok := ctx.Value(routeKey).(string); ok && r != "" {
		route = r
	}

	return
}
