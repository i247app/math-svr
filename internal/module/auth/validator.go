package auth

import (
	"context"
	"errors"
	"strings"

	dto "math-ai.com/math-ai/internal/application/dto/auth"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
)

func ValidateLogin(ctx context.Context, req *dto.LoginReq) error {
	if strings.TrimSpace(req.Phone) == "" {
		return errs.NewError(ctx, status.AUTH_MISSING_PHONE, nil, errors.New("phone is required"))
	}
	return nil
}

func ValidateLogout(ctx context.Context, req *dto.LogoutReq) error {
	return nil
}
