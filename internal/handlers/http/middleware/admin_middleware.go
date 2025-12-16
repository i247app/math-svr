package middleware

import (
	"fmt"
	"net/http"

	di "math-ai.com/math-ai/internal/core/di/services"
	"math-ai.com/math-ai/internal/session"
	"math-ai.com/math-ai/internal/shared/constant/status"
	"math-ai.com/math-ai/internal/shared/logger"
	"math-ai.com/math-ai/internal/shared/utils/response"
)

func AdminRequiredMiddleware(sessionManager *session.SessionManager, userSvc di.IUserService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger := logger.GetLogger(r.Context())

			ctx := r.Context()

			// Check if session exists
			session := session.GetRequestSession(r)
			if session == nil {
				logger.Error("AdminRequiredMiddleware: session not found")
				response.WriteJson(w, ctx, nil, fmt.Errorf("session not found"), status.UNAUTHORIZED)
				return
			}

			// Get uid from session
			uid, ok := session.UID()
			if !ok {
				logger.Error("AdminRequiredMiddleware: session not found")
				response.WriteJson(w, ctx, nil, fmt.Errorf("session not found"), status.UNAUTHORIZED)
				return
			}

			// Check if currentUser is admin
			_, user, err := userSvc.GetUserByID(r.Context(), uid)
			if err != nil {
				logger.Error("AdminRequiredMiddleware: user not found")
				response.WriteJson(w, ctx, nil, fmt.Errorf("user not found"), status.UNAUTHORIZED)
				return
			}

			if user.Role != "admin" {
				logger.Error("AdminRequiredMiddleware: user is not admin")
				response.WriteJson(w, ctx, nil, fmt.Errorf("user is not admin"), status.UNAUTHORIZED)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
