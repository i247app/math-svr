package bot

import (
	"context"
	"strings"
	"sync"
	"time"

	botAdapter "math-ai.com/math-ai/internal/adapter/bot"
	dto "math-ai.com/math-ai/internal/application/dto/bot"
	"math-ai.com/math-ai/internal/infrastructure/logger"
)

const (
	// warmupTTL is how long a successful warm-up is considered "fresh".
	// Within this window repeated /connect/init-ai calls short-circuit to a cached
	// response instead of contacting the LLM again. It bounds the global
	// vendor cost to at most one tiny probe per window no matter how many
	// clients (or how often) call the endpoint — which is exactly what makes
	// the route safe to expose without auth and to call from any screen.
	warmupTTL = 60 * time.Second

	// warmupTimeout caps a single warm-up round trip so a stuck upstream can
	// not pin the handler. It is deliberately generous because the first
	// (cold) connection is precisely the slow case we are paying down.
	warmupTimeout = 30 * time.Second

	// warmupPrompt is the smallest valid prompt that still forces the full
	// network path (DNS, TLS, HTTP/2 handshake) to the vendor. Paired with
	// MaxTokens=1 it costs ~nothing while priming the connection pool.
	warmupPrompt = "ping"
)

// Service owns AI connection warm-up. The expensive part of the very first
// AI request in a process — or the first one after the keep-alive pool has
// gone idle — is establishing the connection to the LLM vendor, not the
// generation itself. Warm-up pays that cost ahead of time, off the critical
// path of a real quiz/exercise generation, so the user-visible AI call is
// fast.
//
// State is process-global on purpose: the connection pool a warm-up primes
// is shared by every request, so one fresh warm-up serves all callers. The
// two-mutex design keeps the cached fast-path from blocking behind an
// in-flight probe.
type Service struct {
	bot *botAdapter.Adapter

	// flightMu serializes the actual LLM probe so a burst of concurrent
	// callers triggers exactly one upstream call (single-flight); the rest
	// fall through to the refreshed cache.
	flightMu sync.Mutex

	// stateMu guards the cached warm-up snapshot below.
	stateMu      sync.Mutex
	lastWarmedAt time.Time
	provider     string
	model        string
}

// NewService wires the warm-up service. bot may be nil when the deploy runs
// with BOT_PROVIDER="" / "disabled"; Warmup then reports warmed=false with
// reason "bot_disabled" instead of erroring, so a best-effort FE ping never
// sees a hard failure.
func NewService(bot *botAdapter.Adapter) *Service {
	return &Service{bot: bot}
}

// Shake primes the LLM connection pool. It never returns an error: a
// warm-up is best-effort and must not surface as a failure on whatever
// screen triggered it. Outcomes are reported in the result
// (Warmed / Cached / Reason) for observability.
func (s *Service) Shake(ctx context.Context) *dto.WarmupRes {
	log := logger.From(ctx)

	if s.bot == nil {
		return &dto.WarmupRes{Warmed: false, Reason: "bot_disabled"}
	}

	// Fast path: a recent successful warm-up is still fresh.
	if res, ok := s.cachedIfFresh(); ok {
		log.Info("Fast path: a recent successful warm-up is still fresh")
		return res
	}

	// Serialize the probe: concurrent callers coalesce onto one upstream
	// call and then return the snapshot it refreshes.
	s.flightMu.Lock()
	defer s.flightMu.Unlock()

	// Re-check under the flight lock — a probe that was in flight while we
	// waited for the lock may have just refreshed the cache.
	if res, ok := s.cachedIfFresh(); ok {
		return res
	}

	// Detach from the request context so a fire-and-forget client that drops
	// the connection still completes the warm-up: the goal is to prime the
	// shared pool, not merely to serve this one response.
	callCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), warmupTimeout)
	defer cancel()

	start := time.Now()
	out, err := s.bot.Chat(callCtx, botAdapter.ChatRequest{
		Messages:    []botAdapter.Message{{Role: botAdapter.RoleUser, Content: warmupPrompt}},
		MaxTokens:   10,
		Temperature: 0,
	})
	latency := time.Since(start)

	if err != nil {
		// The adapter already logged the cause at the right severity; keep
		// this best-effort and let the caller decide to ignore it.
		log.Warnf("bot.warmup_failed latency_ms=%d err=%v", latency.Milliseconds(), err)
		return &dto.WarmupRes{
			Warmed:    false,
			Reason:    "warmup_failed",
			LatencyMs: latency.Milliseconds(),
		}
	}

	provider := string(out.Provider)
	if provider == "" {
		provider = string(s.bot.DefaultName())
	}
	model := strings.TrimSpace(out.Model)

	s.stateMu.Lock()
	s.lastWarmedAt = time.Now()
	s.provider = provider
	s.model = model
	s.stateMu.Unlock()

	log.Infof("bot.warmup_ok provider=%s model=%s latency_ms=%d", provider, model, latency.Milliseconds())

	return &dto.WarmupRes{
		Warmed:    true,
		Cached:    false,
		Provider:  provider,
		Model:     model,
		LatencyMs: latency.Milliseconds(),
	}
}

// cachedIfFresh returns a cached "already warm" result when the last
// successful warm-up is still within warmupTTL.
func (s *Service) cachedIfFresh() (*dto.WarmupRes, bool) {
	s.stateMu.Lock()
	defer s.stateMu.Unlock()
	if s.lastWarmedAt.IsZero() || time.Since(s.lastWarmedAt) >= warmupTTL {
		return nil, false
	}
	return &dto.WarmupRes{
		Warmed:    true,
		Cached:    true,
		Provider:  s.provider,
		Model:     s.model,
		LatencyMs: 0,
	}, true
}
