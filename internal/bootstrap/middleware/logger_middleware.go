package middleware

import (
	"net/http"

	"math-ai.com/math-ai/internal/application/resource"
	sctx "math-ai.com/math-ai/internal/shared/context"

	"math-ai.com/math-ai/internal/infrastructure/logger"
)

const tokenTailLen = 6

// LoggerMiddleware constructs a per-request AppLogger via the Provider and
// binds it to the request context. Handlers and services pull it back out
// with logger.From(ctx). The middleware runs before any auth middleware;
// if auth later validates the bearer token and resolves the user, it
// should also call kctx.WithUserID so subsequent log lines carry [uid].
//
// The full bearer token is never stored or logged — only its last
// tokenTailLen characters land in the context, enough to correlate a
// session in logs without being usable for impersonation if the log file
// leaks.
func LoggerMiddleware(p *logger.Provider, res *resource.Resource) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// Assign a monotonic per-request id (atomic, race-free) so every
			// log line emitted while handling this request carries the same
			// [ReqID: N] tag. This is what lets an operator gather the lines of
			// one request out of the interleaved output of many concurrent
			// requests. Must be set before p.New so the logger binds it.
			ctx = sctx.WithRequestID(ctx, sctx.NextRequestID())

			// The token suffix for the line prefix is already in ctx:
			// GexSessionMiddleware put it there with the session.

			// Set user ID for logger
			if uid, err := res.GetRequestUID(r); err == nil {
				ctx = sctx.WithUserID(ctx, uid)
			}

			lg := p.New(ctx, r)
			ctx = logger.Inject(ctx, lg)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
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
