package query

import (
	"context"

	"math-ai.com/math-ai/internal/domain/profile"
)

type ListProfilesByUserIdQuery struct {
	UserId int64
}

type ListProfilesByUserIdQueryHandler struct {
	profileRepo profile.IRepository
}

func NewListProfilesByUserIdQueryHandler(profileRepo profile.IRepository) *ListProfilesByUserIdQueryHandler {
	return &ListProfilesByUserIdQueryHandler{profileRepo: profileRepo}
}

func (h *ListProfilesByUserIdQueryHandler) Handle(ctx context.Context, query ListProfilesByUserIdQuery) ([]*profile.Profile, error) {
	return h.profileRepo.ListByUserId(ctx, query.UserId)
}
