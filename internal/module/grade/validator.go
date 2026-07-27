package grade

import (
	"context"
	"strings"

	dto "math-ai.com/math-ai/internal/application/dto/grade"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
)

// labelMaxLen / descriptionMaxLen mirror the column widths in
// migration 007 (label VARCHAR(128), description VARCHAR(128)).
const (
	labelMaxLen       = 128
	descriptionMaxLen = 128
)

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
	return nil
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
	return nil
}

func ValidateGetGrade(ctx context.Context, req *dto.GetGradeReq) error {
	if req.GradeID == 0 {
		return errs.NewError(ctx, status.GRADE_MISSING_ID, nil, ErrGradeIDRequired)
	}
	return nil
}

func ValidateListGrades(ctx context.Context, req *dto.ListGradesReq) error {
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
