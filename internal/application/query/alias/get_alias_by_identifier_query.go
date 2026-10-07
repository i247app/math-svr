package alias

import (
	"context"

	"math-ai.com/math-ai/internal/domain/alias"
)

type GetByAkaQuery struct {
	Identifier string
}

type GetAliasByAkaQueryHandler struct {
	aliasRepo alias.IRepository
}

func NewGetAliasByAkaQueryHandler(aliasRepo alias.IRepository) *GetAliasByAkaQueryHandler {
	return &GetAliasByAkaQueryHandler{aliasRepo: aliasRepo}
}

func (h *GetAliasByAkaQueryHandler) Handle(ctx context.Context, query GetByAkaQuery) (*alias.Alias, error) {
	a, err := h.aliasRepo.FindByAka(ctx, query.Identifier)
	if err != nil {
		return nil, err
	}
	if a == nil {
		// Negative caching is intentionally NOT done — a uid that's
		// missing now can be created at any time, and we don't want
		// to serve a stale "not found" until the negative TTL.
		return nil, nil
	}

	return a, nil
}
