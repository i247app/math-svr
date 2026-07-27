package semester

import (
	"context"
	"strings"

	dto "math-ai.com/math-ai/internal/application/dto/semester"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
)

// nameMaxLen mirrors the column width in migration 008
// (name VARCHAR(100)). description is TEXT — no length cap.
const nameMaxLen = 100

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
	return nil
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
	return nil
}

func ValidateGetSemester(ctx context.Context, req *dto.GetSemesterReq) error {
	if req.SemesterID == 0 {
		return errs.NewError(ctx, status.SEMESTER_MISSING_ID, nil,
			ErrSemesterIDRequired)
	}
	return nil
}

func ValidateListSemesters(ctx context.Context, req *dto.ListSemestersReq) error {
	return nil
}

func ValidateDeleteSemester(ctx context.Context, req *dto.DeleteSemesterReq) error {
	if req.SemesterID == 0 {
		return errs.NewError(ctx, status.SEMESTER_MISSING_ID, nil,
			ErrSemesterIDRequired)
	}
	return nil
}
