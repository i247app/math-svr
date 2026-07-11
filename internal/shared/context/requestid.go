package context

import (
	"context"
	"sync/atomic"
)

const (
	ctxKeyRequestID contextKey = "math-ai.request_id"
)

// requestIDSeq is the process-wide monotonic source for per-request ids. It
// is advanced exactly once per inbound HTTP request (by LoggerMiddleware via
// NextRequestID) with a single atomic add. Because atomic.Add both mutates
// and returns the new value in one indivisible step, concurrent requests can
// never be handed the same id and none is ever skipped — 10 simultaneous
// requests always receive 10 distinct ids.
//
// The counter resets to 0 on process restart: request ids only need to
// disambiguate and group log lines within a single server lifetime, not be
// globally unique across restarts.
var requestIDSeq atomic.Uint64

// NextRequestID atomically returns the next request id, starting at 1 for the
// first request after boot. Lock-free and safe for concurrent callers: each
// caller observes a distinct, gap-free value.
func NextRequestID() uint64 {
	return requestIDSeq.Add(1)
}

// WithRequestID binds a per-request sequence id to ctx for log attribution.
// LoggerMiddleware is the canonical writer; the logger handlers read it back
// via RequestID so every line emitted while handling one request shares the
// same [Req: N] tag, making interleaved concurrent requests easy to separate
// in the log stream.
func WithRequestID(ctx context.Context, id uint64) context.Context {
	return context.WithValue(ctx, ctxKeyRequestID, id)
}

// RequestID returns the per-request sequence id bound to ctx, or 0 when
// absent (background loggers, or code paths outside the HTTP middleware
// chain). Logger callers render 0 as "-".
func RequestID(ctx context.Context) uint64 {
	if ctx == nil {
		return 0
	}
	v, _ := ctx.Value(ctxKeyRequestID).(uint64)
	return v
}
