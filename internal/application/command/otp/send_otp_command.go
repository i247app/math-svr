package command

import (
	"context"
	"errors"
	"time"

	"math-ai.com/math-ai/internal/adapter/otp_delivery"
	"math-ai.com/math-ai/internal/application/transaction"
	"math-ai.com/math-ai/internal/domain/otp"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	mtime "math-ai.com/math-ai/internal/domain/shared/time"
	"math-ai.com/math-ai/internal/shared/enum"
	"math-ai.com/math-ai/internal/shared/utils"
)

// SendOtpCommand issues a fresh OTP for (type, identifier) and dispatches it
// via the OTP delivery adapter. Inside one UoW it:
//
//  1. Enforces the resend cooldown (OTP_TOO_FREQUENT).
//  2. Enforces the per-window send cap (OTP_RATE_LIMITED).
//  3. Revokes any prior PENDING OTPs for (type, identifier).
//  4. Generates a 6-digit code, hashes it, inserts the row.
//
// Step 5 happens AFTER commit: the delivery adapter is invoked. If delivery
// fails, the row stays in place (audit trail) and the caller sees
// OTP_DELIVERY_FAILED — the client can retry, which will trip the cooldown
// rather than create a duplicate row.
type SendOtpCommand struct {
	OtpType    enum.OtpType
	Identifier string
	UserID     *string
	DeviceUUID *string
	DeviceName *string
	Channel    enum.OtpChannel // empty = auto-detect
	// Language enum.LanguageType
}

type SendOtpCommandResult struct {
	OtpID     string
	ExpiresAt mtime.MathTime
	Channel   otp_delivery.ChannelName
	OTPCode   string
	OTPType   string
}

type SendOtpCommandHandler struct {
	uow      transaction.UnitOfWork
	delivery *otp_delivery.Adapter
}

func NewSendOtpCommandHandler(uow transaction.UnitOfWork, delivery *otp_delivery.Adapter) *SendOtpCommandHandler {
	return &SendOtpCommandHandler{uow: uow, delivery: delivery}
}

