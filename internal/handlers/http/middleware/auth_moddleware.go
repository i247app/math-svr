package middleware

import (
	"fmt"
	"log"
	"net/http"

	"math-ai.com/math-ai/internal/session"
	"math-ai.com/math-ai/internal/shared/constant/status"
	"math-ai.com/math-ai/internal/shared/logger"
	"math-ai.com/math-ai/internal/shared/utils/response"
)

func AuthRequiredMiddleware(sessionManager *session.SessionManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger := logger.GetLogger(r.Context())
			logger.Infof("AuthRequiredMiddleware: Access granted to admin user")

			// Check if session exists
			session := session.GetRequestSession(r)
			if session == nil {
				response.WriteJson(w, r.Context(), nil, fmt.Errorf("session not found"), status.UNAUTHORIZED)
				return
			}
			sessionKey, _ := session.Get("key") // retrieve the session key to log during errors

			// Check for is_secure as a sanity check
			isSecure, ok := session.Get("is_secure")
			if !ok {
				log.Printf("- session key: %s\n", sessionKey)
				response.WriteJson(w, r.Context(), nil, fmt.Errorf("session is_secure is missing"), status.UNAUTHORIZED)
				return
			}

			// Check for is_secure == false
			if !isSecure.(bool) {
				log.Printf("- session key: %s\n", sessionKey)
				response.WriteJson(w, r.Context(), nil, fmt.Errorf("session is not secure, login required"), status.UNAUTHORIZED)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
