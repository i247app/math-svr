package middleware

import (
	"errors"
	"math"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/infrastructure/logger"
	"math-ai.com/math-ai/internal/infrastructure/metrics"
	"math-ai.com/math-ai/internal/infrastructure/ratelimit"
	sess "math-ai.com/math-ai/internal/infrastructure/session"
	"math-ai.com/math-ai/internal/shared/response"
)

var ErrRateLimited = errors.New("rate limit exceeded")

// RateLimitMode is env RATE_LIMIT_MODE.
type RateLimitMode string

const (
	// RateLimitOff: the middleware is a pass-through.
	RateLimitOff RateLimitMode = "off"
	// RateLimitShadow: decide, log and count, but never block or delay.
	// The default — run it on real traffic to tune limits before enforcing.
	RateLimitShadow RateLimitMode = "shadow"
	// RateLimitEnforce: reject over-limit requests (RATE_LIMIT_EXCEEDED) and
	// apply leaky-bucket delays.
	RateLimitEnforce RateLimitMode = "enforce"
)

// ParseRateLimitMode maps the env value; anything unknown is shadow, so a
// typo can neither block traffic nor silently switch protection off.
func ParseRateLimitMode(s string) (RateLimitMode, bool) {
	switch m := RateLimitMode(strings.ToLower(strings.TrimSpace(s))); m {
	case RateLimitOff, RateLimitShadow, RateLimitEnforce:
		return m, true
	default:
		return RateLimitShadow, false
	}
}

// RateLimitKeyFunc derives the limiter key from a request.
type RateLimitKeyFunc func(r *http.Request) string

// RateLimitKeyIP keys by the real client IP (see ClientIP).
func RateLimitKeyIP(r *http.Request) string {
	return "ip:" + ClientIP(r)
}

// RateLimitKeyUID keys by the session's uid (guests included), falling back
// to the IP when the request carries no session. The session is resolved by
// GexSessionMiddleware, so this only works on per-route limiters.
func RateLimitKeyUID(r *http.Request) string {
	if s := sess.GetRequestSession(r); s != nil {
		if uid, ok := s.UID(); ok && uid > 0 {
			return "uid:" + strconv.FormatInt(uid, 10)
		}
	}
	return RateLimitKeyIP(r)
}

// ClientIP returns the caller's IP for rate limiting. X-Real-IP is trusted
// only when the TCP peer is a private/loopback address — i.e. nginx, which
// overwrites the header with $remote_addr. X-Forwarded-For is never used: our
// nginx APPENDS to it ($proxy_add_x_forwarded_for), so its first entry is
// whatever the client sent. metadata.ip_address is client-reported too.
func ClientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	if ip := net.ParseIP(host); ip != nil && (ip.IsLoopback() || ip.IsPrivate()) {
		if v := strings.TrimSpace(r.Header.Get("X-Real-Ip")); net.ParseIP(v) != nil {
			return v
		}
	}
	return host
}

// RateLimitMiddleware applies one policy. It never wraps the ResponseWriter,
// so it is safe in front of the WebSocket upgrade. A limiter error fails
// open: losing rate limiting briefly beats refusing every request.
func RateLimitMiddleware(policy string, lim ratelimit.Limiter, key RateLimitKeyFunc, mode RateLimitMode, m *metrics.Metrics) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if mode == RateLimitOff {
			return next
		}
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			k := key(r)

			d, err := lim.Allow(ctx, k)
			if err != nil {
				logger.From(ctx).Warnf("ratelimit.error policy=%s err=%v", policy, err)
				next.ServeHTTP(w, r)
				return
			}

			if !d.Allowed {
				retry := int(math.Ceil(d.RetryAfter.Seconds()))
				if mode == RateLimitShadow {
					m.ObserveRateLimit(policy, "shadow_limited")
					logger.From(ctx).Infof("ratelimit.shadow_limited policy=%s key=%s path=%s retry_after=%ds", policy, k, r.URL.Path, retry)
					next.ServeHTTP(w, r)
					return
				}
				m.ObserveRateLimit(policy, "limited")
				logger.From(ctx).Warnf("ratelimit.limited policy=%s key=%s path=%s retry_after=%ds", policy, k, r.URL.Path, retry)
				h := w.Header()
				h.Set("Retry-After", strconv.Itoa(retry))
				h.Set("X-RateLimit-Limit", strconv.Itoa(d.Limit))
				h.Set("X-RateLimit-Remaining", "0")
				// retry_after (seconds) rides in the body too: the mobile client
				// reads the JSON envelope, not headers.
				response.WriteJson(w, map[string]any{"retry_after": retry},
					errs.NewError(ctx, status.RATE_LIMIT_EXCEEDED, nil, ErrRateLimited))
				return
			}

			m.ObserveRateLimit(policy, "allowed")
			if mode == RateLimitEnforce {
				h := w.Header()
				h.Set("X-RateLimit-Limit", strconv.Itoa(d.Limit))
				h.Set("X-RateLimit-Remaining", strconv.Itoa(d.Remaining))
				if d.Delay > 0 { // leaky bucket: wait for this request's slot
					t := time.NewTimer(d.Delay)
					select {
					case <-t.C:
					case <-ctx.Done():
						t.Stop()
						return // client gave up; nothing to answer
					}
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}
