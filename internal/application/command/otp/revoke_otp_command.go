package command

import (
	"context"

	"math-ai.com/math-ai/internal/application/transaction"
	"math-ai.com/math-ai/internal/domain/otp"
	"math-ai.com/math-ai/internal/shared/enum"
	"math-ai.com/math-ai/internal/shared/errs"
)

// RevokeOtpCommand revokes every PENDING OTP for (type, identifier). The
// canonical caller is logout: when a session is killed, any in-flight 2FA
// challenge should die with it so a thief can't complete the challenge after
// the user signs out.
type RevokeOtpCommand struct {
	OtpType    enum.OtpType
	Identifier string
}

type RevokeOtpCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewRevokeOtpCommandHandler(uow transaction.UnitOfWork) *RevokeOtpCommandHandler {
	return &RevokeOtpCommandHandler{uow: uow}
}

func (h *RevokeOtpCommandHandler) Handle(ctx context.Context, cmd RevokeOtpCommand) error {
	if !cmd.OtpType.IsValid() {
		return errs.NewError(ctx, otp.OTPInvalidType, nil, ErrInvalidOtpType)
	}
	if cmd.Identifier == "" {
		return errs.NewError(ctx, otp.OTPMissingIdentifier, nil, ErrIdentifierRequired)
	}

	return h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		if err := repos.Otp.RevokePendingByTypeIdentifier(ctx, cmd.OtpType, cmd.Identifier); err != nil {
			return errs.NewError(ctx, otp.OTPGenerationFailed, nil, err)
		}
		return nil
	})
}
