package query

import (
	"context"

	"math-ai.com/math-ai/internal/domain/question"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/pagination"
)

// ListQuestionsQuery pages the bank, newest first (OFFSET). Filter comes
// checked (module/question ValidateListQuestions).
type ListQuestionsQuery struct {
	Filter question.ListQuestionsFilter
	Page   int64
	Size   int64
}

type ListQuestionsQueryHandler struct {
	repo question.IRepository
}

func NewListQuestionsQueryHandler(repo question.IRepository) *ListQuestionsQueryHandler {
	return &ListQuestionsQueryHandler{repo: repo}
}

func (h *ListQuestionsQueryHandler) Handle(ctx context.Context, q ListQuestionsQuery) ([]*question.Question, *pagination.Pagination, error) {
	rows, page, err := h.repo.ListPage(ctx, q.Filter, q.Page, q.Size)
	if err != nil {
		return nil, nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	return rows, page, nil
}
