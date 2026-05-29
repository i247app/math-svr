package query

import (
	"context"

	"math-ai.com/math-ai/internal/domain/semester"
	"math-ai.com/math-ai/internal/shared/enum"
)

// GetSemesterByIdQuery returns the semester row with its full
// translation set attached. The base read uses Language for the LEFT
// JOIN override; the translation list is independent of Language so the
// caller always sees every defined locale.
type GetSemesterByIdQuery struct {
	SemesterID string
	Language   enum.LanguageType
}

type GetSemesterByIdQueryHandler struct {
	semesterRepo    semester.IRepository
	translationRepo semester.ITranslationRepository
}

func NewGetSemesterByIdQueryHandler(semesterRepo semester.IRepository, translationRepo semester.ITranslationRepository) *GetSemesterByIdQueryHandler {
	return &GetSemesterByIdQueryHandler{
		semesterRepo:    semesterRepo,
		translationRepo: translationRepo,
	}
}

func (h *GetSemesterByIdQueryHandler) Handle(ctx context.Context, q GetSemesterByIdQuery) (*semester.Semester, error) {
	s, err := h.semesterRepo.FindBySemesterId(ctx, q.SemesterID, q.Language)
	if err != nil {
		return nil, err
	}
	if s == nil {
		return nil, nil
	}
	translations, err := h.translationRepo.ListBySemesterId(ctx, q.SemesterID)
	if err != nil {
		return nil, err
	}
	s.SetTranslations(translations)
	return s, nil
}
