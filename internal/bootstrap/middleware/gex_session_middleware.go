package middleware

import (
	"bytes"
	"context"
	"net/http"
	"strings"

	"github.com/i247app/gex/sessionprovider"
	"math-ai.com/math-ai/internal/infrastructure/metadata"
	"math-ai.com/math-ai/internal/infrastructure/session"
	sctx "math-ai.com/math-ai/internal/shared/context"
	"math-ai.com/math-ai/internal/shared/response"
)

// GexSessionMiddleware resolves the request's session, binds it to the
// context, returns the session's token in X-Auth-Token, and wraps the
// response writer to capture the response body.
//
// The token comes from the request body's metadata.authorization (parsed by
// MetadataMiddleware, which must run first) and nowhere else: a client's
// Authorization header is never read for a REST request. WebSocket handshakes
// have no body, so they alone use the Authorization header.
//
// When a handler changes the session's auth state (is_secure / uid — login,
// OTP verify, register, logout, …) it asks sessionManager to persist the
// sessions, so the change survives a crash that skips OnShutdown. That is a
// no-op unless persistence is enabled.
func GexSessionMiddleware(
	sessionProvider sessionprovider.SessionProvider,
	sessionContextKey session.SessionRequestContextKey,
	sessionManager *session.SessionManager,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip session handling if this header is set
			if r.Header.Get("X-Skip-Session") == "true" {
				next.ServeHTTP(w, r)
				return
			}

			// WebSocket upgrades still need the session resolved (auth reuses
			// the same Bearer token), but must NOT be wrapped: responseWriter-
			// Wrapper is not hijackable and buffers the response, both of which
			// break the hijack. Resolve, bind to ctx, and pass the original
			// writer straight through — no wrapping, no post-write buffering.
			if isWebSocketUpgrade(r) {
				sessionResult, err := sessionProvider.GetSessionFromRequest(r)
				if err != nil {
					response.WriteJson(w, nil, err)
					return
				}
				if sessionResult == nil || sessionResult.Session == nil {
					next.ServeHTTP(w, r)
					return
				}
				ctx := bindSession(r.Context(), sessionContextKey, sessionResult)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			token := bodyToken(metadata.GetAuthorization(r.Context()))
			sessionResult, err := sessionProvider.GetSessionFromToken(token)
			if err != nil {
				response.WriteJson(w, map[string]string{
					"error":  "gex panic: " + err.Error(),
					"tag":    "sessionProvider.GetSessionFromToken error",
					"origin": "session_middleware",
				}, err)
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
			authToken := sessionResult.AuthToken

			// Wrap the response writer to capture the response body
			wr := &responseWriterWrapper{
				ResponseWriter: w,
				body:           bytes.NewBuffer(nil),
			}

			// X-Auth-Token is the token the client must use from now on; it
			// replaces the old one whenever gex issued a new session.
			if authToken != "" {
				wr.Header().Set("X-Auth-Token", authToken)
			}

			r = r.WithContext(bindSession(r.Context(), sessionContextKey, sessionResult))

			before := authStateOf(sess)
			next.ServeHTTP(wr, r)
			if authStateOf(sess) != before {
				sessionManager.MarkDirty()
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

// authState is the part of a session whose loss on a crash matters: whether it
// is signed in, and as whom.
type authState struct {
	isSecure any
	uid      any
}

func authStateOf(sess interface{ Get(string) (any, bool) }) authState {
	isSecure, _ := sess.Get("is_secure")
	uid, _ := sess.Get("uid")
	return authState{isSecure: isSecure, uid: uid}
}

// bindSession puts the session in the context, plus the tail of its token for
// the logger's line prefix (the full token is never put in the context).
func bindSession(ctx context.Context, key session.SessionRequestContextKey, res *sessionprovider.SessionResult) context.Context {
	ctx = context.WithValue(ctx, key, res.Session)
	if tail := tokenTail(res.AuthToken); tail != "" {
		ctx = sctx.WithTokenSuffix(ctx, tail)
	}
	return ctx
}

// bodyToken turns metadata.authorization ("Bearer <jwt>", or a bare <jwt>)
// into the raw token gex expects.
func bodyToken(authorization string) string {
	authorization = strings.TrimSpace(authorization)
	if token, ok := strings.CutPrefix(authorization, "Bearer "); ok {
		return strings.TrimSpace(token)
	}
	return authorization
}
