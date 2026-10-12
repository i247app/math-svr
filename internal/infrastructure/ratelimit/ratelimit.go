// Package ratelimit holds the in-process rate-limit algorithms. Each one is a
// Limiter keyed by an opaque string (an IP, a uid, …); the HTTP side lives in
// bootstrap/middleware/rate_limit_middleware.go.
//
// State is process memory, which matches the rest of the server: sessions and
// the socket Hub are in memory too, so math-svr runs as a single instance. A
// multi-instance deploy needs a shared backend (e.g. Redis + Lua) behind the
// same Limiter interface — that is what ctx and the error return are for.
package ratelimit

import (
	"context"
	"fmt"
	"time"
)

// Algorithm names a rate-limit strategy.
type Algorithm string

const (
	// FixedWindow counts requests per aligned window. O(1) memory per key;
	// a client can get up to 2×Limit across a window boundary.
	FixedWindow Algorithm = "fixed_window"
	// SlidingWindowCounter weights the previous window's count by how much of
	// it still overlaps the sliding window. O(1) memory, close approximation.
	SlidingWindowCounter Algorithm = "sliding_window_counter"
	// SlidingWindowLog keeps one timestamp per accepted request. Exact, but
	// memory grows with Limit — use it for small limits only.
	SlidingWindowLog Algorithm = "sliding_window_log"
	// LeakyBucket queues requests and lets them out at a fixed pace: an
	// accepted request may be told to wait (Decision.Delay). Burst is the
	// queue size; a full queue rejects.
	LeakyBucket Algorithm = "leaky_bucket"
	// TokenBucket refills Limit tokens per Window up to Burst; a request
	// spends one. Allows short bursts, then a steady rate.
	TokenBucket Algorithm = "token_bucket"
)

// Config is one policy. Limit requests per Window is the steady rate for
// every algorithm; Burst only matters to the two buckets (capacity / queue
// size) and defaults to Limit.
type Config struct {
	Algorithm Algorithm
	Limit     int
	Window    time.Duration
	Burst     int
	// MaxKeys caps tracked keys so a flood of distinct keys cannot exhaust
	// memory. Default 100 000.
	MaxKeys int
	// Now is the clock; nil means time.Now. Tests inject a fake one.
	Now func() time.Time
}

// Decision is the outcome of one Allow call.
type Decision struct {
	Allowed   bool
	Limit     int
	Remaining int
	// RetryAfter is how long a rejected caller should wait.
	RetryAfter time.Duration
	// Delay is how long an ACCEPTED request must wait before it runs. Only the
	// leaky bucket sets it.
	Delay time.Duration
}

// Limiter decides whether the next request for key may proceed.
type Limiter interface {
	Allow(ctx context.Context, key string) (Decision, error)
}

// New builds the Limiter for cfg.
func New(cfg Config) (Limiter, error) {
	if cfg.Limit <= 0 || cfg.Window <= 0 {
		return nil, fmt.Errorf("ratelimit: limit and window must be positive (limit=%d window=%s)", cfg.Limit, cfg.Window)
	}
	if cfg.Burst <= 0 {
		cfg.Burst = cfg.Limit
	}
	if cfg.MaxKeys <= 0 {
		cfg.MaxKeys = 100_000
	}
	if cfg.Now == nil {
		cfg.Now = time.Now
	}

	// A key idle for longer than this is indistinguishable from a fresh one
	// (its window has rolled over / its bucket is full again), so it can go.
	refill := cfg.Window * time.Duration(cfg.Burst) / time.Duration(cfg.Limit)
	idle := 2 * max(cfg.Window, refill)

	switch cfg.Algorithm {
	case FixedWindow:
		return &fixedWindow{cfg: cfg, st: newStore[fixedWindowState](cfg.MaxKeys, idle)}, nil
	case SlidingWindowCounter:
		return &slidingWindowCounter{cfg: cfg, st: newStore[slidingWindowCounterState](cfg.MaxKeys, idle)}, nil
	case SlidingWindowLog:
		return &slidingWindowLog{cfg: cfg, st: newStore[slidingWindowLogState](cfg.MaxKeys, idle)}, nil
	case LeakyBucket:
		return &leakyBucket{cfg: cfg, st: newStore[leakyBucketState](cfg.MaxKeys, idle)}, nil
	case TokenBucket:
		return &tokenBucket{cfg: cfg, st: newStore[tokenBucketState](cfg.MaxKeys, idle)}, nil
	default:
		return nil, fmt.Errorf("ratelimit: unknown algorithm %q", cfg.Algorithm)
	}
}

// untracked is returned when the key table is full: fail open rather than
// lock out every new client during a key flood.
func untracked(cfg Config) Decision {
	return Decision{Allowed: true, Limit: cfg.Limit, Remaining: cfg.Limit}
}

// interval is the time one request "costs" at the steady rate.
func interval(cfg Config) time.Duration {
	return cfg.Window / time.Duration(cfg.Limit)
}
