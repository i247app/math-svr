package user

import (
	"context"
	"slices"
	"strings"

	"math-ai.com/math-ai/internal/domain/login"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/enum"

	dto "math-ai.com/math-ai/internal/application/dto/user"
)

// ValidateCreateUser checks a self-registration (/users/create): the role,
// if any, must be one a person may give themselves — never ADMIN.
func ValidateCreateUser(ctx context.Context, req *dto.CreateUserReq) error {
	return validateCreateUserReq(ctx, req, enum.RoleType.IsSelfAssignable, enum.ListSelfAssignableRoles())
}

// ValidateAdminCreateUser checks /users/admin/create, where an admin
// registers someone else: any valid role is accepted, ADMIN included.
func ValidateAdminCreateUser(ctx context.Context, req *dto.CreateUserReq) error {
	return validateCreateUserReq(ctx, req, enum.RoleType.IsValid, enum.ListRoles())
}

func validateCreateUserReq(ctx context.Context, req *dto.CreateUserReq, roleAllowed func(enum.RoleType) bool, allowedRoles []string) error {
	// Phone and email are both login keys; either one alone is enough to
	// register, but an account with neither could never sign in.
	if strings.TrimSpace(req.Phone) == "" && strings.TrimSpace(req.Email) == "" {
		return errs.NewError(ctx, status.USER_MISSING_IDENTIFIER, nil, ErrPhoneOrEmailRequired)
	}
	if strings.TrimSpace(req.Name) == "" {
		return errs.NewError(ctx, status.USER_MISSING_NAME, nil, ErrNameRequired)
	}
	// role is optional on the wire — the create command defaults an empty
	// value to STUDENT (ma_users.role is NOT NULL). When supplied it must
	// pass roleAllowed. Normalised in place so the command sees the
	// trimmed token.
	role := strings.TrimSpace(req.Role)
	if role == "" {
		req.Role = ""
	} else {
		if !roleAllowed(enum.RoleType(role)) {
			args := map[string]any{
				"roles": allowedRoles,
			}
			return errs.NewError(ctx, status.USER_INVALID_ROLE, args, ErrRoleInvalid)
		}
		req.Role = role
	}
	// Avatar can come as a file upload OR a string reference, never both
	// — the two sources collide semantically and would force the service
	// to pick a winner silently. Format/host validity lives in the
	// service layer (normalizeAvatarKey).
	if strings.TrimSpace(req.Avatar) != "" && req.AvatarFile != nil {
		return errs.NewError(ctx, status.USER_AVATAR_CONFLICT, nil, ErrProvideEitherAvatarFileOrAvatarReference)
	}
	// Password is optional; only a supplied one is checked. It is not
	// trimmed — spaces are legitimate password characters.
	if req.Password != "" {
		if err := login.ValidatePassword(req.Password); err != nil {
			return errs.NewError(ctx, status.USER_INVALID_PASSWORD, nil, err)
		}
	}
	return nil
}

// ValidateListUsers checks the role filter of /users/list and normalises it
// in place: values are trimmed and de-duplicated, blanks dropped. Any role
// the system knows may be filtered on, ADMIN included — filtering is not
// assigning. An unknown role is refused rather than silently matching no one.
func ValidateListUsers(ctx context.Context, req *dto.ListUsersReq) error {
	roles := make([]string, 0, len(req.Roles))
	for _, r := range req.Roles {
		role := strings.TrimSpace(r)
		if role == "" || slices.Contains(roles, role) {
			continue
		}
		if !enum.RoleType(role).IsValid() {
			args := map[string]any{
				"roles": enum.ListRoles(),
			}
			return errs.NewError(ctx, status.USER_INVALID_ROLE, args, ErrRoleInvalid)
		}
		roles = append(roles, role)
	}
	req.Roles = roles
	return req.Validate(ctx)
}

func ValidateCheckIdentifier(ctx context.Context, req *dto.CheckIdentifierReq) error {
	req.Identifier = strings.TrimSpace(req.Identifier)
	if req.Identifier == "" {
		return errs.NewError(ctx, status.USER_MISSING_IDENTIFIER, nil, ErrPhoneOrEmailRequired)
	}
	return nil
}

func ValidateUpdateUser(ctx context.Context, req *dto.UpdateUserReq) error {
	if req.Email != nil && strings.TrimSpace(*req.Email) == "" {
		return errs.NewError(ctx, status.USER_MISSING_EMAIL, nil, ErrEmailRequired)
	}
	// user_name update is optional, but if supplied it must be non-empty
	// — ma_users.user_name is NOT NULL so an empty rewrite would fail.
	if req.Name != nil && strings.TrimSpace(*req.Name) == "" {
		return errs.NewError(ctx, status.USER_MISSING_NAME, nil, ErrUserNameMustBeNonEmptyWhenProvided)
	}
	// role patch is optional. A blank value is treated as "no change"
	// (nil); a non-blank value must be a valid RoleType. Mirrors the
	// profile update validator.
	if req.Role != nil {
		role := strings.TrimSpace(*req.Role)
		if role == "" {
			req.Role = nil
		} else {
			if !enum.RoleType(role).IsSelfAssignable() {
				return errs.NewError(ctx, status.USER_INVALID_ROLE, nil, ErrRoleInvalid)
			}
			req.Role = &role
		}
	}
	// A non-nil Avatar pointer means the client wants to change the
	// avatar reference; an empty string is not a valid reference.
	// Pointer-nil means "no change" and is allowed.
	if req.Avatar != nil && strings.TrimSpace(*req.Avatar) == "" {
		return errs.NewError(ctx, status.USER_AVATAR_INVALID_REFERENCE, nil, ErrAvatarReferenceMustBeNonEmptyWhenProvided)
	}
	if req.Avatar != nil && strings.TrimSpace(*req.Avatar) != "" && req.AvatarFile != nil {
		return errs.NewError(ctx, status.USER_AVATAR_CONFLICT, nil, ErrProvideEitherAvatarFileOrAvatarReference)
	}
	return nil
}

func ValidateDeleteUser(ctx context.Context, req *dto.DeleteUserReq) error {
	if req.UID == 0 {
		return errs.NewError(ctx, status.BAD_REQUEST, nil, ErrUIDMustBeValidId)
	}
	return nil
}
