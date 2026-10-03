package middleware

import (
	"context"
	"errors"
	"net/http"

	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/infrastructure/logger"
	sess "math-ai.com/math-ai/internal/infrastructure/session"
	"math-ai.com/math-ai/internal/shared/enum"
	"math-ai.com/math-ai/internal/shared/response"
)

var ErrAdminRequired = errors.New("admin role required")

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

			next.ServeHTTP(w, r)
		}))
	}
}
