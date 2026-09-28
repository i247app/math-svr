package context

import (
	"context"
)

const (
	ctxKeyUID         contextKey = "math-ai.uid"
	ctxKeyTokenSuffix contextKey = "math-ai.token_suffix"
)

// WithUID binds the authenticated end-user's id to ctx for log
// attribution. The logger middleware is the canonical writer; auth
// middleware (when added) should also set this so background work
// triggered from a request retains uid context.
func WithUID(ctx context.Context, uid int64) context.Context {
	return context.WithValue(ctx, ctxKeyUID, uid)
}

// UID returns the authenticated end-user id bound to ctx, or 0 when
// absent. Logger callers convert 0 to "anon" — UID itself does not
// pre-format.
func UID(ctx context.Context) int64 {
	if ctx == nil {
		return 0
	}
	v, _ := ctx.Value(ctxKeyUID).(int64)
	return v
}

// WithTokenSuffix binds the last few characters of the bearer token to ctx
// for log attribution. The full token is never stored or logged — the suffix
// is enough to correlate log lines with a specific session without enabling
// session theft from log output.
func WithTokenSuffix(ctx context.Context, suffix string) context.Context {
	return context.WithValue(ctx, ctxKeyTokenSuffix, suffix)
}

// TokenSuffix returns the bearer-token tail bound to ctx, or "" when absent.
// Logger callers convert "" to "anon".
func TokenSuffix(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	v, _ := ctx.Value(ctxKeyTokenSuffix).(string)
	return v
}
