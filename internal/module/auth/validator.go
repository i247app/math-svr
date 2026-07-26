package auth

import (
	"context"
	"strings"

	dto "math-ai.com/math-ai/internal/application/dto/auth"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
)

func ValidateLogin(ctx context.Context, req *dto.LoginReq) error {
	if strings.TrimSpace(req.LoginName) == "" {
		return errs.NewError(ctx, status.AUTH_MISSING_LOGIN_NAME, nil, ErrLoginNameRequired)
	}
	return nil
}

func ValidateLogout(ctx context.Context, req *dto.LogoutReq) error {
	return nil
}
