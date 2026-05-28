package query

import (
	"context"

	"math-ai.com/math-ai/internal/domain/chapter"
	"math-ai.com/math-ai/internal/shared/enum"
	"math-ai.com/math-ai/internal/shared/pagination"
)

// ListChaptersQuery filters by program / grade and applies the language
// LEFT JOIN. Translations are intentionally NOT hydrated for list
// responses — clients that need every locale fetch the detail row.
type ListChaptersQuery struct {
	ProgramID  *string
	GradeID    *string
	SemesterID *string
	Language   enum.LanguageType
	Page       int64
	Limit      int64
}

type ListChaptersQueryHandler struct {
	chapterRepo chapter.IRepository
}

func NewListChaptersQueryHandler(chapterRepo chapter.IRepository) *ListChaptersQueryHandler {
	return &ListChaptersQueryHandler{chapterRepo: chapterRepo}
}

func (h *ListChaptersQueryHandler) Handle(ctx context.Context, q ListChaptersQuery) ([]*chapter.Chapter, *pagination.Pagination, error) {
	return h.chapterRepo.ListChapters(ctx, &chapter.ListChaptersParams{
		ProgramID:  q.ProgramID,
		GradeID:    q.GradeID,
		SemesterID: q.SemesterID,
		Language:   q.Language,
		Page:       q.Page,
		Limit:      q.Limit,
	})
}
