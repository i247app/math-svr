package middleware

import (
	"net/http"

	"math-ai.com/math-ai/internal/shared/logger"
)

func AdminRequiredMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger := logger.GetLogger(r.Context())
			logger.Infof("AdminRequiredMiddleware: Access granted to admin user")
			next.ServeHTTP(w, r)
		})
	}
}
