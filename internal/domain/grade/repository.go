package grade

import (
	"context"

	"github.com/google/uuid"
	"math-ai.com/math-ai/internal/shared/enum"
	"math-ai.com/math-ai/internal/shared/pagination"
)

type IRepository interface {
	ListGrades(ctx context.Context, params *ListGradesParams) ([]*Grade, *pagination.Pagination, error)
	// ListGradesByIds resolves a set of grades in one query. Returns nil slice
	// on empty input; caller maps by GradeId().
	ListGradesByIds(ctx context.Context, ids []uuid.UUID, language enum.LanguageType) ([]*Grade, error)
}

type ListGradesParams struct {
	Language enum.LanguageType
	Page     int64
	Limit    int64
	TakeAll  bool
}
