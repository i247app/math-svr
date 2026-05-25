package command

import (
	"context"
	"errors"
	"time"

	"math-ai.com/math-ai/internal/application/transaction"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/enum"
)

// VerifyOtpCommand consumes a PENDING OTP. On success it returns the
// metadata the caller needs to drive downstream effects (e.g. the LOGIN2FA
// caller flips device.is_verified using DeviceUUID). The OTP module itself
// never invokes downstream services — keeps it reusable.
type VerifyOtpCommand struct {
	OtpType    enum.OtpType
	Identifier string
	Code       string
}

type VerifyOtpCommandResult struct {
	OtpID      string
	UserID     *string
	DeviceUUID *string
}

type VerifyOtpCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewVerifyOtpCommandHandler(uow transaction.UnitOfWork) *VerifyOtpCommandHandler {
	return &VerifyOtpCommandHandler{uow: uow}
}

func (h *VerifyOtpCommandHandler) Handle(ctx context.Context, cmd VerifyOtpCommand) (*VerifyOtpCommandResult, error) {
	if !cmd.OtpType.IsValid() {
		return nil, errs.NewError(ctx, status.OTP_INVALID_TYPE, nil, errors.New("invalid otp type"))
	}
	if cmd.Identifier == "" {
		return nil, errs.NewError(ctx, status.OTP_MISSING_IDENTIFIER, nil, errors.New("identifier is required"))
	}
	if cmd.Code == "" {
		return nil, errs.NewError(ctx, status.OTP_MISSING_CODE, nil, errors.New("otp code is required"))
	}

	var result *VerifyOtpCommandResult

	err := h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		o, err := repos.Otp.FindLatestPending(ctx, cmd.OtpType, cmd.Identifier)
		if err != nil {
			return errs.NewError(ctx, status.OTP_GENERATION_FAILED, nil, err)
		}
		if o == nil {
			// No PENDING row — either never sent, or already consumed
			// /revoked /expired. Treat all of them as "not found" to
			// keep response shape uniform and avoid leaking lifecycle.
			return errs.NewError(ctx, status.OTP_NOT_FOUND, nil, errors.New("no pending otp"))
		}

		// Expiry check. Mark EXPIRED so the row is greppable in audit.
		if o.OtpExpireDt().IsValid() && time.Now().UTC().After(o.OtpExpireDt().Time) {
			_ = repos.Otp.MarkStatusByOtpId(ctx, o.OtpId(), enum.OtpStatusTypeExpired)
			return errs.NewError(ctx, status.OTP_EXPIRED, nil, errors.New("otp expired"))
		}

		// Increment attempts BEFORE comparing — otherwise an attacker can
		// brute-force by flooding mismatches.
		newCount, err := repos.Otp.IncrementAttemptCount(ctx, o.OtpId())
		if err != nil {
			return errs.NewError(ctx, status.OTP_GENERATION_FAILED, nil, err)
		}
		if newCount > OtpMaxAttempts {
			_ = repos.Otp.MarkStatusByOtpId(ctx, o.OtpId(), enum.OtpStatusTypeRevoked)
			return errs.NewError(ctx, status.OTP_TOO_MANY_ATTEMPTS, nil,
				errors.New("attempt cap exceeded"))
		}

		if !codesMatch(cmd.Code, o.OtpCode()) {
			// Wrong code, but row is still PENDING — let the user retry
			// up to OtpMaxAttempts.
			if newCount >= OtpMaxAttempts {
				// Last allowed attempt just failed; revoke now so the
				// next /verify call returns OTP_TOO_MANY_ATTEMPTS.
				_ = repos.Otp.MarkStatusByOtpId(ctx, o.OtpId(), enum.OtpStatusTypeRevoked)
			}
			return errs.NewError(ctx, status.OTP_INVALID_CODE, map[string]any{
				"attempts_remaining": maxInt(0, OtpMaxAttempts-newCount),
			}, errors.New("code mismatch"))
		}

		if err := repos.Otp.MarkStatusByOtpId(ctx, o.OtpId(), enum.OtpStatusTypeVerified); err != nil {
			return errs.NewError(ctx, status.OTP_GENERATION_FAILED, nil, err)
		}

		result = &VerifyOtpCommandResult{
			OtpID:      o.OtpId(),
			UserID:     o.UserId(),
			DeviceUUID: o.DeviceUUID(),
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
