package program

import (
	"context"

	"math-ai.com/math-ai/internal/shared/enum"
	"math-ai.com/math-ai/internal/shared/pagination"
)

type IRepository interface {
	ListPrograms(ctx context.Context, params *ListProgramsParams) ([]*Program, *pagination.Pagination, error)
	// ListProgramsByIds resolves a set of programs in one query. Returns nil
	// slice on empty input; caller maps by ProgramId().
	ListProgramsByIds(ctx context.Context, ids []string, language enum.LanguageType) ([]*Program, error)
}

type ListProgramsParams struct {
	Language enum.LanguageType
	Page     int64
	Limit    int64
	TakeAll  bool
}
