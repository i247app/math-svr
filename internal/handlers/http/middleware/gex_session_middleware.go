package middleware

import (
	"bytes"
	"context"
	"net/http"

	"github.com/i247app/gex/sessionprovider"
	"math-ai.com/math-ai/internal/session"
	"math-ai.com/math-ai/internal/shared/constant/status"
	"math-ai.com/math-ai/internal/shared/utils/response"
)

// GexSessionMiddleware is a unified middleware that handles session management using a SessionProvider.
// It retrieves the session from the provider, handles auto-refresh notifications,
// and wraps the response writer to capture the response body.
func GexSessionMiddleware(
	sessionProvider sessionprovider.SessionProvider,
	sessionContextKey session.SessionRequestContextKey,
) func(http.Handler) http.Handler {

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			// Skip session handling if this header is set
			if r.Header.Get("X-Skip-Session") == "true" {
				next.ServeHTTP(w, r)
				return
			}

			// Get session with metadata from the provider
			sessionResult, err := sessionProvider.GetSessionFromRequest(r)
			if err != nil {
				response.WriteJson(w, r.Context(), map[string]string{
					"error":  "gex panic: " + err.Error(),
					"tag":    "sessionProvider.GetSessionFromRequest error",
					"origin": "session_middleware",
				}, err, status.FAIL)
				return
			}
			if sessionResult == nil || sessionResult.Session == nil {
				// If session is nil proceed without session
				next.ServeHTTP(w, r)
				return
			}

			///////////////////////
			// IS THIS DANGEROUS //
			///////////////////////

			sess := sessionResult.Session
			didAutoRefresh := sessionResult.DidAutoRefresh
			authToken := sessionResult.AuthToken

			// Wrap the response writer to capture the response body
			wr := &responseWriterWrapper{
				ResponseWriter: w,
				body:           bytes.NewBuffer(nil),
			}

			// Set auth token in request header for downstream handlers
			if r.Header.Get("Authorization") == "" {
				r.Header.Add("Authorization", "Bearer "+authToken)
			}

			// Set X-Auth-Token response header
			if authToken != "" {
				wr.Header().Set("X-Auth-Token", authToken)
			}

			// Add session to request context
			r = r.WithContext(context.WithValue(r.Context(), sessionContextKey, sess))

			next.ServeHTTP(wr, r)

			// Notify the client that the session was auto-refreshed
			if didAutoRefresh {
				w.Header().Add("GEX-Session-Auto-Refreshed", "true")
			}

			///////////////////////
			// IS THIS DANGEROUS //
			///////////////////////

			if wr.statusCode != 0 {
				w.WriteHeader(wr.statusCode)
			}
			w.Write(wr.body.Bytes())
		})
	}
}
