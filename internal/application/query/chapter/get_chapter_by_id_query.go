package query

import (
	"context"

	"math-ai.com/math-ai/internal/domain/chapter"
	"math-ai.com/math-ai/internal/shared/enum"
)

// GetChapterByIdQuery returns the chapter row with its full translation
// set attached. The base read uses Language for the LEFT JOIN override;
// the translation list is independent of Language so the caller always
// sees every defined locale.
type GetChapterByIdQuery struct {
	ChapterID string
	Language  enum.LanguageType
}

type GetChapterByIdQueryHandler struct {
	chapterRepo     chapter.IRepository
	translationRepo chapter.ITranslationRepository
}

func NewGetChapterByIdQueryHandler(chapterRepo chapter.IRepository, translationRepo chapter.ITranslationRepository) *GetChapterByIdQueryHandler {
	return &GetChapterByIdQueryHandler{
		chapterRepo:     chapterRepo,
		translationRepo: translationRepo,
	}
}

func (h *GetChapterByIdQueryHandler) Handle(ctx context.Context, q GetChapterByIdQuery) (*chapter.Chapter, error) {
	c, err := h.chapterRepo.FindByChapterId(ctx, q.ChapterID, q.Language)
	if err != nil {
		return nil, err
	}
	if c == nil {
		return nil, nil
	}
	translations, err := h.translationRepo.ListByChapterId(ctx, q.ChapterID)
	if err != nil {
		return nil, err
	}
	c.SetTranslations(translations)
	return c, nil
}
