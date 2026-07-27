package query

import (
	"context"

	"math-ai.com/math-ai/internal/domain/semester"
	"math-ai.com/math-ai/internal/shared/pagination"
)

type ListSemestersQuery struct {
	Page  int64
	Limit int64
}

type ListSemestersQueryHandler struct {
	semesterRepo semester.IRepository
}

func NewListSemestersQueryHandler(semesterRepo semester.IRepository) *ListSemestersQueryHandler {
	return &ListSemestersQueryHandler{semesterRepo: semesterRepo}
}

func (h *ListSemestersQueryHandler) Handle(ctx context.Context, query *ListSemestersQuery) ([]*semester.Semester, *pagination.Pagination, error) {
	params := &semester.ListSemestersParams{
		Page:  query.Page,
		Limit: query.Limit,
	}

	return h.semesterRepo.ListSemesters(ctx, params)
}
