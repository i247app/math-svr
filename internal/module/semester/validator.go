package semester

import (
	"context"
	"errors"
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
			errors.New("language must be 'vn' or 'en'"))
	}
}

// validateTranslationLanguage is intentionally permissive on raw values —
// it accepts any non-empty short string. Keeps the semester module open
// to new languages without requiring a code change here every time.
func validateTranslationLanguage(ctx context.Context, raw string) error {
	v := strings.TrimSpace(raw)
	if v == "" {
		return errs.NewError(ctx, status.SEMESTER_INVALID_TRANSLATION, nil,
			errors.New("translation language is required"))
	}
	if len(v) > 10 {
		return errs.NewError(ctx, status.SEMESTER_INVALID_TRANSLATION, nil,
			errors.New("translation language is too long"))
	}
	return nil
}

func validateTranslations(ctx context.Context, ts []*dto.SemesterTranslationDTO) error {
	seen := make(map[string]struct{}, len(ts))
	for i, t := range ts {
		if t == nil {
			return errs.NewError(ctx, status.SEMESTER_INVALID_TRANSLATION,
				map[string]any{"index": i}, errors.New("translation is nil"))
		}
		if err := validateTranslationLanguage(ctx, t.Language); err != nil {
			return err
		}
		lang := strings.ToLower(strings.TrimSpace(t.Language))
		if _, dup := seen[lang]; dup {
			return errs.NewError(ctx, status.SEMESTER_TRANSLATION_ALREADY_EXISTS,
				map[string]any{"language": lang},
				errors.New("duplicate translation language in payload"))
		}
		seen[lang] = struct{}{}
		if strings.TrimSpace(t.Name) == "" {
			return errs.NewError(ctx, status.SEMESTER_MISSING_NAME,
				map[string]any{"language": lang}, errors.New("translation name is required"))
		}
		if len(t.Name) > nameMaxLen {
			return errs.NewError(ctx, status.SEMESTER_INVALID_TRANSLATION,
				map[string]any{"language": lang, "field": "name"},
				errors.New("translation name too long"))
		}
	}
	return nil
}

func ValidateCreateSemester(ctx context.Context, req *dto.CreateSemesterReq) error {
	if strings.TrimSpace(req.Name) == "" {
		return errs.NewError(ctx, status.SEMESTER_MISSING_NAME, nil,
			errors.New("name is required"))
	}
	if len(req.Name) > nameMaxLen {
		return errs.NewError(ctx, status.SEMESTER_MISSING_NAME, nil,
			errors.New("name too long"))
	}
	if req.DisplayOrder < 0 {
		return errs.NewError(ctx, status.SEMESTER_INVALID_DISPLAY_ORDER, nil,
			errors.New("display_order must be >= 0"))
	}
	return validateTranslations(ctx, req.Translations)
}

func ValidateUpdateSemester(ctx context.Context, req *dto.UpdateSemesterReq) error {
	if strings.TrimSpace(req.SemesterID) == "" {
		return errs.NewError(ctx, status.SEMESTER_MISSING_ID, nil,
			errors.New("semester_id is required"))
	}
	if req.Name != nil {
		if strings.TrimSpace(*req.Name) == "" {
			return errs.NewError(ctx, status.SEMESTER_MISSING_NAME, nil,
				errors.New("name cannot be blank"))
		}
		if len(*req.Name) > nameMaxLen {
			return errs.NewError(ctx, status.SEMESTER_MISSING_NAME, nil,
				errors.New("name too long"))
		}
	}
	if req.DisplayOrder != nil && *req.DisplayOrder < 0 {
		return errs.NewError(ctx, status.SEMESTER_INVALID_DISPLAY_ORDER, nil,
			errors.New("display_order must be >= 0"))
	}
	return validateTranslations(ctx, req.Translations)
}

func ValidateGetSemester(ctx context.Context, req *dto.GetSemesterReq) error {
	if strings.TrimSpace(req.SemesterID) == "" {
		return errs.NewError(ctx, status.SEMESTER_MISSING_ID, nil,
			errors.New("semester_id is required"))
	}
	return validateLanguage(ctx, req.Language)
}

func ValidateListSemesters(ctx context.Context, req *dto.ListSemestersReq) error {
	return validateLanguage(ctx, req.Language)
}

func ValidateDeleteSemester(ctx context.Context, req *dto.DeleteSemesterReq) error {
	if strings.TrimSpace(req.SemesterID) == "" {
		return errs.NewError(ctx, status.SEMESTER_MISSING_ID, nil,
			errors.New("semester_id is required"))
	}
	return nil
}
