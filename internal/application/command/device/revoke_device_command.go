package command

import (
	"context"

	"math-ai.com/math-ai/internal/application/transaction"
	devicedom "math-ai.com/math-ai/internal/domain/device"
	"math-ai.com/math-ai/internal/shared/enum"
	errs "math-ai.com/math-ai/internal/shared/errs"
)

// RevokeDeviceCommand un-trusts a device: flips is_verified back to false and
// kills any still-active login_log for the (user, device_uuid) pair. The row
// itself is preserved so subsequent logins still find it (they will then be
// gated by 2FA again).
type RevokeDeviceCommand struct {
	UID        int64
	DeviceUUID string
}

type RevokeDeviceCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewRevokeDeviceCommandHandler(uow transaction.UnitOfWork) *RevokeDeviceCommandHandler {
	return &RevokeDeviceCommandHandler{uow: uow}
}

func (h *RevokeDeviceCommandHandler) Handle(ctx context.Context, cmd RevokeDeviceCommand) error {
	return h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		d, err := repos.Device.FindByUserDevice(ctx, cmd.UID, cmd.DeviceUUID)
		if err != nil {
			return errs.NewError(ctx, devicedom.DeviceVerificationFail, nil, err)
		}
		if d == nil {
			return errs.NewError(ctx, devicedom.DeviceNotFound, nil, ErrDeviceNotFound)
		}
		if d.UID() == nil || *d.UID() != cmd.UID {
			return errs.NewError(ctx, devicedom.DeviceNotOwned, nil,
				ErrDeviceNotOwnedByUser)
		}

		if err := repos.Device.MarkVerifiedByUserDevice(ctx, cmd.UID, cmd.DeviceUUID, false); err != nil {
			return errs.NewError(ctx, devicedom.DeviceRevokeFail, nil, err)
		}
		if err := repos.Device.MarkStatusByDeviceId(ctx, d.DeviceId(), enum.DeviceStatusTypeRevoked); err != nil {
			return errs.NewError(ctx, devicedom.DeviceRevokeFail, nil, err)
		}

		// Kill the active session bound to this device, if any. The
		// login_log row is preserved (status flipped to REVOKED) so audit
		// history survives.
		if err := repos.LoginLog.MarkStatusByUserDevice(
			ctx, cmd.UID, d.DeviceUUID(), enum.LoginLogStatusTypeRevoked,
		); err != nil {
			return errs.NewError(ctx, devicedom.DeviceRevokeFail, nil, err)
		}
		return nil
	})
}
