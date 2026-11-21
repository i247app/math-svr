package command

import (
	"context"

	"math-ai.com/math-ai/internal/application/transaction"
	"math-ai.com/math-ai/internal/domain/device"
	"math-ai.com/math-ai/internal/shared/errs"
)

// UpdateDeviceCommand patches mutable attributes of an existing device row —
// rename, refresh push token, attach a note. is_verified is intentionally NOT
// in this surface; verification flips only via the 2FA flow or revoke.
type UpdateDeviceCommand struct {
	UID             int64
	DeviceID        int64
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
			return errs.NewError(ctx, device.DeviceRegistrationFail, nil, err)
		}
		if existing == nil {
			return errs.NewError(ctx, device.DeviceNotFound, nil, ErrDeviceNotFound)
		}
		if existing.UID() == nil || *existing.UID() != cmd.UID {
			return errs.NewError(ctx, device.DeviceNotOwned, nil,
				ErrDeviceNotOwnedByUser)
		}

		patch := device.NewDevice()
		patch.SetDeviceId(cmd.DeviceID)
		patch.SetDeviceName(cmd.DeviceName)
		patch.SetDevicePushToken(cmd.DevicePushToken)
		patch.SetNote(cmd.Note)

		if err := repos.Device.Update(ctx, patch); err != nil {
			return errs.NewError(ctx, device.DeviceRegistrationFail, nil, err)
		}

		refreshed, err := repos.Device.FindByDeviceId(ctx, cmd.DeviceID)
		if err != nil {
			return errs.NewError(ctx, device.DeviceRegistrationFail, nil, err)
		}
		updated = refreshed
		return nil
	})
	if err != nil {
		return nil, err
	}
	return updated, nil
}
