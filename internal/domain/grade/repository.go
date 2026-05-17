package grade

import (
	"context"

	"math-ai.com/math-ai/internal/shared/enum"
	"math-ai.com/math-ai/internal/shared/pagination"
)

type IRepository interface {
	ListGrades(ctx context.Context, params *ListGradesParams) ([]*Grade, *pagination.Pagination, error)
}

type ListGradesParams struct {
	Language enum.LanguageType
	Page     int64
	Limit    int64
	TakeAll  bool
}
