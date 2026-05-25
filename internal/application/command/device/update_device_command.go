package command

import (
	"context"
	"errors"

	"math-ai.com/math-ai/internal/application/transaction"
	"math-ai.com/math-ai/internal/domain/device"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
)

// UpdateDeviceCommand patches mutable attributes of an existing device row —
// rename, refresh push token, attach a note. is_verified is intentionally NOT
// in this surface; verification flips only via the 2FA flow or revoke.
type UpdateDeviceCommand struct {
	UserID          string
	DeviceID        string
	DeviceName      string
	DevicePushToken *string
	Note            *string
}

type UpdateDeviceCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewUpdateDeviceCommandHandler(uow transaction.UnitOfWork) *UpdateDeviceCommandHandler {
	return &UpdateDeviceCommandHandler{uow: uow}
}

func (h *UpdateDeviceCommandHandler) Handle(ctx context.Context, cmd UpdateDeviceCommand) (*device.Device, error) {
	var updated *device.Device

	err := h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		existing, err := repos.Device.FindByDeviceId(ctx, cmd.DeviceID)
		if err != nil {
			return errs.NewError(ctx, status.DEVICE_REGISTRATION_FAIL, nil, err)
		}
		if existing == nil {
			return errs.NewError(ctx, status.DEVICE_NOT_FOUND, nil, errors.New("device not found"))
		}
		if existing.UserId() == nil || *existing.UserId() != cmd.UserID {
			return errs.NewError(ctx, status.DEVICE_NOT_OWNED, nil,
				errors.New("device does not belong to user"))
		}

		patch := device.NewDevice()
		patch.SetDeviceId(cmd.DeviceID)
		patch.SetDeviceName(cmd.DeviceName)
		patch.SetDevicePushToken(cmd.DevicePushToken)
		patch.SetNote(cmd.Note)

		if err := repos.Device.Update(ctx, patch); err != nil {
			return errs.NewError(ctx, status.DEVICE_REGISTRATION_FAIL, nil, err)
		}

		refreshed, err := repos.Device.FindByDeviceId(ctx, cmd.DeviceID)
		if err != nil {
			return errs.NewError(ctx, status.DEVICE_REGISTRATION_FAIL, nil, err)
		}
		updated = refreshed
		return nil
	})
	if err != nil {
		return nil, err
	}
	return updated, nil
}
