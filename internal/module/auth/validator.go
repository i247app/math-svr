package auth

import (
	"context"
	"strings"

	"math-ai.com/math-ai/internal/application/dto"
	"math-ai.com/math-ai/internal/domain/login"
	errs "math-ai.com/math-ai/internal/shared/errs"
)

func ValidateLogin(ctx context.Context, req *dto.LoginReq) error {
	if strings.TrimSpace(req.LoginName) == "" {
		return errs.NewError(ctx, login.AuthMissingLoginName, nil, ErrLoginNameRequired)
	}
	return nil
}

func ValidateLogout(ctx context.Context, req *dto.LogoutReq) error {
	if req.UID == nil || *req.UID == 0 {
		return errs.NewUnauthorizedError(ctx, ErrUIDNotFoundInSession)
	}
	if req.DeviceUUID == "" {
		return errs.NewError(ctx, login.AuthMissingDeviceUUID, nil, ErrDeviceUUIDRequired)
	}
	return nil
}
