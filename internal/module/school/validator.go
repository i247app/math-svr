package school

import (
	"context"
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
			ErrNameRequired)
	}
	if len(req.Name) > nameMaxLen {
		return errs.NewError(ctx, status.SCHOOL_NAME_TOO_LONG, nil,
			ErrNameTooLong)
	}
	if req.Description != nil && len(*req.Description) > descriptionMaxLen {
		return errs.NewError(ctx, status.SCHOOL_DESCRIPTION_TOO_LONG, nil,
			ErrDescriptionTooLong)
	}
	if req.ImageKey != nil && len(*req.ImageKey) > imageKeyMaxLen {
		return errs.NewError(ctx, status.FAIL, nil,
			ErrImageKeyTooLong)
	}
	if req.District != nil && len(*req.District) > districtMaxLen {
		return errs.NewError(ctx, status.FAIL, nil,
			ErrDistrictTooLong)
	}
	if req.Province != nil && len(*req.Province) > provinceMaxLen {
		return errs.NewError(ctx, status.FAIL, nil,
			ErrProvinceTooLong)
	}
	if req.Note != nil && len(*req.Note) > noteMaxLen {
		return errs.NewError(ctx, status.FAIL, nil,
			ErrNoteTooLong)
	}
	return nil
}

func ValidateUpdateSchool(ctx context.Context, req *dto.UpdateSchoolReq) error {
	if req.SchoolID == 0 {
		return errs.NewError(ctx, status.SCHOOL_MISSING_ID, nil,
			ErrSchoolIDRequired)
	}
	if req.Name != nil {
		if strings.TrimSpace(*req.Name) == "" {
			return errs.NewError(ctx, status.SCHOOL_MISSING_NAME, nil,
				ErrNameCannotBeBlank)
		}
		if len(*req.Name) > nameMaxLen {
			return errs.NewError(ctx, status.SCHOOL_NAME_TOO_LONG, nil,
				ErrNameTooLong)
		}
	}
	if req.Description != nil && len(*req.Description) > descriptionMaxLen {
		return errs.NewError(ctx, status.SCHOOL_DESCRIPTION_TOO_LONG, nil,
			ErrDescriptionTooLong)
	}
	if req.ImageKey != nil && len(*req.ImageKey) > imageKeyMaxLen {
		return errs.NewError(ctx, status.FAIL, nil,
			ErrImageKeyTooLong)
	}
	if req.District != nil && len(*req.District) > districtMaxLen {
		return errs.NewError(ctx, status.FAIL, nil,
			ErrDistrictTooLong)
	}
	if req.Province != nil && len(*req.Province) > provinceMaxLen {
		return errs.NewError(ctx, status.FAIL, nil,
			ErrProvinceTooLong)
	}
	if req.Note != nil && len(*req.Note) > noteMaxLen {
		return errs.NewError(ctx, status.FAIL, nil,
			ErrNoteTooLong)
	}
	return nil
}

func ValidateGetSchool(ctx context.Context, req *dto.GetSchoolReq) error {
	if req.SchoolID == 0 {
		return errs.NewError(ctx, status.SCHOOL_MISSING_ID, nil,
			ErrSchoolIDRequired)
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
	req.SchoolIDs = sanitizeSchoolIDs(req.SchoolIDs)
	return nil
}

// sanitizeSchoolIDs drops non-positive ids and removes duplicates while
// preserving caller order. Returns nil for an all-invalid input so the
// repo treats it as no filter rather than an empty IN(...) that would
// match zero rows.
func sanitizeSchoolIDs(ids []int64) []int64 {
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

func ValidateDeleteSchool(ctx context.Context, req *dto.DeleteSchoolReq) error {
	if req.SchoolID == 0 {
		return errs.NewError(ctx, status.SCHOOL_MISSING_ID, nil,
			ErrSchoolIDRequired)
	}
	return nil
}
