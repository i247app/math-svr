package command

import (
	"context"
	"errors"

	"math-ai.com/math-ai/internal/application/transaction"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/enum"
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
		return errs.NewError(ctx, status.OTP_INVALID_TYPE, nil, errors.New("invalid otp type"))
	}
	if cmd.Identifier == "" {
		return errs.NewError(ctx, status.OTP_MISSING_IDENTIFIER, nil, errors.New("identifier is required"))
	}

	return h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		if err := repos.Otp.RevokePendingByTypeIdentifier(ctx, cmd.OtpType, cmd.Identifier); err != nil {
			return errs.NewError(ctx, status.OTP_GENERATION_FAILED, nil, err)
		}
		return nil
	})
}
