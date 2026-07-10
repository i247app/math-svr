package middleware

import (
	"errors"
	"net/http"

	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/infrastructure/logger"
	"math-ai.com/math-ai/internal/infrastructure/session"
	"math-ai.com/math-ai/internal/shared/response"
)

var (
	ErrSessionNotFound                 = errors.New("session not found")
	ErrSessionIsSecureMissing          = errors.New("session is_secure is missing")
	ErrSessionIsNotSecureLoginRequired = errors.New("session is not secure, login required")
)

func AuthRequiredMiddleware(sessionManager *session.SessionManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			log := logger.From(ctx)

			// Check if session exists
			session := session.GetRequestSession(r)
			if session == nil {
				response.WriteJson(w, nil, errs.NewUnauthorizedError(ctx, ErrSessionNotFound))
				return
			}
			sessionKey, _ := session.Get("key") // retrieve the session key to log during errors

			log.Infof("sessionKey: %s", sessionKey)

			// Check for is_secure as a sanity check
			isSecure, ok := session.Get("is_secure")
			if !ok {
				log.Warnf("- session key: %s\n", sessionKey)
				response.WriteJson(w, nil, errs.NewUnauthorizedError(ctx, ErrSessionIsSecureMissing))
				return
			}

			// Check for is_secure == false
			if !isSecure.(bool) {
				log.Warnf("- session key: %s\n", sessionKey)
				response.WriteJson(w, nil, errs.NewUnauthorizedError(ctx, ErrSessionIsNotSecureLoginRequired))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
