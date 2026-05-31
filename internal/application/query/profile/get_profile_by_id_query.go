package query

import (
	"context"

	"math-ai.com/math-ai/internal/domain/profile"
)

type GetProfileByIdQuery struct {
	ProfileId int64
}

type GetProfileByIdQueryHandler struct {
	profileRepo profile.IRepository
}

func NewGetProfileByIdQueryHandler(profileRepo profile.IRepository) *GetProfileByIdQueryHandler {
	return &GetProfileByIdQueryHandler{profileRepo: profileRepo}
}

func (h *GetProfileByIdQueryHandler) Handle(ctx context.Context, query GetProfileByIdQuery) (*profile.Profile, error) {
	return h.profileRepo.FindByProfileId(ctx, query.ProfileId)
}
