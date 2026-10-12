package middleware

import (
	"context"
	"net/http"
	"time"

	"math-ai.com/math-ai/internal/infrastructure/logger"
	"math-ai.com/math-ai/internal/infrastructure/metrics"
	"math-ai.com/math-ai/internal/infrastructure/ratelimit"
)

// Rate-limit policies — the one place to change an algorithm or a number.
// The numbers are starting points: run RATE_LIMIT_MODE=shadow, read
// ratelimit_decisions_total{result="shadow_limited"}, then tune and enforce.
//
// Per-IP limits are deliberately generous: Vietnamese mobile carriers use
// CGNAT and a school's Wi-Fi puts a whole class behind one address.
var (
	// L1 — every request, keyed by IP, before the body is even parsed.
	globalPolicy = ratelimit.Config{Algorithm: ratelimit.FixedWindow, Limit: 600, Window: time.Minute}

	// L2 — public sign-in / sign-up / OTP routes, keyed by IP. Small limit,
	// so the exact sliding log is affordable.
	authPolicy = ratelimit.Config{Algorithm: ratelimit.SlidingWindowLog, Limit: 20, Window: time.Minute}

	// L2 — routes that pay for a model call, keyed by uid (IP without a
	// session). A short burst, then one every 3s.
	llmPolicy = ratelimit.Config{Algorithm: ratelimit.TokenBucket, Limit: 20, Window: time.Minute, Burst: 5}

	// L2 — sending chat messages, keyed by uid: one per second, up to 5
	// queued (a queued message waits for its slot instead of failing).
	chatPolicy = ratelimit.Config{Algorithm: ratelimit.LeakyBucket, Limit: 60, Window: time.Minute, Burst: 5}
)

// RouteRateLimits are the per-route (L2) middlewares for reg(...).
type RouteRateLimits struct {
	Auth func(http.Handler) http.Handler
	LLM  func(http.Handler) http.Handler
	Chat func(http.Handler) http.Handler
}

// GlobalRateLimit is the L1 middleware for the server-wide chain.
func GlobalRateLimit(mode string, m *metrics.Metrics) func(http.Handler) http.Handler {
	return RateLimitMiddleware("global", mustLimiter(globalPolicy), RateLimitKeyIP, parseMode(mode), m)
}

// NewRouteRateLimits builds the L2 middlewares. One limiter per policy: all
// routes sharing a policy share its counter per key.
func NewRouteRateLimits(mode string, m *metrics.Metrics) RouteRateLimits {
	md := parseMode(mode)
	return RouteRateLimits{
		Auth: RateLimitMiddleware("auth", mustLimiter(authPolicy), RateLimitKeyIP, md, m),
		LLM:  RateLimitMiddleware("llm", mustLimiter(llmPolicy), RateLimitKeyUID, md, m),
		Chat: RateLimitMiddleware("chat", mustLimiter(chatPolicy), RateLimitKeyUID, md, m),
	}
}

func parseMode(s string) RateLimitMode {
	md, ok := ParseRateLimitMode(s)
	if !ok {
		logger.From(context.Background()).Warnf("ratelimit.unknown_mode value=%q using=%s", s, md)
	}
	return md
}

// mustLimiter panics on a bad policy: the configs above are code, so a bad
// one is a programming error that must stop the boot.
func mustLimiter(cfg ratelimit.Config) ratelimit.Limiter {
	l, err := ratelimit.New(cfg)
	if err != nil {
		panic(err)
	}
	return l
}
