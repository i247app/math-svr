package program

import (
	"context"
	"errors"

	dto "math-ai.com/math-ai/internal/application/dto/program"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/enum"
)

// ValidateListPrograms normalizes and bounds-checks the language input. An
// empty language is fine — the service substitutes the project default.
func ValidateListPrograms(ctx context.Context, req *dto.ListProgramsReq) error {
	if req.Language == "" {
		return nil
	}
	switch req.Language {
	case enum.LanguageTypeVietnamese, enum.LanguageTypeEnglish:
		return nil
	default:
		return errs.NewError(ctx, status.PROGRAM_INVALID_LANGUAGE, nil,
			errors.New("program language must be 'vi' or 'en'"))
	}
}
