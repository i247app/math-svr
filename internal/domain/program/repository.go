package program

import (
	"context"

	"math-ai.com/math-ai/internal/shared/pagination"
)

// IRepository owns ma_programs persistence.
type IRepository interface {
	FindByProgramId(ctx context.Context, programId int64) (*Program, error)
	ListPrograms(ctx context.Context, params *ListProgramsParams) ([]*Program, *pagination.Pagination, error)
	// ListProgramsByIds resolves a set of programs in one query. Returns nil
	// slice on empty input; caller maps by ProgramId().
	ListProgramsByIds(ctx context.Context, ids []int64) ([]*Program, error)
	Create(ctx context.Context, p *Program) (*Program, error)
	Update(ctx context.Context, p *Program) error
	SoftDeleteByProgramId(ctx context.Context, programId int64) error
	ForceDeleteByProgramId(ctx context.Context, programId int64) error
}

type ListProgramsParams struct {
	Page    int64
	Limit   int64
	TakeAll bool
}
