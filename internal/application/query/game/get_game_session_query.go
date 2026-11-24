package query

import (
	"context"
	"fmt"

	"math-ai.com/math-ai/internal/domain/game"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/errs"
)

type GetGameSessionQuery struct {
	GsessID   int64
	UID       int64
	ProfileID int64
}

type GetGameSessionQueryHandler struct {
	repo game.IRepository
}

func NewGetGameSessionQueryHandler(repo game.IRepository) *GetGameSessionQueryHandler {
	return &GetGameSessionQueryHandler{repo: repo}
}

func (h *GetGameSessionQueryHandler) Handle(ctx context.Context, q GetGameSessionQuery) (*game.GameSession, error) {
	s, err := h.repo.FindByGsessId(ctx, q.GsessID)
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	if s == nil {
		return nil, errs.NewError(ctx, game.GameSessionNotFound, nil, fmt.Errorf("game: session %d not found", q.GsessID))
	}
	if !s.OwnedBy(q.UID, q.ProfileID) {
		return nil, errs.NewError(ctx, game.GameSessionNotOwned, nil, fmt.Errorf("game: session %d belongs to another profile", q.GsessID))
	}
	return s, nil
}
