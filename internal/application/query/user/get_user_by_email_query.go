package query

import (
	"context"

	"math-ai.com/math-ai/internal/domain/user"
)

type GetUserByEmailQuery struct {
	Email string
}

type GetUserByEmailQueryHandler struct {
	userRepo user.IRepository
}

// NewGetUserByEmailQueryHandler wires the read path: a plain repo lookup.
func NewGetUserByEmailQueryHandler(userRepo user.IRepository) *GetUserByEmailQueryHandler {
	return &GetUserByEmailQueryHandler{userRepo: userRepo}
}

func (h *GetUserByEmailQueryHandler) Handle(ctx context.Context, query GetUserByEmailQuery) (*user.User, error) {
	u, err := h.userRepo.FindByEmail(ctx, query.Email)
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
