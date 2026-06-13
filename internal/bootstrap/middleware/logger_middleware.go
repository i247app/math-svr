package middleware

import (
	"net/http"
	"strings"

	"math-ai.com/math-ai/internal/application/resource"
	sctx "math-ai.com/math-ai/internal/shared/context"

	"math-ai.com/math-ai/internal/infrastructure/logger"
)

const (
	authHeader   = "Authorization"
	bearerPrefix = "Bearer "
	tokenTailLen = 6
)

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

			// Set token suffix for logger
			if tail := bearerTail(r); tail != "" {
				ctx = sctx.WithTokenSuffix(ctx, tail)
			}

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

// bearerTail extracts the last tokenTailLen characters of the bearer
// token from the Authorization header. Returns "" when the header is
// missing, malformed, or shorter than tokenTailLen.
func bearerTail(r *http.Request) string {
	raw := r.Header.Get(authHeader)
	if !strings.HasPrefix(raw, bearerPrefix) {
		return ""
	}
	tok := strings.TrimPrefix(raw, bearerPrefix)
	if len(tok) < tokenTailLen {
		return ""
	}
	return tok[len(tok)-tokenTailLen:]
}
