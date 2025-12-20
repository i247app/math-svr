package middleware

import (
	"fmt"
	"net/http"

	diSvc "math-ai.com/math-ai/internal/core/di/services"
	"math-ai.com/math-ai/internal/session"
	"math-ai.com/math-ai/internal/shared/constant/status"
	"math-ai.com/math-ai/internal/shared/utils/response"
)

// PermissionRequiredMiddleware checks if the user has permission to access the endpoint
func PermissionRequiredMiddleware(sessionManager *session.SessionManager, authSvc diSvc.IAuthorizationService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// Get session
			sess := session.GetRequestSession(r)
			if sess == nil {
				response.WriteJson(w, ctx, nil, fmt.Errorf("unauthorized: no session"), status.UNAUTHORIZED)
				return
			}

			// Get user ID from session
			userID, ok := sess.UID()
			if !ok || userID == "" {
				response.WriteJson(w, ctx, nil, fmt.Errorf("unauthorized: invalid session"), status.UNAUTHORIZED)
				return
			}

			// Check if user has permission for this endpoint
			hasPermission, err := authSvc.HasPermission(ctx, userID, r.Method, r.URL.Path)
			if err != nil {
				response.WriteJson(w, ctx, nil, fmt.Errorf("error checking permissions: %v", err), status.UNAUTHORIZED)
				return
			}

			if !hasPermission {
				response.WriteJson(w, ctx, nil, fmt.Errorf("forbidden: insufficient permissions"), status.UNAUTHORIZED)
				return
			}

			// User has permission, continue
			next.ServeHTTP(w, r)
		})
	}
}
