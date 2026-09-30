package permission

import (
	"context"
	"regexp"
	"strings"
	"unicode/utf8"

	dto "math-ai.com/math-ai/internal/application/dto/role"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/enum"
)

// Column widths of migrations/up/030_ma_roles.sql, counted in characters as
// MySQL counts VARCHAR — a Vietnamese name is longer in bytes than it looks.
const (
	roleNameMaxLen    = 128
	descriptionMaxLen = 500
	noteMaxLen        = 500
)

// roleCodePattern: the code is a machine key (compared, never displayed), so
// it is kept to one spelling — upper-case, starts with a letter, ≤ 64 chars.
var roleCodePattern = regexp.MustCompile(`^[A-Z][A-Z0-9_]{0,63}$`)

// ValidateCreateRole normalises req in place (role_code upper-cased, name
// trimmed) before checking it.
func ValidateCreateRole(ctx context.Context, req *dto.CreateRoleReq) error {
	req.RoleCode = strings.ToUpper(strings.TrimSpace(req.RoleCode))
	if req.RoleCode == "" {
		return errs.NewError(ctx, status.ROLE_MISSING_CODE, nil, ErrRoleCodeRequired)
	}
	if !roleCodePattern.MatchString(req.RoleCode) {
		return errs.NewError(ctx, status.ROLE_INVALID_CODE, nil, ErrInvalidRoleCode)
	}
	req.RoleName = strings.TrimSpace(req.RoleName)
	if err := validateRoleName(ctx, req.RoleName); err != nil {
		return err
	}
	return validateRoleOptionalFields(ctx, req.Description, req.Note, req.RoleStatus)
}

func ValidateUpdateRole(ctx context.Context, req *dto.UpdateRoleReq) error {
	if req.RoleID <= 0 {
		return errs.NewError(ctx, status.ROLE_MISSING_ID, nil, ErrRoleIDRequired)
	}
	if req.RemoveRoleImage && req.RoleImageFile != nil {
		return errs.NewError(ctx, status.ROLE_IMAGE_CONFLICT, nil, ErrRoleImageConflict)
	}
	if req.RoleName != nil {
		name := strings.TrimSpace(*req.RoleName)
		if err := validateRoleName(ctx, name); err != nil {
			return err
		}
		req.RoleName = &name
	}
	return validateRoleOptionalFields(ctx, req.Description, req.Note, req.RoleStatus)
}

func validateRoleName(ctx context.Context, name string) error {
	if name == "" {
		return errs.NewError(ctx, status.ROLE_MISSING_NAME, nil, ErrRoleNameRequired)
	}
	if utf8.RuneCountInString(name) > roleNameMaxLen {
		return errs.NewError(ctx, status.ROLE_NAME_TOO_LONG, nil, ErrRoleNameTooLong)
	}
	return nil
}

// validateRoleOptionalFields checks the fields shared by create and update;
// a nil pointer skips its check.
func validateRoleOptionalFields(ctx context.Context, description, note, roleStatus *string) error {
	if description != nil && utf8.RuneCountInString(*description) > descriptionMaxLen {
		return errs.NewError(ctx, status.ROLE_DESCRIPTION_TOO_LONG, nil, ErrDescriptionTooLong)
	}
	if note != nil && utf8.RuneCountInString(*note) > noteMaxLen {
		return errs.NewError(ctx, status.ROLE_NOTE_TOO_LONG, nil, ErrNoteTooLong)
	}
	if roleStatus != nil && !enum.RoleRecordStatusType(*roleStatus).IsValid() {
		return errs.NewError(ctx, status.ROLE_INVALID_STATUS, nil, ErrInvalidRoleStatus)
	}
	return nil
}

func ValidateGetRole(ctx context.Context, req *dto.GetRoleReq) error {
	if req.RoleID <= 0 {
		return errs.NewError(ctx, status.ROLE_MISSING_ID, nil, ErrRoleIDRequired)
	}
	return nil
}

func ValidateDeleteRole(ctx context.Context, req *dto.DeleteRoleReq) error {
	if req.RoleID <= 0 {
		return errs.NewError(ctx, status.ROLE_MISSING_ID, nil, ErrRoleIDRequired)
	}
	return nil
}

// ValidateListRoles collapses blank filters to nil so the repo skips them,
// and drops non-positive / duplicate ids.
func ValidateListRoles(ctx context.Context, req *dto.ListRolesReq) error {
	if req.Search != nil && strings.TrimSpace(*req.Search) == "" {
		req.Search = nil
	}
	if req.RoleStatus != nil {
		trimmed := strings.TrimSpace(*req.RoleStatus)
		if trimmed == "" {
			req.RoleStatus = nil
		} else if !enum.RoleRecordStatusType(trimmed).IsValid() {
			return errs.NewError(ctx, status.ROLE_INVALID_STATUS, nil, ErrInvalidRoleStatus)
		}
	}
	req.RoleIDs = sanitizeRoleIDs(req.RoleIDs)
	return nil
}

// sanitizeRoleIDs returns nil for an all-invalid input so the repo treats it
// as no filter rather than an IN(...) that matches nothing.
func sanitizeRoleIDs(ids []int64) []int64 {
	seen := make(map[int64]struct{}, len(ids))
	var out []int64
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
	return out
}