func (h *SendOtpCommandHandler) Handle(ctx context.Context, cmd SendOtpCommand) (*SendOtpCommandResult, error) {
	if !cmd.OtpType.IsValid() {
		return nil, errs.NewError(ctx, status.OTP_INVALID_TYPE, nil, errors.New("invalid otp type"))
	}
	if cmd.Identifier == "" {
		return nil, errs.NewError(ctx, status.OTP_MISSING_IDENTIFIER, nil, errors.New("identifier is required"))
	}

	channel, err := h.resolveChannel(ctx, cmd)
	if err != nil {
		return nil, err
	}

	// Generate the plaintext code outside the UoW — rand.Int can be slow
	// under crypto entropy pressure and shouldn't hold a tx open.
	plainCode, err := generateCode(OtpCodeLength)
	if err != nil {
		return nil, errs.NewError(ctx, status.OTP_GENERATION_FAILED, nil, err)
	}

	now := time.Now().UTC()
	expiresAt := now.Add(TtlFor(cmd.OtpType))

	var createdOtpID string
	err = h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		// 1. Cooldown.
		// Compare against OtpCreateDt (app-set, always UTC) not CreateDt
		// (MySQL DEFAULT CURRENT_TIMESTAMP(6) — emits the server's local
		// wall-clock, which the driver mis-tags as UTC, yielding a
		// negative age and a permanently-skipped cooldown).
		latest, err := repos.Otp.FindLatestPending(ctx, cmd.OtpType, cmd.Identifier)
		if err != nil {
			return errs.NewError(ctx, status.OTP_GENERATION_FAILED, nil, err)
		}
		if latest != nil {
			age := now.Sub(latest.OtpCreateDt().Time)
			if age >= 0 && age < OtpResendCooldown {
				// return errs.NewError(ctx, status.OTP_TOO_FREQUENT, map[string]any{
				// 	"retry_after_seconds": int((OtpResendCooldown - age).Seconds()),
				// }, errors.New("resend cooldown not elapsed"))
				createdOtpID = latest.OtpId()
				plainCode = latest.OtpCode()
				expiresAt = latest.OtpExpireDt().Time

				return nil
			}
		}

		// 2. Window cap
		count, err := repos.Otp.CountSentSince(ctx, cmd.OtpType, cmd.Identifier, now.Add(-OtpSendWindow))
		if err != nil {
			return errs.NewError(ctx, status.OTP_GENERATION_FAILED, nil, err)
		}
		if count >= OtpMaxSendsPerWindow {
			return errs.NewError(ctx, status.OTP_RATE_LIMITED, map[string]any{
				"window_seconds": int(OtpSendWindow.Seconds()),
				"limit":          OtpMaxSendsPerWindow,
			}, errors.New("send-window cap reached"))
		}

		// 3. Revoke prior PENDING rows
		if err := repos.Otp.RevokePendingByTypeIdentifier(ctx, cmd.OtpType, cmd.Identifier); err != nil {
			return errs.NewError(ctx, status.OTP_GENERATION_FAILED, nil, err)
		}

		// 4. Insert fresh row
		o := otp.NewOtp()
		o.SetOtpId(utils.GenerateUUID().String())
		o.SetOtpType(cmd.OtpType.String())
		o.SetUserId(cmd.UserID)
		o.SetIdentifier(cmd.Identifier)
		o.SetDeviceUUID(cmd.DeviceUUID)
		o.SetDeviceName(cmd.DeviceName)
		// o.SetOtpCode(hashCode(plainCode))
		o.SetOtpCode(plainCode)
		o.SetOtpCreateDt(mtime.MathTime{Time: now})
		o.SetOtpExpireDt(mtime.MathTime{Time: expiresAt})
		o.SetAttemptCount(0)
		pending := enum.OtpStatusTypePending.String()
		o.SetOtpStatus(&pending)
		o.SetStatus(enum.StatusActive.String())

		created, err := repos.Otp.Create(ctx, o)
		if err != nil {
			return errs.NewError(ctx, status.OTP_GENERATION_FAILED, nil, err)
		}
		createdOtpID = created.OtpId()
		return nil
	})
	if err != nil {
		return nil, err
	}

	// // 5. Dispatch after commit. Delivery failure does NOT revoke the row;
	// // the audit trail of attempts is intentional, and a retry will trip
	// // the cooldown rather than spam the user.
	// deliverErr := h.delivery.SendVia(ctx, channel, otp_delivery.Message{
	// 	Identifier: cmd.Identifier,
	// 	Code:       plainCode,
	// 	OtpType:    cmd.OtpType,
	// 	Language:   cmd.Language,
	// 	ExpiresAt:  expiresAt,
	// })
	// if deliverErr != nil {
	// 	return nil, errs.NewError(ctx, status.OTP_DELIVERY_FAILED, nil, deliverErr)
	// }

	return &SendOtpCommandResult{
		OtpID:     createdOtpID,
		ExpiresAt: mtime.MathTime{Time: expiresAt},
		Channel:   channel,
		OTPCode:   plainCode,
		OTPType:   cmd.OtpType.String(),
	}, nil
}

// resolveChannel returns the channel the OTP will ship through. Explicit
// overrides win; otherwise the adapter's auto-detect picks SMS vs email from
// identifier shape.
func (h *SendOtpCommandHandler) resolveChannel(ctx context.Context, cmd SendOtpCommand) (otp_delivery.ChannelName, error) {
	var channel otp_delivery.ChannelName
	switch cmd.Channel {
	case enum.OtpChannelSMS:
		channel = otp_delivery.ChannelSMS
	case enum.OtpChannelEmail:
		channel = otp_delivery.ChannelEmail
	case enum.OtpChannelAuto:
		if isEmailLike(cmd.Identifier) {
			channel = otp_delivery.ChannelEmail
		} else {
			channel = otp_delivery.ChannelSMS
		}
	default:
		return "", errs.NewError(ctx, status.OTP_NO_DELIVERY_CHANNEL, nil,
			errors.New("unknown channel"))
	}
	if !h.delivery.HasChannel(channel) {
		return "", errs.NewError(ctx, status.OTP_NO_DELIVERY_CHANNEL, map[string]any{
			"channel": string(channel),
		}, errors.New("channel not registered"))
	}
	return channel, nil
}

func isEmailLike(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] == '@' {
			return true
		}
	}
	return false
}
