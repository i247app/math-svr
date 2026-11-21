package semester

import (
	"context"
	"strings"

	"math-ai.com/math-ai/internal/application/dto"
	domain "math-ai.com/math-ai/internal/domain/semester"
	"math-ai.com/math-ai/internal/shared/errs"
)

// nameMaxLen mirrors the column width in migration 008
// (name VARCHAR(100)). description is TEXT — no length cap.
const nameMaxLen = 100

func ValidateCreateSemester(ctx context.Context, req *dto.CreateSemesterReq) error {
	if strings.TrimSpace(req.Name) == "" {
		return errs.NewError(ctx, domain.SemesterMissingName, nil,
			ErrNameRequired)
	}
	if len(req.Name) > nameMaxLen {
		return errs.NewError(ctx, domain.SemesterMissingName, nil,
			ErrNameTooLong)
	}
	if req.DisplayOrder < 0 {
		return errs.NewError(ctx, domain.SemesterInvalidDisplayOrder, nil,
			ErrDisplayOrderMustBe0)
	}
	return nil
}

func ValidateUpdateSemester(ctx context.Context, req *dto.UpdateSemesterReq) error {
	if req.SemesterID == 0 {
		return errs.NewError(ctx, domain.SemesterMissingID, nil,
			ErrSemesterIDRequired)
	}
	if req.Name != nil {
		if strings.TrimSpace(*req.Name) == "" {
			return errs.NewError(ctx, domain.SemesterMissingName, nil,
				ErrNameCannotBeBlank)
		}
		if len(*req.Name) > nameMaxLen {
			return errs.NewError(ctx, domain.SemesterMissingName, nil,
				ErrNameTooLong)
		}
	}
	if req.DisplayOrder != nil && *req.DisplayOrder < 0 {
		return errs.NewError(ctx, domain.SemesterInvalidDisplayOrder, nil,
			ErrDisplayOrderMustBe0)
	}
	return nil
}

func ValidateGetSemester(ctx context.Context, req *dto.GetSemesterReq) error {
	if req.SemesterID == 0 {
		return errs.NewError(ctx, domain.SemesterMissingID, nil,
			ErrSemesterIDRequired)
	}
	return nil
}

func ValidateListSemesters(ctx context.Context, req *dto.ListSemestersReq) error {
	return nil
}

func ValidateDeleteSemester(ctx context.Context, req *dto.DeleteSemesterReq) error {
	if req.SemesterID == 0 {
		return errs.NewError(ctx, domain.SemesterMissingID, nil,
			ErrSemesterIDRequired)
	}
	return nil
}
