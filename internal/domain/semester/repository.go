package semester

import (
	"context"

	"math-ai.com/math-ai/internal/shared/enum"
	"math-ai.com/math-ai/internal/shared/pagination"
)

// IRepository owns ma_semesters persistence. Reads LEFT JOIN against
// ma_semester_translations on the caller's language; writes mutate only
// the base row. Translation-row mutations live on ITranslationRepository
// so the two surfaces compose cleanly inside a UoW.
type IRepository interface {
	FindBySemesterId(ctx context.Context, semesterId int64, language enum.LanguageType) (*Semester, error)
	ListSemesters(ctx context.Context, params *ListSemestersParams) ([]*Semester, *pagination.Pagination, error)
	// ListSemestersByIds resolves a set of semesters in one query. Returns nil
	// slice on empty input; caller maps by SemesterId().
	ListSemestersByIds(ctx context.Context, ids []int64, language enum.LanguageType) ([]*Semester, error)
	Create(ctx context.Context, s *Semester) (*Semester, error)
	Update(ctx context.Context, s *Semester) error
	SoftDeleteBySemesterId(ctx context.Context, semesterId int64) error
	ForceDeleteBySemesterId(ctx context.Context, semesterId int64) error
}

// ITranslationRepository owns the per-language override rows. Listing /
// upsert / delete live here; reads are typically piggybacked on the parent
// LEFT JOIN, so callers that need the full set use ListBySemesterId.
type ITranslationRepository interface {
	ListBySemesterId(ctx context.Context, semesterId int64) ([]*SemesterTranslation, error)
	FindBySemesterIdAndLanguage(ctx context.Context, semesterId int64, language enum.LanguageType) (*SemesterTranslation, error)
	Create(ctx context.Context, t *SemesterTranslation) (*SemesterTranslation, error)
	Update(ctx context.Context, t *SemesterTranslation) error
	SoftDeleteBySemesterId(ctx context.Context, semesterId int64) error
	SoftDeleteByTranslationId(ctx context.Context, semesterTranslationId int64) error
	ForceDeleteBySemesterId(ctx context.Context, semesterId int64) error
	ForceDeleteByTranslationId(ctx context.Context, semesterTranslationId int64) error
}

type ListSemestersParams struct {
	Language enum.LanguageType
	Page     int64
	Limit    int64
	TakeAll  bool
}
