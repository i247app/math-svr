package bot

import (
	"context"
	"strings"
	"sync"
	"time"

	botAdapter "math-ai.com/math-ai/internal/adapter/bot"
	convcommand "math-ai.com/math-ai/internal/application/command/conversation"
	dto "math-ai.com/math-ai/internal/application/dto/bot"
	"math-ai.com/math-ai/internal/application/transaction"
	conversationDomain "math-ai.com/math-ai/internal/domain/conversation"
	"math-ai.com/math-ai/internal/infrastructure/logger"
)

const (
	// shakeTTL bounds how often the connection warm-up actually hits the
	// vendor. The per-user session init below is NOT throttled (it is a cheap
	// idempotent DB lookup).
	shakeTTL = 60 * time.Second

	// shakeTimeout caps a single warm-up round trip.
	shakeTimeout = 30 * time.Second

	// shakePrompt is the smallest valid prompt that still forces the full
	// network path to the vendor.
	shakePrompt = "ping"
)

// Service backs POST /ai/shake. It does two independent things per call:
//
//  1. Warms the shared LLM connection pool (handshake + keep-alive priming),
//     globally throttled by shakeTTL — a transport optimization with no user
//     identity.
//  2. Initializes the AUTHENTICATED user's AI session: ensures (find-or-create)
//     that user's CHAT conversation thread and returns its conversation_id.
//
// Important: the LLM provider is stateless and cannot recognise end users —
// the connection it warms carries only the server's API key. "Which user is
// this" lives entirely on the server: the session (login) identifies the user,
// and the returned conversation_id is the durable, per-user context handle the
// client reuses on /ai/conversations/send.
type Service struct {
	bot *botAdapter.Adapter

	// per-user session init
	convRepo      conversationDomain.IRepository
	createConvCmd *convcommand.CreateConversationCommandHandler

	// warm-up throttle state
	flightMu sync.Mutex

	stateMu      sync.Mutex
	lastShakedAt time.Time
	provider     string
	model        string
}

// NewService wires the handshake service. bot may be nil (BOT_PROVIDER
// disabled) — the warm-up then reports shaked=false while the per-user
// session init (DB only) still works.
func NewService(bot *botAdapter.Adapter, convRepo conversationDomain.IRepository, uow transaction.UnitOfWork) *Service {
	return &Service{
		bot:           bot,
		convRepo:      convRepo,
		createConvCmd: convcommand.NewCreateConversationCommandHandler(uow),
	}
}

// Shake warms the LLM connection and initializes the user's AI session,
// returning the user's conversation_id for subsequent turns.
// force bypasses the warm-up TTL cache so each call really hits the vendor —
// use it (e.g. /ai/shake?force=true) to observe connection reuse across two
// calls; otherwise a 2nd call within shakeTTL is served from cache.
func (s *Service) Shake(ctx context.Context, req dto.ShakeReq, force bool) *dto.ShakeRes {
	res := s.warm(ctx, req, force)
	res.UserID = req.UID
	res.ConversationID = s.ensureUserSession(ctx, req.UID)
	return res
}

// ensureUserSession resolves (find-or-create) the user's CHAT thread and
// returns its external id. Best-effort: returns 0 on any failure so a session
// hiccup never turns the handshake into a hard error.
func (s *Service) ensureUserSession(ctx context.Context, userID int64) int64 {
	if s.convRepo == nil || s.createConvCmd == nil {
		return 0
	}
	log := logger.From(ctx)

	conv, err := s.convRepo.FindLatestActiveByUserAndPurpose(ctx, userID, conversationDomain.PurposeChat)
	if err != nil {
		log.Warnf("bot.shake_session_failed uid=%d err=%v", userID, err)
		return 0
	}
	if conv != nil {
		return conv.ConversationId()
	}

	created, err := s.createConvCmd.Handle(ctx, convcommand.CreateConversationCommand{
		UserID:  userID,
		Purpose: conversationDomain.PurposeChat,
	})
	if err != nil {
		log.Warnf("bot.shake_session_failed uid=%d err=%v", userID, err)
		return 0
	}
	log.Infof("bot.shake_session_created uid=%d conversation_id=%d", userID, created.ConversationId())
	return created.ConversationId()
}

// warm primes the shared LLM connection pool, globally throttled by shakeTTL.
func (s *Service) warm(ctx context.Context, req dto.ShakeReq, force bool) *dto.ShakeRes {
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

	log.Infof("PROMPT SHAKE: %s", shakePrompt)

	start := time.Now()
	out, err := s.bot.Chat(callCtx, botAdapter.ChatRequest{
		Provider:    botAdapter.BotProviderName(req.ProviderBotName),
		Messages:    []botAdapter.Message{{Role: botAdapter.RoleUser, Content: shakePrompt}},
		MaxTokens:   7,
		Temperature: 0,
	})
	latency := time.Since(start)

	if out != nil {
		log.Infof("BOT RESPONSE: %s", out.Content)
	}

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

	return &dto.ShakeRes{
		Shaked:    true,
		Cached:    false,
		Provider:  provider,
		Model:     model,
		LatencyMs: latency.Milliseconds(),
	}
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
