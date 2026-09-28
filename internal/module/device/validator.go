package device

import (
	"context"

	dto "math-ai.com/math-ai/internal/application/dto/device"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
)

func validateUID(ctx context.Context, uid int64) error {
	if uid == 0 {
		return errs.NewError(ctx, status.DEVICE_MISSING_USER_ID, nil, ErrUIDRequired)
	}
	return nil
}

func validateDeviceUUID(ctx context.Context, deviceUUID string) error {
	if deviceUUID == "" {
		return errs.NewError(ctx, status.DEVICE_NOT_FOUND, nil, ErrDeviceIDRequired)
	}
	return nil
}

func validateDeviceID(ctx context.Context, deviceID int64) error {
	if deviceID == 0 {
		return errs.NewError(ctx, status.DEVICE_NOT_FOUND, nil, ErrDeviceIDRequired)
	}
	return nil
}

func ValidateGetDevice(ctx context.Context, req *dto.GetDeviceByIdReq) error {
	return validateDeviceID(ctx, req.DeviceID)
}

func ValidateListDevices(ctx context.Context, req *dto.ListDevicesReq) error {
	return validateUID(ctx, req.UID)
}

func ValidateUpdateDevice(ctx context.Context, req *dto.UpdateDeviceReq) error {
	if err := validateUID(ctx, req.UID); err != nil {
		return err
	}
	return validateDeviceID(ctx, req.DeviceID)
}

func ValidateRevokeDevice(ctx context.Context, req *dto.RevokeDeviceReq) error {
	if err := validateUID(ctx, req.UID); err != nil {
		return err
	}
	return validateDeviceUUID(ctx, req.DevicUUID)
}

func ValidateDeleteDevice(ctx context.Context, req *dto.DeleteDeviceReq) error {
	if err := validateUID(ctx, req.UID); err != nil {
		return err
	}
	return validateDeviceID(ctx, req.DeviceID)
}

func ValidateVerifyDevice(ctx context.Context, req *dto.VerifyDeviceReq) error {
	if err := validateUID(ctx, req.UID); err != nil {
		return err
	}
	return validateDeviceUUID(ctx, req.DeviceUUID)
}
