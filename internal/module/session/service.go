package session

import (
	"context"
	"errors"
	"strings"

	dto "math-ai.com/math-ai/internal/application/dto/session"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/infrastructure/logger"
	"math-ai.com/math-ai/internal/infrastructure/session"
)

var (
	ErrTokenRequired   = errors.New("token is required")
	ErrSessionNotFound = errors.New("no session is stored under that token")
)

type Service struct {
	manager *session.SessionManager
}

func NewService(manager *session.SessionManager) *Service {
	return &Service{manager: manager}
}

// DeleteByToken signs one session out: the session stored under the token
// is deleted, so the next request carrying it gets a fresh, empty session —
// exactly what an expired token gets. The token is accepted with or without
// the "Bearer " prefix, the two shapes metadata.authorization takes.
func (s *Service) DeleteByToken(ctx context.Context, req *dto.DeleteSessionReq) (*dto.DeleteSessionRes, error) {
	// Cut the prefix before trimming the right: "Bearer " trimmed first
	// would become "Bearer" and be taken for a token.
	token := strings.TrimLeft(req.Token, " \t")
	if t, ok := strings.CutPrefix(token, "Bearer "); ok {
		token = t
	}
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, errs.NewError(ctx, status.SESSION_MISSING_TOKEN, nil, ErrTokenRequired)
	}

	sess, ok := s.manager.TakeSession(token)
	if !ok {
		return nil, errs.NewError(ctx, status.SESSION_NOT_FOUND, nil, ErrSessionNotFound)
	}

	res := &dto.DeleteSessionRes{IsSecure: sess.IsSecure()}
	uid, hasUID := sess.UID()
	if hasUID {
		res.UID = &uid
	}
	// The key is a live credential until this moment: log only its ends.
	// uid 0 = the session carried none.
	logger.From(ctx).Info("session.deleted", "key", session.ShortKey(token), "uid", uid, "is_secure", res.IsSecure)
	return res, nil
}
