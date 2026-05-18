package query

import (
	"context"

	"math-ai.com/math-ai/internal/domain/user"
)

type GetUserByPhoneQuery struct {
	Phone string
}

type GetUserByPhoneQueryHandler struct {
	userRepo user.IRepository
}

// NewGetUserByIdQueryHandler wires the read path. cacheAdapter may be
// nil — the handler degrades to a pure repo lookup in that case so
// boot order and tests don't have to fabricate a cache. Production
// always passes a real adapter via the bootstrap container.
func NewGetUserByPhoneQueryHandler(userRepo user.IRepository) *GetUserByPhoneQueryHandler {
	return &GetUserByPhoneQueryHandler{userRepo: userRepo}
}

func (h *GetUserByPhoneQueryHandler) Handle(ctx context.Context, query GetUserByPhoneQuery) (*user.User, error) {
	u, err := h.userRepo.FindByPhone(ctx, query.Phone)
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
