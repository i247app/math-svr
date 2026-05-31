package query

import (
	"context"

	"math-ai.com/math-ai/internal/domain/grade"
	"math-ai.com/math-ai/internal/shared/enum"
)

// GetGradeByIdQuery returns the grade row with its full translation
// set attached. The base read uses Language for the LEFT JOIN override;
// the translation list is independent of Language so the caller always
// sees every defined locale.
type GetGradeByIdQuery struct {
	GradeID  int64
	Language enum.LanguageType
}

type GetGradeByIdQueryHandler struct {
	gradeRepo       grade.IRepository
	translationRepo grade.ITranslationRepository
}

func NewGetGradeByIdQueryHandler(gradeRepo grade.IRepository, translationRepo grade.ITranslationRepository) *GetGradeByIdQueryHandler {
	return &GetGradeByIdQueryHandler{
		gradeRepo:       gradeRepo,
		translationRepo: translationRepo,
	}
}

func (h *GetGradeByIdQueryHandler) Handle(ctx context.Context, q GetGradeByIdQuery) (*grade.Grade, error) {
	g, err := h.gradeRepo.FindByGradeId(ctx, q.GradeID, q.Language)
	if err != nil {
		return nil, err
	}
	if g == nil {
		return nil, nil
	}
	translations, err := h.translationRepo.ListByGradeId(ctx, q.GradeID)
	if err != nil {
		return nil, err
	}
	g.SetTranslations(translations)
	return g, nil
}
