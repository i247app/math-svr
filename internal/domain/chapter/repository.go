package chapter

import (
	"context"

	"math-ai.com/math-ai/internal/shared/enum"
	"math-ai.com/math-ai/internal/shared/pagination"
)

// ListChaptersParams narrows the listing query. ProgramID, GradeID, and
// SemesterID are optional and AND'd when set; Language drives the LEFT
// JOIN that overrides the base label/description; TakeAll bypasses
// pagination for callers that genuinely need every chapter (e.g. an
// admin export).
type ListChaptersParams struct {
	ProgramID  *int64
	GradeID    *int64
	SemesterID *int64
	Language   enum.LanguageType
	Page       int64
	Limit      int64
	TakeAll    bool
}

// IRepository owns ma_chapters persistence. Reads LEFT JOIN against
// ma_chapter_translations on the caller's language; writes mutate only the
// base row. Translation-row mutations live on ITranslationRepository so
// the two surfaces compose cleanly inside a UoW.
type IRepository interface {
	FindByChapterId(ctx context.Context, chapterId int64, language enum.LanguageType) (*Chapter, error)
	ListChapters(ctx context.Context, params *ListChaptersParams) ([]*Chapter, *pagination.Pagination, error)
	Create(ctx context.Context, c *Chapter) (*Chapter, error)
	Update(ctx context.Context, c *Chapter) error
	SoftDeleteByChapterId(ctx context.Context, chapterId int64) error
	ForceDeleteByChapterId(ctx context.Context, chapterId int64) error
}

// ITranslationRepository owns the per-language override rows. Listing /
// upsert / delete live here; reads are typically piggybacked on the parent
// LEFT JOIN, so callers that need the full set use ListByChapterId.
type ITranslationRepository interface {
	ListByChapterId(ctx context.Context, chapterId int64) ([]*ChapterTranslation, error)
	FindByChapterIdAndLanguage(ctx context.Context, chapterId int64, language string) (*ChapterTranslation, error)
	Create(ctx context.Context, t *ChapterTranslation) (*ChapterTranslation, error)
	Update(ctx context.Context, t *ChapterTranslation) error
	SoftDeleteByChapterId(ctx context.Context, chapterId int64) error
	SoftDeleteByTranslationId(ctx context.Context, chapterTranslationId int64) error
	ForceDeleteByChapterId(ctx context.Context, chapterId int64) error
	ForceDeleteByTranslationId(ctx context.Context, chapterTranslationId int64) error
}
