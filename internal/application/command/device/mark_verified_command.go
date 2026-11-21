package command

import (
	"context"
	"time"

	"math-ai.com/math-ai/internal/application/command/shared/seqgen"
	"math-ai.com/math-ai/internal/application/transaction"
	"math-ai.com/math-ai/internal/domain/device"
	"math-ai.com/math-ai/internal/domain/seq"
	"math-ai.com/math-ai/internal/infrastructure/logger"
	"math-ai.com/math-ai/internal/shared/enum"
	"math-ai.com/math-ai/internal/shared/errs"
)

// MarkDeviceVerifiedCommand promotes a device to the trusted set. The future
// 2FA module is the canonical caller — after the OTP / Authenticator code is
// confirmed, it invokes this command, then the next /auth/login on the same
// device skips the 2FA gate.
//
// Ownership is enforced so a leaked DeviceID for a different user cannot be
// flipped.
type MarkDeviceVerifiedCommand struct {
	UID             int64
	DeviceUUID      string
	DeviceName      string
	Platform        string
	DevicePushToken *string
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
		d, err := repos.Device.FindByUserDevice(ctx, cmd.UID, cmd.DeviceUUID)
		if err != nil {
			return errs.NewError(ctx, device.DeviceVerificationFail, nil, err)
		}
		if d == nil {
			// return errs.NewError(ctx, device.DeviceNotFound, nil, ErrDeviceNotFound)
			d := device.NewDevice()
			deviceID, err := seqgen.Next(ctx, repos.Seq, seq.NameDevice)
			if err != nil {
				return err
			}

			d.SetDeviceId(deviceID)
			d.SetUID(&cmd.UID)
			d.SetDeviceUUID(cmd.DeviceUUID)
			d.SetDeviceName(cmd.DeviceName)
			d.SetPlatform(enum.ParsePlatformType(cmd.Platform).String())
			d.SetDevicePushToken(cmd.DevicePushToken)
			d.SetIsVerified(true)
			d.SetTrustDt(time.Now().UTC())
			active := enum.DeviceStatusTypeActive.String()
			d.SetDeviceStatus(&active)
			d.SetStatus(enum.StatusActive.String())

			_, err = repos.Device.Create(ctx, d)
			if err != nil {
				return errs.NewError(ctx, device.DeviceRegistrationFail, nil, err)
			}
			return nil
		}
		if d.UID() == nil || *d.UID() != cmd.UID {
			return errs.NewError(ctx, device.DeviceNotOwned, nil, ErrDeviceNotOwnedByUser)
		}
		if d.IsVerified() {
			// return errs.NewError(ctx, device.DeviceAlreadyVerified, nil, ErrDeviceAlreadyVerified)
			log.Info("device is already verified")
		}

		if err := repos.Device.MarkVerifiedByUserDevice(ctx, cmd.UID, cmd.DeviceUUID, true); err != nil {
			return errs.NewError(ctx, device.DeviceVerificationFail, nil, err)
		}
		return nil
	})
}
