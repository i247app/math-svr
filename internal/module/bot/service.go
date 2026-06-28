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
	// shakeTTL is how long a successful handshake is considered "fresh".
	// Within this window repeated /ai/shake calls short-circuit to a cached
	// response instead of contacting the LLM again. It bounds the global
	// vendor cost to at most one tiny probe per window no matter how many
	// clients (or how often) call the endpoint — which is exactly what makes
	// the route safe to expose without auth and to call from any screen.
	shakeTTL = 60 * time.Second

	// shakeTimeout caps a single handshake round trip so a stuck upstream can
	// not pin the handler. It is deliberately generous because the first
	// (cold) connection is precisely the slow case we are paying down.
	shakeTimeout = 30 * time.Second

	// shakePrompt is the smallest valid prompt that still forces the full
	// network path (DNS, TLS, HTTP/2 handshake) to the vendor. Paired with
	// MaxTokens=10 it costs ~nothing while priming the connection pool.
	shakePrompt = "ping"
)

// Service owns AI connection handshake. The expensive part of the very first
// AI request in a process — or the first one after the keep-alive pool has
// gone idle — is establishing the connection to the LLM vendor, not the
// generation itself. Handshake pays that cost ahead of time, off the critical
// path of a real quiz/exercise generation, so the user-visible AI call is
// fast.
//
// State is process-global on purpose: the connection pool a handshake primes
// is shared by every request, so one fresh handshake serves all callers. The
// two-mutex design keeps the cached fast-path from blocking behind an
// in-flight probe.
type Service struct {
	bot *botAdapter.Adapter

	// flightMu serializes the actual LLM probe so a burst of concurrent
	// callers triggers exactly one upstream call (single-flight); the rest
	// fall through to the refreshed cache.
	flightMu sync.Mutex

	// stateMu guards the cached handshake snapshot below.
	stateMu      sync.Mutex
	lastShakedAt time.Time
	provider     string
	model        string
}

// NewService wires the handshake service. bot may be nil when the deploy runs
// with BOT_PROVIDER="" / "disabled"; Handshake then reports shaked=false with
// reason "bot_disabled" instead of erroring, so a best-effort FE ping never
// sees a hard failure.
func NewService(bot *botAdapter.Adapter) *Service {
	return &Service{bot: bot}
}

// Handshake primes the LLM connection pool. It never returns an error: a
// shake is best-effort and must not surface as a failure on whatever
// screen triggered it. Outcomes are reported in the result
// (Shaked / Cached / Reason) for observability.
func (s *Service) Shake(ctx context.Context) *dto.ShakeRes {
	log := logger.From(ctx)

	if s.bot == nil {
		return &dto.ShakeRes{Shaked: false, Reason: "bot_disabled"}
	}

	// Fast path: a recent successful shake is still fresh.
	if res, ok := s.cachedIfFresh(); ok {
		log.Info("Fast path: a recent successful shake is still fresh")
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
	// the connection still completes the shake: the goal is to prime the
	// shared pool, not merely to serve this one response.
	callCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), shakeTimeout)
	defer cancel()

	start := time.Now()
	out, err := s.bot.Chat(callCtx, botAdapter.ChatRequest{
		Messages:    []botAdapter.Message{{Role: botAdapter.RoleUser, Content: shakePrompt}},
		MaxTokens:   10,
		Temperature: 0,
	})
	latency := time.Since(start)

	if err != nil {
		// The adapter already logged the cause at the right severity; keep
		// this best-effort and let the caller decide to ignore it.
		log.Warnf("bot.shake_failed latency_ms=%d err=%v", latency.Milliseconds(), err)
		return &dto.ShakeRes{
			Shaked:    false,
			Reason:    "shake_failed",
			LatencyMs: latency.Milliseconds(),
		}
	}

	provider := string(out.Provider)
	if provider == "" {
		provider = string(s.bot.DefaultName())
	}
	model := strings.TrimSpace(out.Model)

	s.stateMu.Lock()
	s.lastShakedAt = time.Now()
	s.provider = provider
	s.model = model
	s.stateMu.Unlock()

	log.Infof("bot.shake_ok provider=%s model=%s latency_ms=%d", provider, model, latency.Milliseconds())

	return &dto.ShakeRes{
		Shaked:    true,
		Cached:    false,
		Provider:  provider,
		Model:     model,
		LatencyMs: latency.Milliseconds(),
	}
}

// cachedIfFresh returns a cached "already shake" result when the last
// successful shake is still within shakeTTL.
func (s *Service) cachedIfFresh() (*dto.ShakeRes, bool) {
	s.stateMu.Lock()
	defer s.stateMu.Unlock()
	if s.lastShakedAt.IsZero() || time.Since(s.lastShakedAt) >= shakeTTL {
		return nil, false
	}
	return &dto.ShakeRes{
		Shaked:    true,
		Cached:    true,
		Provider:  s.provider,
		Model:     s.model,
		LatencyMs: 0,
	}, true
}
