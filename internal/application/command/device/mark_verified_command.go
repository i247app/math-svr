package command

import (
	"context"

	"math-ai.com/math-ai/internal/application/transaction"
	"math-ai.com/math-ai/internal/domain/device"
	"math-ai.com/math-ai/internal/domain/seq"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/mtime"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/infrastructure/logger"
	"math-ai.com/math-ai/internal/shared/enum"
)

// MarkDeviceVerifiedCommand promotes a device to the trusted set. The future
// 2FA module is the canonical caller — after the OTP / Authenticator code is
// confirmed, it invokes this command, then the next /auth/login on the same
// device skips the 2FA gate.
//
// Ownership is enforced so a leaked DeviceID for a different user cannot be
// flipped.
type MarkDeviceVerifiedCommand struct {
	UserID     int64
	DeviceUUID string
	DeviceName string
}

type MarkDeviceVerifiedCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewMarkDeviceVerifiedCommandHandler(uow transaction.UnitOfWork) *MarkDeviceVerifiedCommandHandler {
	return &MarkDeviceVerifiedCommandHandler{uow: uow}
}

func (h *MarkDeviceVerifiedCommandHandler) Handle(ctx context.Context, cmd MarkDeviceVerifiedCommand) error {
	log := logger.From(ctx)
	return h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		d, err := repos.Device.FindByUserDevice(ctx, cmd.UserID, cmd.DeviceUUID)
		if err != nil {
			return errs.NewError(ctx, status.DEVICE_VERIFICATION_FAIL, nil, err)
		}
		if d == nil {
			// return errs.NewError(ctx, status.DEVICE_NOT_FOUND, nil, ErrDeviceNotFound)
			d := device.NewDevice()
			deviceID, err := nextSeqID(ctx, repos, seq.NameDevice)
			if err != nil {
				return err
			}

			d.SetDeviceId(deviceID)
			d.SetUserId(&cmd.UserID)
			d.SetDeviceUUID(cmd.DeviceUUID)
			d.SetDeviceName(cmd.DeviceName)
			d.SetIsVerified(true)
			d.SetTrustDt(mtime.Now())
			active := enum.DeviceStatusTypeActive.String()
			d.SetDeviceStatus(&active)
			d.SetStatus(enum.StatusActive.String())

			_, err = repos.Device.Create(ctx, d)
			if err != nil {
				return errs.NewError(ctx, status.DEVICE_REGISTRATION_FAIL, nil, err)
			}
			return nil
		}
		if d.UserId() == nil || *d.UserId() != cmd.UserID {
			return errs.NewError(ctx, status.DEVICE_NOT_OWNED, nil, ErrDeviceNotOwnedByUser)
		}
		if d.IsVerified() {
			// return errs.NewError(ctx, status.DEVICE_ALREADY_VERIFIED, nil, ErrDeviceAlreadyVerified)
			log.Info("device is already verified")
		}

		if err := repos.Device.MarkVerifiedByUserDevice(ctx, cmd.UserID, cmd.DeviceUUID, true); err != nil {
			return errs.NewError(ctx, status.DEVICE_VERIFICATION_FAIL, nil, err)
		}
		return nil
	})
}
