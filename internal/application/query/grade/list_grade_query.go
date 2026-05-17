package query

import (
	"context"

	"math-ai.com/math-ai/internal/domain/grade"
	"math-ai.com/math-ai/internal/shared/enum"
	"math-ai.com/math-ai/internal/shared/pagination"
)

type ListGradesQuery struct {
	Language enum.LanguageType
	Page     int64
	Limit    int64
}

type ListGradesQueryHandler struct {
	gradeRepo grade.IRepository
}

func NewListGradesQueryHandler(gradeRepo grade.IRepository) *ListGradesQueryHandler {
	return &ListGradesQueryHandler{gradeRepo: gradeRepo}
}

func (h *ListGradesQueryHandler) Handle(ctx context.Context, query *ListGradesQuery) ([]*grade.Grade, *pagination.Pagination, error) {
	params := &grade.ListGradesParams{
		Language: query.Language,
		Page:     query.Page,
		Limit:    query.Limit,
	}

	return h.gradeRepo.ListGrades(ctx, params)
}
