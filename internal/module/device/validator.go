package device

import (
	"context"
	"errors"

	"github.com/google/uuid"

	dto "math-ai.com/math-ai/internal/application/dto/device"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
)

func validateUserID(ctx context.Context, userID uuid.UUID) error {
	if userID == uuid.Nil {
		return errs.NewError(ctx, status.DEVICE_MISSING_USER_ID, nil, errors.New("user_id is required"))
	}
	return nil
}

func validateDeviceID(ctx context.Context, deviceID uuid.UUID) error {
	if deviceID == uuid.Nil {
		return errs.NewError(ctx, status.DEVICE_NOT_FOUND, nil, errors.New("device_id is required"))
	}
	return nil
}

func ValidateGetDevice(ctx context.Context, req *dto.GetDeviceByIdReq) error {
	return validateDeviceID(ctx, req.DeviceID)
}

func ValidateListDevices(ctx context.Context, req *dto.ListDevicesReq) error {
	return validateUserID(ctx, req.UserID)
}

func ValidateUpdateDevice(ctx context.Context, req *dto.UpdateDeviceReq) error {
	if err := validateUserID(ctx, req.UserID); err != nil {
		return err
	}
	return validateDeviceID(ctx, req.DeviceID)
}

func ValidateRevokeDevice(ctx context.Context, req *dto.RevokeDeviceReq) error {
	if err := validateUserID(ctx, req.UserID); err != nil {
		return err
	}
	return validateDeviceID(ctx, req.DeviceID)
}

func ValidateDeleteDevice(ctx context.Context, req *dto.DeleteDeviceReq) error {
	if err := validateUserID(ctx, req.UserID); err != nil {
		return err
	}
	return validateDeviceID(ctx, req.DeviceID)
}

func ValidateVerifyDevice(ctx context.Context, req *dto.VerifyDeviceReq) error {
	if err := validateUserID(ctx, req.UserID); err != nil {
		return err
	}
	return validateDeviceID(ctx, req.DeviceID)
}
