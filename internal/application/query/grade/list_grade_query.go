package query

import (
	"context"

	"math-ai.com/math-ai/internal/domain/grade"
	"math-ai.com/math-ai/internal/shared/pagination"
)

// ListGradesQuery narrows the grade listing. GradeIDs restricts the
// result set to the supplied external ids (IN-clause); empty leaves it
// unfiltered.
type ListGradesQuery struct {
	GradeIDs []int64
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
		GradeIds: query.GradeIDs,
		Page:     query.Page,
		Limit:    query.Limit,
	}

	return h.gradeRepo.ListGrades(ctx, params)
}
