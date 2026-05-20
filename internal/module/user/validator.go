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
		return errs.NewError(ctx, status.PROFILE_MISSING_NAME, nil, errors.New("name is required"))
	}
	return nil
}

func ValidateUpdateUser(ctx context.Context, req *dto.UpdateUserReq) error {
	if req.Email != nil && strings.TrimSpace(*req.Email) == "" {
		return errs.NewError(ctx, status.USER_MISSING_EMAIL, nil, errors.New("email is required"))
	}

	return nil
}

func ValidateDeleteUser(ctx context.Context, req *dto.DeleteUserReq) error {
	if req.UserID.String() == "" {
		return errs.NewError(ctx, status.BAD_REQUEST, nil, errors.New("user id must be a valid uuid"))
	}
	return nil
}
