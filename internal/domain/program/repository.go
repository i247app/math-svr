package program

import (
	"context"

	"math-ai.com/math-ai/internal/shared/enum"
	"math-ai.com/math-ai/internal/shared/pagination"
)

type IRepository interface {
	ListPrograms(ctx context.Context, params *ListProgramsParams) ([]*Program, *pagination.Pagination, error)
}

type ListProgramsParams struct {
	Language enum.LanguageType
	Page     int64
	Limit    int64
	TakeAll  bool
}
