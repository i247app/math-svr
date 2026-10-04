package command

import (
	"context"

	"math-ai.com/math-ai/internal/application/transaction"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/enum"
)

// ForceDeleteDeviceCommand physically removes the device row. Same ownership
// check and session cascade as SoftDeleteDeviceCommand.
type ForceDeleteDeviceCommand struct {
	UID      int64
	DeviceID int64
}

type ForceDeleteDeviceCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewForceDeleteDeviceCommandHandler(uow transaction.UnitOfWork) *ForceDeleteDeviceCommandHandler {
	return &ForceDeleteDeviceCommandHandler{uow: uow}
}

func (h *ForceDeleteDeviceCommandHandler) Handle(ctx context.Context, cmd ForceDeleteDeviceCommand) error {
	return h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		d, err := repos.Device.FindByDeviceId(ctx, cmd.DeviceID)
		if err != nil {
			return errs.NewError(ctx, status.DEVICE_REGISTRATION_FAIL, nil, err)
		}
		if d == nil {
			return errs.NewError(ctx, status.DEVICE_NOT_FOUND, nil, ErrDeviceNotFound)
		}
		if d.Uid() == nil || *d.Uid() != cmd.UID {
			return errs.NewError(ctx, status.DEVICE_NOT_OWNED, nil,
				ErrDeviceNotOwnedByUser)
		}

		if err := repos.Device.ForceDeleteByDeviceId(ctx, cmd.DeviceID); err != nil {
			return errs.NewError(ctx, status.DEVICE_REGISTRATION_FAIL, nil, err)
		}

		// Cascade: kill any active session bound to this device.
		if err := repos.LoginLog.MarkStatusByUserDevice(
			ctx, cmd.UID, d.DeviceUUID(), enum.LoginLogStatusTypeRevoked,
		); err != nil {
			return errs.NewError(ctx, status.DEVICE_REGISTRATION_FAIL, nil, err)
		}
		return nil
	})
}
