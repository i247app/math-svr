package semester

import (
	"context"
	"strings"

	dto "math-ai.com/math-ai/internal/application/dto/semester"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/enum"
)

// nameMaxLen mirrors the column width in migration 008
// (name VARCHAR(100)). description is TEXT — no length cap.
const nameMaxLen = 100

func validateLanguage(ctx context.Context, lang enum.LanguageType) error {
	if lang == "" {
		return nil
	}
	switch lang {
	case enum.LanguageTypeVietnamese, enum.LanguageTypeEnglish:
		return nil
	default:
		return errs.NewError(ctx, status.SEMESTER_INVALID_LANGUAGE, nil,
			ErrLanguageMustBeVnOrEn)
	}
}

// validateTranslationLanguage is intentionally permissive on raw values —
// it accepts any non-empty short string. Keeps the semester module open
// to new languages without requiring a code change here every time.
func validateTranslationLanguage(ctx context.Context, raw string) error {
	v := strings.TrimSpace(raw)
	if v == "" {
		return errs.NewError(ctx, status.SEMESTER_INVALID_TRANSLATION, nil,
			ErrTranslationLanguageRequired)
	}
	if len(v) > 10 {
		return errs.NewError(ctx, status.SEMESTER_INVALID_TRANSLATION, nil,
			ErrTranslationLanguageTooLong)
	}
	return nil
}

func validateTranslations(ctx context.Context, ts []*dto.SemesterTranslationDTO) error {
	seen := make(map[string]struct{}, len(ts))
	for i, t := range ts {
		if t == nil {
			return errs.NewError(ctx, status.SEMESTER_INVALID_TRANSLATION,
				map[string]any{"index": i}, ErrTranslationNil)
		}
		if err := validateTranslationLanguage(ctx, t.Language); err != nil {
			return err
		}
		lang := strings.ToLower(strings.TrimSpace(t.Language))
		if _, dup := seen[lang]; dup {
			return errs.NewError(ctx, status.SEMESTER_TRANSLATION_ALREADY_EXISTS,
				map[string]any{"language": lang},
				ErrDuplicateTranslationLanguageInPayload)
		}
		seen[lang] = struct{}{}
		if strings.TrimSpace(t.Name) == "" {
			return errs.NewError(ctx, status.SEMESTER_MISSING_NAME,
				map[string]any{"language": lang}, ErrTranslationNameRequired)
		}
		if len(t.Name) > nameMaxLen {
			return errs.NewError(ctx, status.SEMESTER_INVALID_TRANSLATION,
				map[string]any{"language": lang, "field": "name"},
				ErrTranslationNameTooLong)
		}
	}
	return nil
}

func ValidateCreateSemester(ctx context.Context, req *dto.CreateSemesterReq) error {
	if strings.TrimSpace(req.Name) == "" {
		return errs.NewError(ctx, status.SEMESTER_MISSING_NAME, nil,
			ErrNameRequired)
	}
	if len(req.Name) > nameMaxLen {
		return errs.NewError(ctx, status.SEMESTER_MISSING_NAME, nil,
			ErrNameTooLong)
	}
	if req.DisplayOrder < 0 {
		return errs.NewError(ctx, status.SEMESTER_INVALID_DISPLAY_ORDER, nil,
			ErrDisplayOrderMustBe0)
	}
	return validateTranslations(ctx, req.Translations)
}

func ValidateUpdateSemester(ctx context.Context, req *dto.UpdateSemesterReq) error {
	if req.SemesterID == 0 {
		return errs.NewError(ctx, status.SEMESTER_MISSING_ID, nil,
			ErrSemesterIDRequired)
	}
	if req.Name != nil {
		if strings.TrimSpace(*req.Name) == "" {
			return errs.NewError(ctx, status.SEMESTER_MISSING_NAME, nil,
				ErrNameCannotBeBlank)
		}
		if len(*req.Name) > nameMaxLen {
			return errs.NewError(ctx, status.SEMESTER_MISSING_NAME, nil,
				ErrNameTooLong)
		}
	}
	if req.DisplayOrder != nil && *req.DisplayOrder < 0 {
		return errs.NewError(ctx, status.SEMESTER_INVALID_DISPLAY_ORDER, nil,
			ErrDisplayOrderMustBe0)
	}
	return validateTranslations(ctx, req.Translations)
}

func ValidateGetSemester(ctx context.Context, req *dto.GetSemesterReq) error {
	if req.SemesterID == 0 {
		return errs.NewError(ctx, status.SEMESTER_MISSING_ID, nil,
			ErrSemesterIDRequired)
	}
	return validateLanguage(ctx, req.Language)
}

func ValidateListSemesters(ctx context.Context, req *dto.ListSemestersReq) error {
	return validateLanguage(ctx, req.Language)
}

func ValidateDeleteSemester(ctx context.Context, req *dto.DeleteSemesterReq) error {
	if req.SemesterID == 0 {
		return errs.NewError(ctx, status.SEMESTER_MISSING_ID, nil,
			ErrSemesterIDRequired)
	}
	return nil
}
