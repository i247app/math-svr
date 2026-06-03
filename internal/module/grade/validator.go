package grade

import (
	"context"
	"strings"

	dto "math-ai.com/math-ai/internal/application/dto/grade"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/enum"
)

// labelMaxLen / descriptionMaxLen mirror the column widths in
// migration 007 (label VARCHAR(128), description VARCHAR(128) on the
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
		return errs.NewError(ctx, status.GRADE_INVALID_LANGUAGE, nil,
			ErrLanguageMustBeVnOrEn)
	}
}

// validateTranslationLanguage is intentionally permissive on raw values —
// it accepts any non-empty short string. Keeps the grade module open
// to new languages without requiring a code change here every time.
func validateTranslationLanguage(ctx context.Context, raw string) error {
	v := strings.TrimSpace(raw)
	if v == "" {
		return errs.NewError(ctx, status.GRADE_INVALID_TRANSLATION, nil, ErrTranslationLanguageRequired)
	}
	if len(v) > 10 {
		return errs.NewError(ctx, status.GRADE_INVALID_TRANSLATION, nil, ErrTranslationLanguageTooLong)
	}
	return nil
}

func validateTranslations(ctx context.Context, ts []*dto.GradeTranslationDTO) error {
	seen := make(map[string]struct{}, len(ts))
	for i, t := range ts {
		if t == nil {
			return errs.NewError(ctx, status.GRADE_INVALID_TRANSLATION, map[string]any{"index": i}, ErrTranslationNil)
		}
		if err := validateTranslationLanguage(ctx, t.Language); err != nil {
			return err
		}
		lang := strings.ToLower(strings.TrimSpace(t.Language))
		if _, dup := seen[lang]; dup {
			return errs.NewError(ctx, status.GRADE_TRANSLATION_ALREADY_EXISTS,
				map[string]any{"language": lang}, ErrDuplicateTranslationLanguageInPayload)
		}
		seen[lang] = struct{}{}
		if strings.TrimSpace(t.Label) == "" {
			return errs.NewError(ctx, status.GRADE_MISSING_LABEL, map[string]any{"language": lang}, ErrTranslationLabelRequired)
		}
		if len(t.Label) > labelMaxLen {
			return errs.NewError(ctx, status.GRADE_INVALID_TRANSLATION, map[string]any{"language": lang, "field": "label"}, ErrTranslationLabelTooLong)
		}
		if strings.TrimSpace(t.Description) == "" {
			return errs.NewError(ctx, status.GRADE_MISSING_DESCRIPTION, map[string]any{"language": lang}, ErrTranslationDescriptionRequired)
		}
		if len(t.Description) > descriptionMaxLen {
			return errs.NewError(ctx, status.GRADE_INVALID_TRANSLATION, map[string]any{"language": lang, "field": "description"}, ErrTranslationDescriptionTooLong)
		}
	}
	return nil
}

func ValidateCreateGrade(ctx context.Context, req *dto.CreateGradeReq) error {
	if strings.TrimSpace(req.Label) == "" {
		return errs.NewError(ctx, status.GRADE_MISSING_LABEL, nil, ErrLabelRequired)
	}
	if len(req.Label) > labelMaxLen {
		return errs.NewError(ctx, status.GRADE_MISSING_LABEL, nil, ErrLabelTooLong)
	}
	if strings.TrimSpace(req.Description) == "" {
		return errs.NewError(ctx, status.GRADE_MISSING_DESCRIPTION, nil, ErrDescriptionRequired)
	}
	if len(req.Description) > descriptionMaxLen {
		return errs.NewError(ctx, status.GRADE_MISSING_DESCRIPTION, nil, ErrDescriptionTooLong)
	}
	if req.DisplayOrder < 0 {
		return errs.NewError(ctx, status.GRADE_INVALID_DISPLAY_ORDER, nil, ErrDisplayOrderMustBe0)
	}
	return validateTranslations(ctx, req.Translations)
}

func ValidateUpdateGrade(ctx context.Context, req *dto.UpdateGradeReq) error {
	if req.GradeID == 0 {
		return errs.NewError(ctx, status.GRADE_MISSING_ID, nil, ErrGradeIDRequired)
	}
	if req.Label != nil {
		if strings.TrimSpace(*req.Label) == "" {
			return errs.NewError(ctx, status.GRADE_MISSING_LABEL, nil, ErrLabelCannotBeBlank)
		}
		if len(*req.Label) > labelMaxLen {
			return errs.NewError(ctx, status.GRADE_MISSING_LABEL, nil, ErrLabelTooLong)
		}
	}
	if req.Description != nil {
		if strings.TrimSpace(*req.Description) == "" {
			return errs.NewError(ctx, status.GRADE_MISSING_DESCRIPTION, nil, ErrDescriptionCannotBeBlank)
		}
		if len(*req.Description) > descriptionMaxLen {
			return errs.NewError(ctx, status.GRADE_MISSING_DESCRIPTION, nil, ErrDescriptionTooLong)
		}
	}
	if req.DisplayOrder != nil && *req.DisplayOrder < 0 {
		return errs.NewError(ctx, status.GRADE_INVALID_DISPLAY_ORDER, nil, ErrDisplayOrderMustBe0)
	}
	return validateTranslations(ctx, req.Translations)
}

func ValidateGetGrade(ctx context.Context, req *dto.GetGradeReq) error {
	if req.GradeID == 0 {
		return errs.NewError(ctx, status.GRADE_MISSING_ID, nil, ErrGradeIDRequired)
	}
	return validateLanguage(ctx, req.Language)
}

func ValidateListGrades(ctx context.Context, req *dto.ListGradesReq) error {
	if err := validateLanguage(ctx, req.Language); err != nil {
		return err
	}
	req.GradeIDs = sanitizeGradeIDs(req.GradeIDs)
	return nil
}

// sanitizeGradeIDs drops non-positive ids and removes duplicates while
// preserving caller order. Returns nil for an all-invalid input so the
// repo treats it as no filter rather than an empty IN(...) that would
// match zero rows.
func sanitizeGradeIDs(ids []int64) []int64 {
	if len(ids) == 0 {
		return nil
	}
	seen := make(map[int64]struct{}, len(ids))
	out := make([]int64, 0, len(ids))
	for _, id := range ids {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func ValidateDeleteGrade(ctx context.Context, req *dto.DeleteGradeReq) error {
	if req.GradeID == 0 {
		return errs.NewError(ctx, status.GRADE_MISSING_ID, nil, ErrGradeIDRequired)
	}
	return nil
}
