package query

import (
	"context"

	"math-ai.com/math-ai/internal/domain/questionpool"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/pagination"
)

// ListQuestionsQuery pages the bank, newest first (OFFSET). Filter comes
// checked (module/question ValidateListQuestions).
type ListQuestionPoolsQuery struct {
	Filter questionpool.ListQuestionPoolsFilter
	Page   int64
	Size   int64
}

type ListQuestionPoolsQueryHandler struct {
	repo questionpool.IRepository
}

func NewListQuestionPoolsQueryHandler(repo questionpool.IRepository) *ListQuestionPoolsQueryHandler {
	return &ListQuestionPoolsQueryHandler{repo: repo}
}

func (h *ListQuestionPoolsQueryHandler) Handle(ctx context.Context, q ListQuestionPoolsQuery) ([]*questionpool.QuestionPool, *pagination.Pagination, error) {
	rows, page, err := h.repo.ListPage(ctx, q.Filter, q.Page, q.Size)
	if err != nil {
		return nil, nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	return rows, page, nil
}
