package program

import (
	"context"
	"strings"

	dto "math-ai.com/math-ai/internal/application/dto/program"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/enum"
)

// labelMaxLen / descriptionMaxLen mirror the column widths in
// migration 006 (label VARCHAR(128), description VARCHAR(128) on the
// parent; description VARCHAR(255) on the translation row, but we cap
// to the tighter side so a translation stays drop-in displayable next
// to the base column).
const (
	labelMaxLen       = 128
	descriptionMaxLen = 128
)

func validateLanguage(ctx context.Context, lang enum.LanguageType) error {
	if lang == "" {
		return nil
	}
	switch lang {
	case enum.LanguageTypeVietnamese, enum.LanguageTypeEnglish:
		return nil
	default:
		return errs.NewError(ctx, status.PROGRAM_INVALID_LANGUAGE, nil,
			ErrLanguageMustBeVnOrEn)
	}
}

// validateTranslationLanguage is intentionally permissive on raw values —
// it accepts any non-empty short string. Keeps the program module open
// to new languages without requiring a code change here every time.
func validateTranslationLanguage(ctx context.Context, raw string) error {
	v := strings.TrimSpace(raw)
	if v == "" {
		return errs.NewError(ctx, status.PROGRAM_INVALID_TRANSLATION, nil,
			ErrTranslationLanguageRequired)
	}
	if len(v) > 10 {
		return errs.NewError(ctx, status.PROGRAM_INVALID_TRANSLATION, nil,
			ErrTranslationLanguageTooLong)
	}
	return nil
}

func validateTranslations(ctx context.Context, ts []*dto.ProgramTranslationDTO) error {
	seen := make(map[string]struct{}, len(ts))
	for i, t := range ts {
		if t == nil {
			return errs.NewError(ctx, status.PROGRAM_INVALID_TRANSLATION,
				map[string]any{"index": i}, ErrTranslationNil)
		}
		if err := validateTranslationLanguage(ctx, t.Language); err != nil {
			return err
		}
		lang := strings.ToLower(strings.TrimSpace(t.Language))
		if _, dup := seen[lang]; dup {
			return errs.NewError(ctx, status.PROGRAM_TRANSLATION_ALREADY_EXISTS,
				map[string]any{"language": lang},
				ErrDuplicateTranslationLanguageInPayload)
		}
		seen[lang] = struct{}{}
		if strings.TrimSpace(t.Label) == "" {
			return errs.NewError(ctx, status.PROGRAM_MISSING_LABEL,
				map[string]any{"language": lang}, ErrTranslationLabelRequired)
		}
		if len(t.Label) > labelMaxLen {
			return errs.NewError(ctx, status.PROGRAM_INVALID_TRANSLATION,
				map[string]any{"language": lang, "field": "label"},
				ErrTranslationLabelTooLong)
		}
		if strings.TrimSpace(t.Description) == "" {
			return errs.NewError(ctx, status.PROGRAM_MISSING_DESCRIPTION,
				map[string]any{"language": lang}, ErrTranslationDescriptionRequired)
		}
		if len(t.Description) > descriptionMaxLen {
			return errs.NewError(ctx, status.PROGRAM_INVALID_TRANSLATION,
				map[string]any{"language": lang, "field": "description"},
				ErrTranslationDescriptionTooLong)
		}
	}
	return nil
}

func ValidateCreateProgram(ctx context.Context, req *dto.CreateProgramReq) error {
	if strings.TrimSpace(req.Label) == "" {
		return errs.NewError(ctx, status.PROGRAM_MISSING_LABEL, nil,
			ErrLabelRequired)
	}
	if len(req.Label) > labelMaxLen {
		return errs.NewError(ctx, status.PROGRAM_MISSING_LABEL, nil,
			ErrLabelTooLong)
	}
	if strings.TrimSpace(req.Description) == "" {
		return errs.NewError(ctx, status.PROGRAM_MISSING_DESCRIPTION, nil,
			ErrDescriptionRequired)
	}
	if len(req.Description) > descriptionMaxLen {
		return errs.NewError(ctx, status.PROGRAM_MISSING_DESCRIPTION, nil,
			ErrDescriptionTooLong)
	}
	if req.DisplayOrder < 0 {
		return errs.NewError(ctx, status.PROGRAM_INVALID_DISPLAY_ORDER, nil,
			ErrDisplayOrderMustBe0)
	}
	return validateTranslations(ctx, req.Translations)
}

func ValidateUpdateProgram(ctx context.Context, req *dto.UpdateProgramReq) error {
	if req.ProgramID == 0 {
		return errs.NewError(ctx, status.PROGRAM_MISSING_ID, nil,
			ErrProgramIDRequired)
	}
	if req.Label != nil {
		if strings.TrimSpace(*req.Label) == "" {
			return errs.NewError(ctx, status.PROGRAM_MISSING_LABEL, nil,
				ErrLabelCannotBeBlank)
		}
		if len(*req.Label) > labelMaxLen {
			return errs.NewError(ctx, status.PROGRAM_MISSING_LABEL, nil,
				ErrLabelTooLong)
		}
	}
	if req.Description != nil {
		if strings.TrimSpace(*req.Description) == "" {
			return errs.NewError(ctx, status.PROGRAM_MISSING_DESCRIPTION, nil,
				ErrDescriptionCannotBeBlank)
		}
		if len(*req.Description) > descriptionMaxLen {
			return errs.NewError(ctx, status.PROGRAM_MISSING_DESCRIPTION, nil,
				ErrDescriptionTooLong)
		}
	}
	if req.DisplayOrder != nil && *req.DisplayOrder < 0 {
		return errs.NewError(ctx, status.PROGRAM_INVALID_DISPLAY_ORDER, nil,
			ErrDisplayOrderMustBe0)
	}
	return validateTranslations(ctx, req.Translations)
}

func ValidateGetProgram(ctx context.Context, req *dto.GetProgramReq) error {
	if req.ProgramID == 0 {
		return errs.NewError(ctx, status.PROGRAM_MISSING_ID, nil,
			ErrProgramIDRequired)
	}
	return validateLanguage(ctx, req.Language)
}

func ValidateListPrograms(ctx context.Context, req *dto.ListProgramsReq) error {
	return validateLanguage(ctx, req.Language)
}

func ValidateDeleteProgram(ctx context.Context, req *dto.DeleteProgramReq) error {
	if req.ProgramID == 0 {
		return errs.NewError(ctx, status.PROGRAM_MISSING_ID, nil,
			ErrProgramIDRequired)
	}
	return nil
}
