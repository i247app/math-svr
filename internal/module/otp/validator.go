package otp

import (
	"context"
	"strings"

	"math-ai.com/math-ai/internal/application/dto"
	domain "math-ai.com/math-ai/internal/domain/otp"
	"math-ai.com/math-ai/internal/shared/enum"
	"math-ai.com/math-ai/internal/shared/errs"
)

func ValidateSendOtp(ctx context.Context, req *dto.SendOtpReq) error {
	if strings.TrimSpace(req.OtpType) == "" {
		return errs.NewError(ctx, domain.OTPMissingType, nil, ErrOTPTypeRequired)
	}
	if !enum.OtpType(req.OtpType).IsValid() {
		return errs.NewError(ctx, domain.OTPInvalidType, nil, ErrOTPTypeInvalid)
	}
	if strings.TrimSpace(req.Identifier) == "" {
		return errs.NewError(ctx, domain.OTPMissingIdentifier, nil, ErrIdentifierRequired)
	}
	if req.TargetDeviceID != nil && req.OtpType != enum.OtpTypeLogin2FA.String() {
		return errs.NewError(ctx, domain.OTPTargetDeviceRequiresLogin2FA, nil, ErrTargetDeviceRequiresLogin2FA)
	}
	return nil
}

func ValidateVerifyOtp(ctx context.Context, req *dto.VerifyOtpReq) error {
	if strings.TrimSpace(req.OtpType) == "" {
		return errs.NewError(ctx, domain.OTPMissingType, nil, ErrOTPTypeRequired)
	}
	if !enum.OtpType(req.OtpType).IsValid() {
		return errs.NewError(ctx, domain.OTPInvalidType, nil, ErrOTPTypeInvalid)
	}
	if strings.TrimSpace(req.Identifier) == "" {
		return errs.NewError(ctx, domain.OTPMissingIdentifier, nil, ErrIdentifierRequired)
	}
	if strings.TrimSpace(req.OtpCode) == "" {
		return errs.NewError(ctx, domain.OTPMissingCode, nil, ErrOTPCodeRequired)
	}
	return nil
}

func ValidateRevokeOtp(ctx context.Context, req *dto.RevokeOtpReq) error {
	if strings.TrimSpace(req.OtpType) == "" {
		return errs.NewError(ctx, domain.OTPMissingType, nil, ErrOTPTypeRequired)
	}
	if !enum.OtpType(req.OtpType).IsValid() {
		return errs.NewError(ctx, domain.OTPInvalidType, nil, ErrOTPTypeInvalid)
	}
	if strings.TrimSpace(req.Identifier) == "" {
		return errs.NewError(ctx, domain.OTPMissingIdentifier, nil, ErrIdentifierRequired)
	}
	return nil
}
