package query

import (
	"context"

	"github.com/google/uuid"
	"math-ai.com/math-ai/internal/domain/user"
)

type GetUserByUserIdQuery struct {
	UserId uuid.UUID
}

type GetUserByUserIdQueryHandler struct {
	userRepo user.IRepository
}

// NewGetUserByIdQueryHandler wires the read path. cacheAdapter may be
// nil — the handler degrades to a pure repo lookup in that case so
// boot order and tests don't have to fabricate a cache. Production
// always passes a real adapter via the bootstrap container.
func NewGetUserByUserIdQueryHandler(userRepo user.IRepository) *GetUserByUserIdQueryHandler {
	return &GetUserByUserIdQueryHandler{userRepo: userRepo}
}

func (h *GetUserByUserIdQueryHandler) Handle(ctx context.Context, query GetUserByUserIdQuery) (*user.User, error) {
	u, err := h.userRepo.FindByUserId(ctx, query.UserId)
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
