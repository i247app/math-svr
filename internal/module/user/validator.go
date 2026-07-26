package user

import (
	"context"
	"strings"

	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/enum"

	dto "math-ai.com/math-ai/internal/application/dto/user"
)

func ValidateCreateUser(ctx context.Context, req *dto.CreateUserReq) error {
	if strings.TrimSpace(req.Phone) == "" {
		return errs.NewError(ctx, status.USER_MISSING_PHONE, nil, ErrPhoneRequired)
	}
	if strings.TrimSpace(req.Name) == "" {
		return errs.NewError(ctx, status.USER_MISSING_NAME, nil, ErrNameRequired)
	}
	// role is optional on the wire — the create command defaults an empty
	// value to STUDENT (ma_users.role is NOT NULL). When supplied it must
	// be a valid RoleType. Normalised in place so the command sees the
	// trimmed token.
	role := strings.TrimSpace(req.Role)
	if role == "" {
		req.Role = ""
	} else {
		if !enum.RoleType(role).IsValid() {
			args := map[string]any{
				"roles": enum.ListRoles(),
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
			if !enum.RoleType(role).IsValid() {
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
	if req.UserID == 0 {
		return errs.NewError(ctx, status.BAD_REQUEST, nil, ErrUserIDMustBeValidUUID)
	}
	return nil
}
