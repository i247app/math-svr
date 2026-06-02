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
	FindByGradeId(ctx context.Context, gradeId int64, language enum.LanguageType) (*Grade, error)
	ListGrades(ctx context.Context, params *ListGradesParams) ([]*Grade, *pagination.Pagination, error)
	// ListGradesByIds resolves a set of grades in one query. Returns nil slice
	// on empty input; caller maps by GradeId().
	ListGradesByIds(ctx context.Context, ids []int64, language enum.LanguageType) ([]*Grade, error)
	Create(ctx context.Context, g *Grade) (*Grade, error)
	Update(ctx context.Context, g *Grade) error
	SoftDeleteByGradeId(ctx context.Context, gradeId int64) error
	ForceDeleteByGradeId(ctx context.Context, gradeId int64) error
}

// ITranslationRepository owns the per-language override rows. Listing /
// upsert / delete live here; reads are typically piggybacked on the parent
// LEFT JOIN, so callers that need the full set use ListByGradeId.
type ITranslationRepository interface {
	ListByGradeId(ctx context.Context, gradeId int64) ([]*GradeTranslation, error)
	FindByGradeIdAndLanguage(ctx context.Context, gradeId int64, language string) (*GradeTranslation, error)
	Create(ctx context.Context, t *GradeTranslation) (*GradeTranslation, error)
	Update(ctx context.Context, t *GradeTranslation) error
	SoftDeleteByGradeId(ctx context.Context, gradeId int64) error
	SoftDeleteByTranslationId(ctx context.Context, gradeTranslationId int64) error
	ForceDeleteByGradeId(ctx context.Context, gradeId int64) error
	ForceDeleteByTranslationId(ctx context.Context, gradeTranslationId int64) error
}

// ListGradesParams narrows the listing query. GradeIds restricts the
// result set to the supplied external ids via an IN clause; nil or empty
// leaves it unfiltered. TakeAll bypasses pagination for admin exports.
type ListGradesParams struct {
	Language enum.LanguageType
	GradeIds []int64
	Page     int64
	Limit    int64
	TakeAll  bool
}
