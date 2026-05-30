package classroom

import (
	"context"
	"errors"
	"strings"

	dto "math-ai.com/math-ai/internal/application/dto/classroom"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
)

// Column-width mirrors for migration 014. Keeping the bounds here lets
// the validator reject over-long input with a friendly status code
// instead of letting MySQL truncate.
const (
	nameMaxLen        = 128
	descriptionMaxLen = 500
	noteMaxLen        = 500
	coverKeyMaxLen    = 256
	searchMaxLen      = 128
)

func ValidateCreateClassroom(ctx context.Context, req *dto.CreateClassroomReq) error {
	if strings.TrimSpace(req.ProfileID) == "" {
		return errs.NewError(ctx, status.CLASSROOM_MISSING_OWNER_PROFILE_ID, nil,
			errors.New("profile_id is required"))
	}
	if strings.TrimSpace(req.Name) == "" {
		return errs.NewError(ctx, status.CLASSROOM_MISSING_NAME, nil,
			errors.New("name is required"))
	}
	if len(req.Name) > nameMaxLen {
		return errs.NewError(ctx, status.CLASSROOM_NAME_TOO_LONG, nil,
			errors.New("name too long"))
	}
	if req.Description != nil && len(*req.Description) > descriptionMaxLen {
		return errs.NewError(ctx, status.CLASSROOM_DESCRIPTION_TOO_LONG, nil,
			errors.New("description too long"))
	}
	if req.CoverKey != nil && len(*req.CoverKey) > coverKeyMaxLen {
		return errs.NewError(ctx, status.FAIL, nil,
			errors.New("cover_key too long"))
	}
	if req.Note != nil && len(*req.Note) > noteMaxLen {
		return errs.NewError(ctx, status.FAIL, nil,
			errors.New("note too long"))
	}
	if req.MaxMembers != nil && *req.MaxMembers <= 0 {
		return errs.NewError(ctx, status.CLASSROOM_INVALID_MAX_MEMBERS, nil,
			errors.New("max_members must be > 0"))
	}
	return nil
}

func ValidateUpdateClassroom(ctx context.Context, req *dto.UpdateClassroomReq) error {
	if strings.TrimSpace(req.ProfileID) == "" {
		return errs.NewError(ctx, status.CLASSROOM_MISSING_OWNER_PROFILE_ID, nil,
			errors.New("profile_id is required"))
	}
	if strings.TrimSpace(req.ClassroomID) == "" {
		return errs.NewError(ctx, status.CLASSROOM_MISSING_ID, nil,
			errors.New("classroom_id is required"))
	}
	if req.Name != nil {
		if strings.TrimSpace(*req.Name) == "" {
			return errs.NewError(ctx, status.CLASSROOM_MISSING_NAME, nil,
				errors.New("name cannot be blank"))
		}
		if len(*req.Name) > nameMaxLen {
			return errs.NewError(ctx, status.CLASSROOM_NAME_TOO_LONG, nil,
				errors.New("name too long"))
		}
	}
	if req.Description != nil && len(*req.Description) > descriptionMaxLen {
		return errs.NewError(ctx, status.CLASSROOM_DESCRIPTION_TOO_LONG, nil,
			errors.New("description too long"))
	}
	if req.CoverKey != nil && len(*req.CoverKey) > coverKeyMaxLen {
		return errs.NewError(ctx, status.FAIL, nil,
			errors.New("cover_key too long"))
	}
	if req.Note != nil && len(*req.Note) > noteMaxLen {
		return errs.NewError(ctx, status.FAIL, nil,
			errors.New("note too long"))
	}
	if req.MaxMembers != nil && *req.MaxMembers <= 0 {
		return errs.NewError(ctx, status.CLASSROOM_INVALID_MAX_MEMBERS, nil,
			errors.New("max_members must be > 0"))
	}
	return nil
}

func ValidateGetClassroom(ctx context.Context, req *dto.GetClassroomReq) error {
	if strings.TrimSpace(req.ProfileID) == "" {
		return errs.NewError(ctx, status.CLASSROOM_MISSING_OWNER_PROFILE_ID, nil,
			errors.New("profile_id is required"))
	}
	if strings.TrimSpace(req.ClassroomID) == "" {
		return errs.NewError(ctx, status.CLASSROOM_MISSING_ID, nil,
			errors.New("classroom_id is required"))
	}
	return nil
}

func ValidateListClassrooms(ctx context.Context, req *dto.ListClassroomsReq) error {
	if strings.TrimSpace(req.ProfileID) == "" {
		return errs.NewError(ctx, status.CLASSROOM_MISSING_OWNER_PROFILE_ID, nil,
			errors.New("profile_id is required"))
	}
	// Collapse blank optional filters to nil so the repo skips the predicate.
	if req.OwnerProfileID != nil && strings.TrimSpace(*req.OwnerProfileID) == "" {
		req.OwnerProfileID = nil
	}
	if req.ProgramID != nil && strings.TrimSpace(*req.ProgramID) == "" {
		req.ProgramID = nil
	}
	if req.GradeID != nil && strings.TrimSpace(*req.GradeID) == "" {
		req.GradeID = nil
	}
	if req.Search != nil {
		s := strings.TrimSpace(*req.Search)
		if s == "" {
			req.Search = nil
		} else if len(s) > searchMaxLen {
			return errs.NewError(ctx, status.FAIL, nil,
				errors.New("search too long"))
		}
	}
	return nil
}

func ValidateArchiveClassroom(ctx context.Context, req *dto.ArchiveClassroomReq) error {
	if strings.TrimSpace(req.ProfileID) == "" {
		return errs.NewError(ctx, status.CLASSROOM_MISSING_OWNER_PROFILE_ID, nil,
			errors.New("profile_id is required"))
	}
	if strings.TrimSpace(req.ClassroomID) == "" {
		return errs.NewError(ctx, status.CLASSROOM_MISSING_ID, nil,
			errors.New("classroom_id is required"))
	}
	return nil
}

func ValidateRestoreClassroom(ctx context.Context, req *dto.RestoreClassroomReq) error {
	if strings.TrimSpace(req.ProfileID) == "" {
		return errs.NewError(ctx, status.CLASSROOM_MISSING_OWNER_PROFILE_ID, nil,
			errors.New("profile_id is required"))
	}
	if strings.TrimSpace(req.ClassroomID) == "" {
		return errs.NewError(ctx, status.CLASSROOM_MISSING_ID, nil,
			errors.New("classroom_id is required"))
	}
	return nil
}

func ValidateDeleteClassroom(ctx context.Context, req *dto.DeleteClassroomReq) error {
	if strings.TrimSpace(req.ProfileID) == "" {
		return errs.NewError(ctx, status.CLASSROOM_MISSING_OWNER_PROFILE_ID, nil,
			errors.New("profile_id is required"))
	}
	if strings.TrimSpace(req.ClassroomID) == "" {
		return errs.NewError(ctx, status.CLASSROOM_MISSING_ID, nil,
			errors.New("classroom_id is required"))
	}
	return nil
}
