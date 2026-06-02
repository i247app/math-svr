package command

import (
	"context"

	"math-ai.com/math-ai/internal/application/transaction"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
)

// MarkDeviceVerifiedCommand promotes a device to the trusted set. The future
// 2FA module is the canonical caller — after the OTP / Authenticator code is
// confirmed, it invokes this command, then the next /auth/login on the same
// device skips the 2FA gate.
//
// Ownership is enforced so a leaked DeviceID for a different user cannot be
// flipped.
type MarkDeviceVerifiedCommand struct {
	UserID   int64
	DeviceID int64
}

type MarkDeviceVerifiedCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewMarkDeviceVerifiedCommandHandler(uow transaction.UnitOfWork) *MarkDeviceVerifiedCommandHandler {
	return &MarkDeviceVerifiedCommandHandler{uow: uow}
}

func (h *MarkDeviceVerifiedCommandHandler) Handle(ctx context.Context, cmd MarkDeviceVerifiedCommand) error {
	return h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		d, err := repos.Device.FindByDeviceId(ctx, cmd.DeviceID)
		if err != nil {
			return errs.NewError(ctx, status.DEVICE_VERIFICATION_FAIL, nil, err)
		}
		if d == nil {
			return errs.NewError(ctx, status.DEVICE_NOT_FOUND, nil, ErrDeviceNotFound)
		}
		if d.UserId() == nil || *d.UserId() != cmd.UserID {
			return errs.NewError(ctx, status.DEVICE_NOT_OWNED, nil,
				ErrDeviceNotOwnedByUser)
		}
		if d.IsVerified() {
			return errs.NewError(ctx, status.DEVICE_ALREADY_VERIFIED, nil,
				ErrDeviceAlreadyVerified)
		}

		if err := repos.Device.MarkVerified(ctx, cmd.DeviceID, true); err != nil {
			return errs.NewError(ctx, status.DEVICE_VERIFICATION_FAIL, nil, err)
		}
		return nil
	})
}
