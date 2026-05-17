package grade

import (
	"context"
	"errors"

	dto "math-ai.com/math-ai/internal/application/dto/grade"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/enum"
)

func ValidateListGrades(ctx context.Context, req *dto.ListGradesReq) error {
	if req.Language == "" {
		return nil
	}
	switch req.Language {
	case enum.LanguageTypeVietnamese, enum.LanguageTypeEnglish:
		return nil
	default:
		return errs.NewError(ctx, status.GRADE_INVALID_LANGUAGE, nil,
			errors.New("grade language must be 'vi' or 'en'"))
	}
}
