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
	ErrTokenRequired    = errors.New("token is required")
	ErrSessionNotFound  = errors.New("no session is stored under that token")
	ErrIsSecureRequired = errors.New("is_secure is required")
	ErrSecureWithoutUID = errors.New("a session without a uid cannot be marked secure")
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
	token, err := normalizeToken(ctx, req.Token)
	if err != nil {
		return nil, err
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

// MarkSecureByToken sets is_secure on one session, leaving everything else
// on it (uid, login name, expiry) as it is. false is a soft sign-out: the
// session keeps its uid but every auth-gated route refuses it until the user
// signs in again. true is refused on a session holding no uid —
// AuthRequiredMiddleware checks is_secure alone, so such a session would
// pass every gated route as nobody.
func (s *Service) MarkSecureByToken(ctx context.Context, req *dto.MarkSessionSecureReq) (*dto.MarkSessionSecureRes, error) {
	token, err := normalizeToken(ctx, req.Token)
	if err != nil {
		return nil, err
	}
	if req.IsSecure == nil {
		return nil, errs.NewError(ctx, status.SESSION_MISSING_IS_SECURE, nil, ErrIsSecureRequired)
	}
	secure := *req.IsSecure

	sess, ok := s.manager.Session(token)
	if !ok {
		return nil, errs.NewError(ctx, status.SESSION_NOT_FOUND, nil, ErrSessionNotFound)
	}
	uid, hasUID := sess.UID()
	if secure && (!hasUID || uid <= 0) {
		return nil, errs.NewError(ctx, status.SESSION_MISSING_UID, nil, ErrSecureWithoutUID)
	}
	if _, ok := s.manager.SetSecure(token, secure); !ok {
		// Deleted between the read and the write.
		return nil, errs.NewError(ctx, status.SESSION_NOT_FOUND, nil, ErrSessionNotFound)
	}

	res := &dto.MarkSessionSecureRes{IsSecure: secure}
	if hasUID {
		res.UID = &uid
	}
	logger.From(ctx).Info("session.marked_secure", "key", session.ShortKey(token), "uid", uid, "is_secure", secure)
	return res, nil
}

// normalizeToken accepts the token with or without the "Bearer " prefix,
// the two shapes metadata.authorization takes. The prefix is cut before the
// right is trimmed: "Bearer " trimmed first would become "Bearer" and be
// taken for a token.
func normalizeToken(ctx context.Context, raw string) (string, error) {
	token := strings.TrimLeft(raw, " \t")
	if t, ok := strings.CutPrefix(token, "Bearer "); ok {
		token = t
	}
	token = strings.TrimSpace(token)
	if token == "" {
		return "", errs.NewError(ctx, status.SESSION_MISSING_TOKEN, nil, ErrTokenRequired)
	}
	return token, nil
}
