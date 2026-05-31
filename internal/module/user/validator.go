package user

import (
	"context"
	"errors"
	"strings"

	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"

	dto "math-ai.com/math-ai/internal/application/dto/user"
)

func ValidateCreateUser(ctx context.Context, req *dto.CreateUserReq) error {
	if strings.TrimSpace(req.Phone) == "" {
		return errs.NewError(ctx, status.USER_MISSING_PHONE, nil, errors.New("phone is required"))
	}
	if strings.TrimSpace(req.Name) == "" {
		return errs.NewError(ctx, status.USER_MISSING_NAME, nil, errors.New("name is required"))
	}
	// Avatar can come as a file upload OR a string reference, never both
	// — the two sources collide semantically and would force the service
	// to pick a winner silently. Format/host validity lives in the
	// service layer (normalizeAvatarKey).
	if strings.TrimSpace(req.Avatar) != "" && req.AvatarFile != nil {
		return errs.NewError(ctx, status.USER_AVATAR_CONFLICT, nil,
			errors.New("provide either avatar file or avatar reference"))
	}
	return nil
}

func ValidateUpdateUser(ctx context.Context, req *dto.UpdateUserReq) error {
	if req.Email != nil && strings.TrimSpace(*req.Email) == "" {
		return errs.NewError(ctx, status.USER_MISSING_EMAIL, nil, errors.New("email is required"))
	}
	// user_name update is optional, but if supplied it must be non-empty
	// — ma_users.user_name is NOT NULL so an empty rewrite would fail.
	if req.Name != nil && strings.TrimSpace(*req.Name) == "" {
		return errs.NewError(ctx, status.USER_MISSING_NAME, nil, errors.New("user_name must be non-empty when provided"))
	}
	// A non-nil Avatar pointer means the client wants to change the
	// avatar reference; an empty string is not a valid reference.
	// Pointer-nil means "no change" and is allowed.
	if req.Avatar != nil && strings.TrimSpace(*req.Avatar) == "" {
		return errs.NewError(ctx, status.USER_AVATAR_INVALID_REFERENCE, nil,
			errors.New("avatar reference must be non-empty when provided"))
	}
	if req.Avatar != nil && strings.TrimSpace(*req.Avatar) != "" && req.AvatarFile != nil {
		return errs.NewError(ctx, status.USER_AVATAR_CONFLICT, nil,
			errors.New("provide either avatar file or avatar reference"))
	}
	return nil
}

func ValidateDeleteUser(ctx context.Context, req *dto.DeleteUserReq) error {
	if req.UserID == 0 {
		return errs.NewError(ctx, status.BAD_REQUEST, nil, errors.New("user id must be a valid uuid"))
	}
	return nil
}
