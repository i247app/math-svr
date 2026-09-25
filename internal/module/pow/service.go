package pow

import (
	"context"
	"crypto/sha256"
	"time"

	dto "math-ai.com/math-ai/internal/application/dto/pow"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/infrastructure/logger"
	"math-ai.com/math-ai/internal/infrastructure/session"
	"math-ai.com/math-ai/internal/shared/utils"
)

const (
	// challengeTTL bounds how long a client may take to solve a challenge.
	challengeTTL = 15 * time.Minute
	// passTTL bounds how long a solved challenge may wait before it is spent.
	passTTL = 15 * time.Minute
	// seedLength makes seeds unguessable, so solutions cannot be precomputed.
	seedLength = 5

	sessionKeyChallenge = "pow_challenge"
	sessionKeyPassedAt  = "pow_passed_at"
)

type Service struct {
	enabled    bool
	difficulty int
}

func NewService(enabled bool, difficulty int) *Service {
	if difficulty < 1 {
		difficulty = 1
	}
	if difficulty > sha256.Size*2 { // a SHA-256 hex digest has 64 chars
		difficulty = sha256.Size * 2
	}
	return &Service{enabled: enabled, difficulty: difficulty}
}

func (s *Service) GenerateChallenge(ctx context.Context, sess *session.AppSession) (*dto.ChallengeRes, error) {
	seed, err := utils.RandomStringWithLenght(seedLength)
	if err != nil {
		return nil, errs.NewError(ctx, status.POW_CHALLENGE_FAILED, nil, err)
	}

	sess.Put(sessionKeyChallenge, session.Pow{
		Seed:       seed,
		Difficulty: s.difficulty,
		CreatedAt:  time.Now(),
	})

	logger.From(ctx).Info("pow.challenge_issued", "difficulty", s.difficulty)

	res := &dto.ChallengeRes{
		Message:    seed,
		Difficulty: s.difficulty,
	}

	return res, nil
}

func (s *Service) VerifyChallenge(ctx context.Context, sess *session.AppSession, nonce int64) (*dto.VerifyChallengeRes, error) {
	log := logger.From(ctx)

	challenge, _ := sess.Get(sessionKeyChallenge)
	pending, _ := challenge.(session.Pow)
	if pending.Seed == "" {
		return nil, errs.NewError(ctx, status.POW_CHALLENGE_NOT_FOUND, nil, ErrChallengeNotFound)
	}

	// Consume. A zero value, not a delete: the session store has no delete,
	// and a nil value cannot be persisted (gob).
	sess.Put(sessionKeyChallenge, session.Pow{})

	if time.Since(pending.CreatedAt) > challengeTTL {
		log.Info("pow.verify_rejected", "reason", "expired")
		return nil, errs.NewError(ctx, status.POW_CHALLENGE_EXPIRED, nil, ErrChallengeExpired)
	}
	if !Solves(pending.Seed, nonce, pending.Difficulty) {
		log.Info("pow.verify_rejected", "reason", "invalid_proof")
		return nil, errs.NewError(ctx, status.POW_INVALID_PROOF, nil, ErrInvalidProof)
	}

	sess.Put(sessionKeyPassedAt, time.Now())
	log.Info("pow.verified", "difficulty", pending.Difficulty)

	res := &dto.VerifyChallengeRes{
		IsValid: true,
	}

	return res, nil
}

func (s *Service) ConsumePass(ctx context.Context, sess *session.AppSession) error {
	if !s.enabled {
		return nil
	}

	raw, _ := sess.Get(sessionKeyPassedAt)
	passedAt, _ := raw.(time.Time)
	if passedAt.IsZero() || time.Since(passedAt) > passTTL {
		return errs.NewError(ctx, status.POW_REQUIRED, nil, ErrPowRequired)
	}

	sess.Put(sessionKeyPassedAt, time.Time{})
	return nil
}
