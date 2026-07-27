package program

import (
	"context"
	"strings"

	dto "math-ai.com/math-ai/internal/application/dto/program"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
)

// labelMaxLen / descriptionMaxLen mirror the column widths in
// migration 006 (label VARCHAR(128), description VARCHAR(128)).
const (
	labelMaxLen       = 128
	descriptionMaxLen = 128
)

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
	return nil
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
	return nil
}

func ValidateGetProgram(ctx context.Context, req *dto.GetProgramReq) error {
	if req.ProgramID == 0 {
		return errs.NewError(ctx, status.PROGRAM_MISSING_ID, nil,
			ErrProgramIDRequired)
	}
	return nil
}

func ValidateListPrograms(ctx context.Context, req *dto.ListProgramsReq) error {
	return nil
}

func ValidateDeleteProgram(ctx context.Context, req *dto.DeleteProgramReq) error {
	if req.ProgramID == 0 {
		return errs.NewError(ctx, status.PROGRAM_MISSING_ID, nil,
			ErrProgramIDRequired)
	}
	return nil
}
