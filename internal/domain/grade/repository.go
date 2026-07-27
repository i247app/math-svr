package grade

import (
	"context"

	"math-ai.com/math-ai/internal/shared/pagination"
)

// IRepository owns ma_grades persistence.
type IRepository interface {
	FindByGradeId(ctx context.Context, gradeId int64) (*Grade, error)
	ListGrades(ctx context.Context, params *ListGradesParams) ([]*Grade, *pagination.Pagination, error)
	// ListGradesByIds resolves a set of grades in one query. Returns nil slice
	// on empty input; caller maps by GradeId().
	ListGradesByIds(ctx context.Context, ids []int64) ([]*Grade, error)
	Create(ctx context.Context, g *Grade) (*Grade, error)
	Update(ctx context.Context, g *Grade) error
	SoftDeleteByGradeId(ctx context.Context, gradeId int64) error
	ForceDeleteByGradeId(ctx context.Context, gradeId int64) error
}

// ListGradesParams narrows the listing query. GradeIds restricts the
// result set to the supplied external ids via an IN clause; nil or empty
// leaves it unfiltered. TakeAll bypasses pagination for admin exports.
type ListGradesParams struct {
	GradeIds []int64
	Page     int64
	Limit    int64
	TakeAll  bool
}
