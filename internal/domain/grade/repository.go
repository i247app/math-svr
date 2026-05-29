package grade

import (
	"context"

	"math-ai.com/math-ai/internal/shared/enum"
	"math-ai.com/math-ai/internal/shared/pagination"
)

// IRepository owns ma_grades persistence. Reads LEFT JOIN against
// ma_grade_translations on the caller's language; writes mutate only the
// base row. Translation-row mutations live on ITranslationRepository so
// the two surfaces compose cleanly inside a UoW.
type IRepository interface {
	FindByGradeId(ctx context.Context, gradeId string, language enum.LanguageType) (*Grade, error)
	ListGrades(ctx context.Context, params *ListGradesParams) ([]*Grade, *pagination.Pagination, error)
	// ListGradesByIds resolves a set of grades in one query. Returns nil slice
	// on empty input; caller maps by GradeId().
	ListGradesByIds(ctx context.Context, ids []string, language enum.LanguageType) ([]*Grade, error)
	Create(ctx context.Context, g *Grade) (*Grade, error)
	Update(ctx context.Context, g *Grade) error
	SoftDeleteByGradeId(ctx context.Context, gradeId string) error
	ForceDeleteByGradeId(ctx context.Context, gradeId string) error
}

// ITranslationRepository owns the per-language override rows. Listing /
// upsert / delete live here; reads are typically piggybacked on the parent
// LEFT JOIN, so callers that need the full set use ListByGradeId.
type ITranslationRepository interface {
	ListByGradeId(ctx context.Context, gradeId string) ([]*GradeTranslation, error)
	FindByGradeIdAndLanguage(ctx context.Context, gradeId string, language string) (*GradeTranslation, error)
	Create(ctx context.Context, t *GradeTranslation) (*GradeTranslation, error)
	Update(ctx context.Context, t *GradeTranslation) error
	SoftDeleteByGradeId(ctx context.Context, gradeId string) error
	SoftDeleteByTranslationId(ctx context.Context, gradeTranslationId string) error
	ForceDeleteByGradeId(ctx context.Context, gradeId string) error
	ForceDeleteByTranslationId(ctx context.Context, gradeTranslationId string) error
}

type ListGradesParams struct {
	Language enum.LanguageType
	Page     int64
	Limit    int64
	TakeAll  bool
}
