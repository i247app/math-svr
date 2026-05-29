package school

import (
	"context"
	"errors"
	"strings"

	dto "math-ai.com/math-ai/internal/application/dto/school"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
)

// Column-width mirrors of migration 017. Keeping the bounds here lets
// the validator reject over-long input with a friendly status code
// instead of letting MySQL truncate.
const (
	nameMaxLen        = 100
	descriptionMaxLen = 65535 // TEXT — a sanity cap, not a column limit
	noteMaxLen        = 500
	imageKeyMaxLen    = 128
	districtMaxLen    = 100
	provinceMaxLen    = 100
)

func ValidateCreateSchool(ctx context.Context, req *dto.CreateSchoolReq) error {
	if strings.TrimSpace(req.Name) == "" {
		return errs.NewError(ctx, status.SCHOOL_MISSING_NAME, nil,
			errors.New("name is required"))
	}
	if len(req.Name) > nameMaxLen {
		return errs.NewError(ctx, status.SCHOOL_NAME_TOO_LONG, nil,
			errors.New("name too long"))
	}
	if req.Description != nil && len(*req.Description) > descriptionMaxLen {
		return errs.NewError(ctx, status.SCHOOL_DESCRIPTION_TOO_LONG, nil,
			errors.New("description too long"))
	}
	if req.ImageKey != nil && len(*req.ImageKey) > imageKeyMaxLen {
		return errs.NewError(ctx, status.FAIL, nil,
			errors.New("image_key too long"))
	}
	if req.District != nil && len(*req.District) > districtMaxLen {
		return errs.NewError(ctx, status.FAIL, nil,
			errors.New("district too long"))
	}
	if req.Province != nil && len(*req.Province) > provinceMaxLen {
		return errs.NewError(ctx, status.FAIL, nil,
			errors.New("province too long"))
	}
	if req.Note != nil && len(*req.Note) > noteMaxLen {
		return errs.NewError(ctx, status.FAIL, nil,
			errors.New("note too long"))
	}
	return nil
}

func ValidateUpdateSchool(ctx context.Context, req *dto.UpdateSchoolReq) error {
	if strings.TrimSpace(req.SchoolID) == "" {
		return errs.NewError(ctx, status.SCHOOL_MISSING_ID, nil,
			errors.New("school_id is required"))
	}
	if req.Name != nil {
		if strings.TrimSpace(*req.Name) == "" {
			return errs.NewError(ctx, status.SCHOOL_MISSING_NAME, nil,
				errors.New("name cannot be blank"))
		}
		if len(*req.Name) > nameMaxLen {
			return errs.NewError(ctx, status.SCHOOL_NAME_TOO_LONG, nil,
				errors.New("name too long"))
		}
	}
	if req.Description != nil && len(*req.Description) > descriptionMaxLen {
		return errs.NewError(ctx, status.SCHOOL_DESCRIPTION_TOO_LONG, nil,
			errors.New("description too long"))
	}
	if req.ImageKey != nil && len(*req.ImageKey) > imageKeyMaxLen {
		return errs.NewError(ctx, status.FAIL, nil,
			errors.New("image_key too long"))
	}
	if req.District != nil && len(*req.District) > districtMaxLen {
		return errs.NewError(ctx, status.FAIL, nil,
			errors.New("district too long"))
	}
	if req.Province != nil && len(*req.Province) > provinceMaxLen {
		return errs.NewError(ctx, status.FAIL, nil,
			errors.New("province too long"))
	}
	if req.Note != nil && len(*req.Note) > noteMaxLen {
		return errs.NewError(ctx, status.FAIL, nil,
			errors.New("note too long"))
	}
	return nil
}

func ValidateGetSchool(ctx context.Context, req *dto.GetSchoolReq) error {
	if strings.TrimSpace(req.SchoolID) == "" {
		return errs.NewError(ctx, status.SCHOOL_MISSING_ID, nil,
			errors.New("school_id is required"))
	}
	return nil
}

func ValidateListSchools(ctx context.Context, req *dto.ListSchoolsReq) error {
	// Empty/blank filters collapse to nil so the repo skips the predicate.
	if req.Search != nil && strings.TrimSpace(*req.Search) == "" {
		req.Search = nil
	}
	if req.District != nil && strings.TrimSpace(*req.District) == "" {
		req.District = nil
	}
	if req.Province != nil && strings.TrimSpace(*req.Province) == "" {
		req.Province = nil
	}
	return nil
}

func ValidateDeleteSchool(ctx context.Context, req *dto.DeleteSchoolReq) error {
	if strings.TrimSpace(req.SchoolID) == "" {
		return errs.NewError(ctx, status.SCHOOL_MISSING_ID, nil,
			errors.New("school_id is required"))
	}
	return nil
}
