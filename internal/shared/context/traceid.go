package context

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"strings"
)

const (
	ctxKeyTraceID contextKey = "math-ai.trace_id"
)

const (
	traceIDMinLen = 8
	traceIDMaxLen = 64
)

// WithTraceID binds a normalized trace id to ctx. Empty or invalid input is
// replaced with a freshly generated id so downstream code can always rely on
// TraceID returning a non-empty value.
func WithTraceID(ctx context.Context, id string) context.Context {
	normalized := NormalizeTraceID(id)
	if normalized == "" {
		normalized = NewTraceID()
	}
	return context.WithValue(ctx, ctxKeyTraceID, normalized)
}

// TraceID returns the trace id bound to ctx, or an empty string if none is set.
// Callers that need a guaranteed non-empty id should use EnsureTraceID.
func TraceID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	v, _ := ctx.Value(ctxKeyTraceID).(string)
	return v
}

// EnsureTraceID returns the trace id from ctx, generating and binding a new
// one when absent. It returns the (possibly updated) context alongside the id.
func EnsureTraceID(ctx context.Context) (context.Context, string) {
	if id := TraceID(ctx); id != "" {
		return ctx, id
	}
	id := NewTraceID()
	return context.WithValue(ctx, ctxKeyTraceID, id), id
}

// NormalizeTraceID keeps only characters in [A-Za-z0-9_-] and trims the result
// to TraceIDMaxLen. It returns "" when the normalized id is shorter than
// TraceIDMinLen so callers can detect unusable input.
func NormalizeTraceID(raw string) string {
	if raw == "" {
		return ""
	}
	var b strings.Builder
	b.Grow(len(raw))
	for _, r := range raw {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
		case r >= 'A' && r <= 'Z':
			b.WriteRune(r)
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '_' || r == '-':
			b.WriteRune(r)
		}
		if b.Len() >= traceIDMaxLen {
			break
		}
	}
	if b.Len() < traceIDMinLen {
		return ""
	}
	return b.String()
}

// NewTraceID returns a fresh 16-byte random id encoded as 32 hex chars.
func NewTraceID() string {
	var buf [16]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return ""
	}
	return hex.EncodeToString(buf[:])
}
