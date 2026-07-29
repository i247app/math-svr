package otp

import (
	"context"
	"strings"

	dto "math-ai.com/math-ai/internal/application/dto/otp"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/enum"
)

func ValidateSendOtp(ctx context.Context, req *dto.SendOtpReq) error {
	if strings.TrimSpace(req.OtpType) == "" {
		return errs.NewError(ctx, status.OTP_MISSING_TYPE, nil, ErrOTPTypeRequired)
	}
	if !enum.OtpType(req.OtpType).IsValid() {
		return errs.NewError(ctx, status.OTP_INVALID_TYPE, nil, ErrOTPTypeInvalid)
	}
	if strings.TrimSpace(req.Identifier) == "" {
		return errs.NewError(ctx, status.OTP_MISSING_IDENTIFIER, nil, ErrIdentifierRequired)
	}
	if req.TargetDeviceID != nil && req.OtpType != string(enum.OtpTypeLogin2FA) {
		return errs.NewError(ctx, status.OTP_TARGET_DEVICE_REQUIRES_LOGIN2FA, nil, ErrTargetDeviceRequiresLogin2FA)
	}
	return nil
}

func ValidateVerifyOtp(ctx context.Context, req *dto.VerifyOtpReq) error {
	if strings.TrimSpace(req.OtpType) == "" {
		return errs.NewError(ctx, status.OTP_MISSING_TYPE, nil, ErrOTPTypeRequired)
	}
	if !enum.OtpType(req.OtpType).IsValid() {
		return errs.NewError(ctx, status.OTP_INVALID_TYPE, nil, ErrOTPTypeInvalid)
	}
	if strings.TrimSpace(req.Identifier) == "" {
		return errs.NewError(ctx, status.OTP_MISSING_IDENTIFIER, nil, ErrIdentifierRequired)
	}
	if strings.TrimSpace(req.OtpCode) == "" {
		return errs.NewError(ctx, status.OTP_MISSING_CODE, nil, ErrOTPCodeRequired)
	}
	return nil
}

func ValidateRevokeOtp(ctx context.Context, req *dto.RevokeOtpReq) error {
	if strings.TrimSpace(req.OtpType) == "" {
		return errs.NewError(ctx, status.OTP_MISSING_TYPE, nil, ErrOTPTypeRequired)
	}
	if !enum.OtpType(req.OtpType).IsValid() {
		return errs.NewError(ctx, status.OTP_INVALID_TYPE, nil, ErrOTPTypeInvalid)
	}
	if strings.TrimSpace(req.Identifier) == "" {
		return errs.NewError(ctx, status.OTP_MISSING_IDENTIFIER, nil, ErrIdentifierRequired)
	}
	return nil
}
