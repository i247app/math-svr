package command

import (
	"context"
	"errors"

	"math-ai.com/math-ai/internal/application/transaction"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/enum"
)

// RevokeDeviceCommand un-trusts a device: flips is_verified back to false and
// kills any still-active login_log for the (user, device_uuid) pair. The row
// itself is preserved so subsequent logins still find it (they will then be
// gated by 2FA again).
type RevokeDeviceCommand struct {
	UserID   int64
	DeviceID int64
}

type RevokeDeviceCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewRevokeDeviceCommandHandler(uow transaction.UnitOfWork) *RevokeDeviceCommandHandler {
	return &RevokeDeviceCommandHandler{uow: uow}
}

func (h *RevokeDeviceCommandHandler) Handle(ctx context.Context, cmd RevokeDeviceCommand) error {
	return h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		d, err := repos.Device.FindByDeviceId(ctx, cmd.DeviceID)
		if err != nil {
			return errs.NewError(ctx, status.DEVICE_REVOKE_FAIL, nil, err)
		}
		if d == nil {
			return errs.NewError(ctx, status.DEVICE_NOT_FOUND, nil, errors.New("device not found"))
		}
		if d.UserId() == nil || *d.UserId() != cmd.UserID {
			return errs.NewError(ctx, status.DEVICE_NOT_OWNED, nil,
				errors.New("device does not belong to user"))
		}

		if err := repos.Device.MarkVerified(ctx, cmd.DeviceID, false); err != nil {
			return errs.NewError(ctx, status.DEVICE_REVOKE_FAIL, nil, err)
		}
		if err := repos.Device.MarkStatusByDeviceId(ctx, cmd.DeviceID, enum.DeviceStatusTypeRevoked); err != nil {
			return errs.NewError(ctx, status.DEVICE_REVOKE_FAIL, nil, err)
		}

		// Kill the active session bound to this device, if any. The
		// login_log row is preserved (status flipped to REVOKED) so audit
		// history survives.
		if err := repos.LoginLog.MarkStatusByUserDevice(
			ctx, cmd.UserID, d.DeviceUUID(), enum.LoginLogStatusTypeRevoked,
		); err != nil {
			return errs.NewError(ctx, status.DEVICE_REVOKE_FAIL, nil, err)
		}
		return nil
	})
}
