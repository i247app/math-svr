package query

import (
	"context"

	"math-ai.com/math-ai/internal/domain/llmroute"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
)

// ListLLMRoutesQuery returns every live route, ordered by id.
type ListLLMRoutesQuery struct {
	repo llmroute.IRepository
}

func NewListLLMRoutesQuery(repo llmroute.IRepository) *ListLLMRoutesQuery {
	return &ListLLMRoutesQuery{repo: repo}
}

func (q *ListLLMRoutesQuery) Handle(ctx context.Context) ([]*llmroute.LLMRoute, error) {
	routes, err := q.repo.List(ctx)
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	return routes, nil
}
