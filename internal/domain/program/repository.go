package program

import (
	"context"

	"math-ai.com/math-ai/internal/shared/enum"
	"math-ai.com/math-ai/internal/shared/pagination"
)

// IRepository owns ma_programs persistence. Reads LEFT JOIN against
// ma_program_translations on the caller's language; writes mutate only
// the base row. Translation-row mutations live on ITranslationRepository
// so the two surfaces compose cleanly inside a UoW.
type IRepository interface {
	FindByProgramId(ctx context.Context, programId int64, language enum.LanguageType) (*Program, error)
	ListPrograms(ctx context.Context, params *ListProgramsParams) ([]*Program, *pagination.Pagination, error)
	// ListProgramsByIds resolves a set of programs in one query. Returns nil
	// slice on empty input; caller maps by ProgramId().
	ListProgramsByIds(ctx context.Context, ids []int64, language enum.LanguageType) ([]*Program, error)
	Create(ctx context.Context, p *Program) (*Program, error)
	Update(ctx context.Context, p *Program) error
	SoftDeleteByProgramId(ctx context.Context, programId int64) error
	ForceDeleteByProgramId(ctx context.Context, programId int64) error
}

// ITranslationRepository owns the per-language override rows. Listing /
// upsert / delete live here; reads are typically piggybacked on the parent
// LEFT JOIN, so callers that need the full set use ListByProgramId.
type ITranslationRepository interface {
	ListByProgramId(ctx context.Context, programId int64) ([]*ProgramTranslation, error)
	FindByProgramIdAndLanguage(ctx context.Context, programId int64, language string) (*ProgramTranslation, error)
	Create(ctx context.Context, t *ProgramTranslation) (*ProgramTranslation, error)
	Update(ctx context.Context, t *ProgramTranslation) error
	SoftDeleteByProgramId(ctx context.Context, programId int64) error
	SoftDeleteByTranslationId(ctx context.Context, programTranslationId int64) error
	ForceDeleteByProgramId(ctx context.Context, programId int64) error
	ForceDeleteByTranslationId(ctx context.Context, programTranslationId int64) error
}

type ListProgramsParams struct {
	Language enum.LanguageType
	Page     int64
	Limit    int64
	TakeAll  bool
}
