package query

import (
	"context"

	"math-ai.com/math-ai/internal/domain/school"
	"math-ai.com/math-ai/internal/shared/pagination"
)

// ListSchoolsQuery covers both the plain list and the search endpoint —
// Search is just a non-nil filter on the same handler so the repo only
// has one code path to maintain.
type ListSchoolsQuery struct {
	Search   *string
	District *string
	Province *string
	Page     int64
	Limit    int64
}

type ListSchoolsQueryHandler struct {
	schoolRepo school.IRepository
}

func NewListSchoolsQueryHandler(schoolRepo school.IRepository) *ListSchoolsQueryHandler {
	return &ListSchoolsQueryHandler{schoolRepo: schoolRepo}
}

func (h *ListSchoolsQueryHandler) Handle(ctx context.Context, q ListSchoolsQuery) ([]*school.School, *pagination.Pagination, error) {
	return h.schoolRepo.ListSchools(ctx, &school.ListSchoolsParams{
		Search:   q.Search,
		District: q.District,
		Province: q.Province,
		Page:     q.Page,
		Limit:    q.Limit,
	})
}
