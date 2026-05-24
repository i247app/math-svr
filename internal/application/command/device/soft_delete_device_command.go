package command

import (
	"context"
	"errors"

	"math-ai.com/math-ai/internal/application/transaction"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/enum"
)

type SoftDeleteDeviceCommand struct {
	UserID   string
	DeviceID string
}

type SoftDeleteDeviceCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewSoftDeleteDeviceCommandHandler(uow transaction.UnitOfWork) *SoftDeleteDeviceCommandHandler {
	return &SoftDeleteDeviceCommandHandler{uow: uow}
}

func (h *SoftDeleteDeviceCommandHandler) Handle(ctx context.Context, cmd SoftDeleteDeviceCommand) error {
	return h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		d, err := repos.Device.FindByDeviceId(ctx, cmd.DeviceID)
		if err != nil {
			return errs.NewError(ctx, status.DEVICE_REGISTRATION_FAIL, nil, err)
		}
		if d == nil {
			return errs.NewError(ctx, status.DEVICE_NOT_FOUND, nil, errors.New("device not found"))
		}
		if d.UserId() == nil || d.UserId().String() != cmd.UserID {
			return errs.NewError(ctx, status.DEVICE_NOT_OWNED, nil,
				errors.New("device does not belong to user"))
		}

		if err := repos.Device.SoftDeleteByDeviceId(ctx, cmd.DeviceID); err != nil {
			return errs.NewError(ctx, status.DEVICE_REGISTRATION_FAIL, nil, err)
		}

		// Cascade: kill any active session bound to this device.
		if err := repos.LoginLog.MarkStatusByUserDevice(
			ctx, cmd.UserID, d.DeviceUUID(), enum.LoginLogStatusTypeRevoked,
		); err != nil {
			return errs.NewError(ctx, status.DEVICE_REGISTRATION_FAIL, nil, err)
		}
		return nil
	})
}
