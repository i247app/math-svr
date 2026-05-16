package logger

import (
	"context"
	"sync/atomic"
)

// Context plumbing for the per-request AppLogger. The key type is
// package-private so no other code can read or write the slot — all
// access goes through Inject / From.

type ctxKey struct{}

// defaultProvider is consulted by From / fallback when no AppLogger is
// bound to ctx. Bootstrap sets it via SetDefault right after NewProvider so
// that boot-time and background-goroutine logs reach the same destination
// (stdout + LogFile) as request logs. Atomic so SetDefault may be called
// from any goroutine; reads in From() are likewise lock-free.
var defaultProvider atomic.Pointer[Provider]

// Inject binds lg to ctx so that downstream code can retrieve it via From.
// The middleware is the canonical caller; tests may also construct a
// throwaway logger and inject it before invoking a handler.
func Inject(ctx context.Context, lg *AppLogger) context.Context {
	return context.WithValue(ctx, ctxKey{}, lg)
}

// From returns the AppLogger bound to ctx, or a no-op fallback when none
// is bound. The fallback writes to the same destination but loses the
// [METHOD path] prefix — never returns nil so callers can dot-chain
// without nil checks.
//
// Typical use:
//
//	logger.From(ctx).Info("user created", "uid", uid)
func From(ctx context.Context) *AppLogger {
	if ctx == nil {
		return fallback()
	}
	if lg, ok := ctx.Value(ctxKey{}).(*AppLogger); ok && lg != nil {
		return lg
	}
	return fallback()
}

// fallback is used when From is called outside an HTTP request flow and no
// background logger has been injected. It prefers the SetDefault-registered
// Provider (so boot / background logs hit the configured LogFile), and
// degrades to a stdout-only logger if no Provider has been registered yet.
func fallback() *AppLogger {
	if p := defaultProvider.Load(); p != nil {
		return p.Background(context.Background())
	}
	return NewBackground(context.Background(), Options{})
}

// SetDefault registers p as the process-wide fallback Provider used by
// From() when no AppLogger is bound to ctx. Calling with nil clears the
// registration. Safe to call concurrently. Bootstrap is the canonical
// caller; Provider.Close() clears the registration so a stale pointer to
// a closed Provider is never returned.
func SetDefault(p *Provider) {
	defaultProvider.Store(p)
}

// Default returns the currently-registered default Provider, or nil when
// none has been set. Most code should call From(ctx) instead — it consults
// the default automatically.
func Default() *Provider {
	return defaultProvider.Load()
}
