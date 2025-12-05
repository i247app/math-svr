package middleware

import (
	"fmt"
	"net/http"

	"math-ai.com/math-ai/internal/session"
	"math-ai.com/math-ai/internal/shared/constant/status"
	"math-ai.com/math-ai/internal/shared/utils/response"
)

func ValidateSessionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session := session.GetRequestSession(r)
		if session == nil {
			response.WriteJson(w, r.Context(), nil, fmt.Errorf("ValidateSessionMiddleware: session not found"), status.UNAUTHORIZED)
			return
		}

		switch {
		case session.IsExpired():
			response.WriteJson(w, r.Context(), nil, fmt.Errorf("ValidateSessionMiddleware: session expired"), status.UNAUTHORIZED)
			return
		case session.IsMarkedForDeletion():
			response.WriteJson(w, r.Context(), nil, fmt.Errorf("ValidateSessionMiddleware: session marked for deletion"), status.UNAUTHORIZED)
			return
		case !session.IsValid():
			response.WriteJson(w, r.Context(), nil, fmt.Errorf("ValidateSessionMiddleware: session is invalid"), status.UNAUTHORIZED)
			return
		}

		next.ServeHTTP(w, r)
	})
}
