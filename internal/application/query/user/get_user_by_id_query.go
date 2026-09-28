package query

import (
	"context"

	"math-ai.com/math-ai/internal/domain/user"
)

type GetUserByUidQuery struct {
	Uid int64
}

type GetUserByUidQueryHandler struct {
	userRepo user.IRepository
}

// NewGetUserByIdQueryHandler wires the read path. cacheAdapter may be
// nil — the handler degrades to a pure repo lookup in that case so
// boot order and tests don't have to fabricate a cache. Production
// always passes a real adapter via the bootstrap container.
func NewGetUserByUidQueryHandler(userRepo user.IRepository) *GetUserByUidQueryHandler {
	return &GetUserByUidQueryHandler{userRepo: userRepo}
}

func (h *GetUserByUidQueryHandler) Handle(ctx context.Context, query GetUserByUidQuery) (*user.User, error) {
	u, err := h.userRepo.FindByUid(ctx, query.Uid)
	if err != nil {
		return nil, err
	}
	if u == nil {
		// Negative caching is intentionally NOT done — a uid that's
		// missing now can be created at any time, and we don't want
		// to serve a stale "not found" until the negative TTL.
		return nil, nil
	}

	return u, nil
}
