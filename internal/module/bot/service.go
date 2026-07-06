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
	// shakeTTL bounds how often the warm-up actually hits the vendor.
	shakeTTL = 60 * time.Second

	// shakeTimeout caps a single warm-up round trip.
	shakeTimeout = 30 * time.Second

	// shakePrompt is the smallest valid prompt that still forces the full
	// network path to the vendor.
	shakePrompt = "ping"
)

// Service backs POST /ai/shake: it warms the shared LLM connection pool
// (handshake + keep-alive priming) so the first real quiz/exercise call is
// fast. Globally throttled by shakeTTL + single-flight so the vendor is hit
// at most once per window regardless of caller volume.
type Service struct {
	bot *botAdapter.Adapter

	flightMu sync.Mutex

	stateMu      sync.Mutex
	lastShakedAt time.Time
	provider     string
	model        string
}

// NewService wires the handshake service. bot may be nil (BOT_PROVIDER
// disabled) — Shake then reports shaked=false.
func NewService(bot *botAdapter.Adapter) *Service {
	return &Service{bot: bot}
}

// Shake warms the LLM connection. force bypasses the TTL cache so each call
// really hits the vendor — use /ai/shake?force=true to observe connection
// reuse across two calls; otherwise a 2nd call within shakeTTL is cached.
func (s *Service) Shake(ctx context.Context, req dto.ShakeReq, force bool) *dto.ShakeRes {
	log := logger.From(ctx)

	if s.bot == nil {
		return &dto.ShakeRes{Shaked: false, Reason: "bot_disabled"}
	}

	if !force {
		if res, ok := s.cachedIfFresh(); ok {
			return res
		}
	}

	s.flightMu.Lock()
	defer s.flightMu.Unlock()

	if !force {
		if res, ok := s.cachedIfFresh(); ok {
			return res
		}
	}

	callCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), shakeTimeout)
	defer cancel()

	start := time.Now()
	out, err := s.bot.Chat(callCtx, botAdapter.ChatRequest{
		Provider:    botAdapter.BotProviderName(req.ProviderBotName),
		Messages:    []botAdapter.Message{{Role: botAdapter.RoleUser, Content: shakePrompt}},
		MaxTokens:   10,
		Temperature: 0,
	})
	latency := time.Since(start)

	if err != nil {
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

	res := &dto.ShakeRes{
		Shaked:    true,
		Cached:    false,
		Provider:  provider,
		Model:     model,
		LatencyMs: latency.Milliseconds(),
	}

	if out != nil {
		res.Content = out.Content
	}

	return res
}

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
