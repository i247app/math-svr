package query

import (
	"context"

	"math-ai.com/math-ai/internal/domain/program"
	"math-ai.com/math-ai/internal/shared/enum"
)

// GetProgramByIdQuery returns the program row with its full
// translation set attached. The base read uses Language for the LEFT
// JOIN override; the translation list is independent of Language so the
// caller always sees every defined locale.
type GetProgramByIdQuery struct {
	ProgramID int64
	Language  enum.LanguageType
}

type GetProgramByIdQueryHandler struct {
	programRepo     program.IRepository
	translationRepo program.ITranslationRepository
}

func NewGetProgramByIdQueryHandler(programRepo program.IRepository, translationRepo program.ITranslationRepository) *GetProgramByIdQueryHandler {
	return &GetProgramByIdQueryHandler{
		programRepo:     programRepo,
		translationRepo: translationRepo,
	}
}

func (h *GetProgramByIdQueryHandler) Handle(ctx context.Context, q GetProgramByIdQuery) (*program.Program, error) {
	p, err := h.programRepo.FindByProgramId(ctx, q.ProgramID, q.Language)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, nil
	}
	translations, err := h.translationRepo.ListByProgramId(ctx, q.ProgramID)
	if err != nil {
		return nil, err
	}
	p.SetTranslations(translations)
	return p, nil
}
