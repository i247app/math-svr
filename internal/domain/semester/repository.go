package semester

import (
	"context"

	"math-ai.com/math-ai/internal/shared/enum"
	"math-ai.com/math-ai/internal/shared/pagination"
)

type IRepository interface {
	ListSemesters(ctx context.Context, params *ListSemestersParams) ([]*Semester, *pagination.Pagination, error)
}

type ListSemestersParams struct {
	Language enum.LanguageType
	Page     int64
	Limit    int64
	TakeAll  bool
}
