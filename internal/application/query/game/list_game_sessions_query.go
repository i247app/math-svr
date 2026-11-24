package query

import (
	"context"

	"math-ai.com/math-ai/internal/domain/game"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/errs"
	"math-ai.com/math-ai/internal/shared/pagination"
)

// ListGameSessionsQuery pages a child's COMPLETE rounds, newest first.
type ListGameSessionsQuery struct {
	UID       int64
	ProfileID int64
	Paging    pagination.Request
}

type ListGameSessionsResult struct {
	GameSessions []*game.GameSession
	Pagination   *pagination.Pagination
	Cursor       *pagination.CursorPagination
}

// gameCursor is a round's position in the list (gsess_id DESC, unique).
type gameCursor struct {
	GsessId int64 `json:"gsess_id"`
}

type ListGameSessionsQueryHandler struct {
	repo game.IRepository
}

func NewListGameSessionsQueryHandler(repo game.IRepository) *ListGameSessionsQueryHandler {
	return &ListGameSessionsQueryHandler{repo: repo}
}

func (h *ListGameSessionsQueryHandler) Handle(ctx context.Context, q ListGameSessionsQuery) (*ListGameSessionsResult, error) {
	var res ListGameSessionsResult
	var err error
	if q.Paging.IsCursor() {
		res.GameSessions, res.Cursor, err = pagination.KeysetPage(ctx, h.keyset(q.UID, q.ProfileID), q.Paging.Next, q.Paging.Previous, q.Paging.Size)
	} else {
		res.GameSessions, res.Pagination, err = h.repo.ListCompletePage(ctx, q.UID, q.ProfileID, q.Paging.Page, q.Paging.Size)
	}
	if err != nil {
		if _, ok := errs.IsMathError(err); ok {
			return nil, err // a bad cursor — already PAGINATION_INVALID_CURSOR
		}
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	return &res, nil
}

func (h *ListGameSessionsQueryHandler) keyset(uid, profileId int64) pagination.Keyset[*game.GameSession, gameCursor] {
	return pagination.Keyset[*game.GameSession, gameCursor]{
		Read: func(ctx context.Context, after, before *gameCursor, limit int64) ([]*game.GameSession, error) {
			params := game.KeysetParams{Limit: limit}
			if after != nil {
				params.Older = &after.GsessId
			}
			if before != nil {
				params.Newer = &before.GsessId
			}
			return h.repo.ListCompleteKeyset(ctx, uid, profileId, params)
		},
		// Newest first: "before" a round in the list means newer.
		ExistsAtOrBefore: func(ctx context.Context, c gameCursor) (bool, error) {
			return h.repo.ExistsCompleteAtOrNewer(ctx, uid, profileId, c.GsessId)
		},
		ExistsAtOrAfter: func(ctx context.Context, c gameCursor) (bool, error) {
			return h.repo.ExistsCompleteAtOrOlder(ctx, uid, profileId, c.GsessId)
		},
		Key:   func(s *game.GameSession) gameCursor { return gameCursor{GsessId: s.GsessId()} },
		Valid: func(c gameCursor) bool { return c.GsessId > 0 },
	}
}
