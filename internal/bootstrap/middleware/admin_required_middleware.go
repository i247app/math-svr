package middleware

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"errors"
	"net/http"

	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/infrastructure/logger"
	"math-ai.com/math-ai/internal/infrastructure/metadata"
	sess "math-ai.com/math-ai/internal/infrastructure/session"
	sctx "math-ai.com/math-ai/internal/shared/context"
	"math-ai.com/math-ai/internal/shared/enum"
	"math-ai.com/math-ai/internal/shared/response"
)

var (
	ErrAdminRequired   = errors.New("admin role required")
	ErrAdminKeyInvalid = errors.New("admin api key invalid or disabled")
)

// AdminKeyHeader carries the shared admin secret (env ADMIN_API_KEY).
// LogRequestMiddleware redacts it.
const AdminKeyHeader = "X-Api-Admin-Key"

// RoleLookup returns the role currently stored on an account.
type RoleLookup func(ctx context.Context, uid int64) (enum.RoleType, error)

// AdminRequiredMiddleware lets a request through only when its session is
// signed in AND the account's role is ADMIN. It runs AuthRequiredMiddleware
// itself first, so a route needs only this one middleware and the two
// checks can never be registered in the wrong order.
//
// The role is looked up on every request rather than cached in the
// session: an admin who is demoted loses access on their next call.
// A lookup failure refuses the request — never fail open on an admin gate.
func AdminRequiredMiddleware(sessionManager *sess.SessionManager, roleOf RoleLookup) func(http.Handler) http.Handler {
	auth := AuthRequiredMiddleware(sessionManager)
	return func(next http.Handler) http.Handler {
		return auth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			log := logger.From(ctx)

			session := sess.GetRequestSession(r)
			uid, ok := session.UID()
			if !ok || uid == 0 {
				response.WriteJson(w, nil, errs.NewUnauthorizedError(ctx, ErrSessionNotFound))
				return
			}

			role, err := roleOf(ctx, uid)
			if err != nil {
				log.Warnf("admin_gate.lookup_failed uid=%d err=%v", uid, err)
				response.WriteJson(w, nil, err)
				return
			}
			if role != enum.RoleTypeAdmin {
				log.Warnf("admin_gate.denied uid=%d role=%q path=%s", uid, role, r.URL.Path)
				response.WriteJson(w, nil, errs.NewError(ctx, status.FORBIDDEN, nil, ErrAdminRequired))
				return
			}

			next.ServeHTTP(w, r.WithContext(sctx.WithAdminVia(ctx, sctx.AdminViaSession)))
		}))
	}
}

// AdminOrApiKeyMiddleware admits a request that EITHER carries the admin
// API key in X-Api-Admin-Key OR passes AdminRequiredMiddleware. The key is
// what makes the first admin: before one exists, no session can pass the
// role check.
//
//   - No header → the ADMIN-session check decides, unchanged.
//   - Header present → the key alone decides; a wrong key is refused, it
//     does NOT fall back to the session. A caller that sent a key meant
//     the key path, and a silent fallback would hide a bad key.
//   - apiKey "" (env unset) → the key path is closed: any header is
//     refused, so an empty secret can never match an empty header.
//
// Both sides are hashed before the constant-time compare so neither the
// key's content nor its length leaks through response timing. The key is
// never logged — only that one was accepted or refused, and from where.
func AdminOrApiKeyMiddleware(sessionManager *sess.SessionManager, roleOf RoleLookup, apiKey string) func(http.Handler) http.Handler {
	adminOnly := AdminRequiredMiddleware(sessionManager, roleOf)
	want := sha256.Sum256([]byte(apiKey))
	return func(next http.Handler) http.Handler {
		viaSession := adminOnly(next)
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			got := r.Header.Get(AdminKeyHeader)
			if got == "" {
				viaSession.ServeHTTP(w, r)
				return
			}

			ctx := r.Context()
			log := logger.From(ctx)
			gotSum := sha256.Sum256([]byte(got))
			if apiKey == "" || subtle.ConstantTimeCompare(gotSum[:], want[:]) != 1 {
				log.Warnf("admin_key.refused ip=%s path=%s enabled=%t", metadata.GetIPAddress(ctx), r.URL.Path, apiKey != "")
				response.WriteJson(w, nil, errs.NewError(ctx, status.FORBIDDEN, nil, ErrAdminKeyInvalid))
				return
			}

			log.Infof("admin_key.accepted ip=%s path=%s", metadata.GetIPAddress(ctx), r.URL.Path)
			next.ServeHTTP(w, r.WithContext(sctx.WithAdminVia(ctx, sctx.AdminViaAPIKey)))
		})
	}
}
