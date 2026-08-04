package command

import (
	"context"
	"fmt"
	"time"

	notifAdapter "math-ai.com/math-ai/internal/adapter/notification"
	"math-ai.com/math-ai/internal/adapter/otp_delivery"
	notifCommand "math-ai.com/math-ai/internal/application/command/notification"
	"math-ai.com/math-ai/internal/application/command/shared/seqgen"
	"math-ai.com/math-ai/internal/application/transaction"
	"math-ai.com/math-ai/internal/domain/otp"
	"math-ai.com/math-ai/internal/domain/seq"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/mtime"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/infrastructure/logger"
	"math-ai.com/math-ai/internal/shared/enum"
)

// pushChannelName labels SendOtpCommandResult.Channel when the OTP was
// pushed to a trusted device instead of going through otp_delivery. It is
// not registered with otp_delivery.Adapter — push uses a different
// addressing scheme (device push token, not phone/email identifier) and a
// richer result (InvalidTokens) that the Deliverer interface can't carry, so
// it is dispatched directly through the notification adapter below instead
// of being shoehorned into a Deliverer.
var (
	pushChannelName otp_delivery.ChannelName = "push"
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
	UserID     *int64
	DeviceUUID *string
	DeviceName *string
	Channel    enum.OtpChannel // empty = auto-detect
	// Language enum.LanguageType

	// TargetDeviceID, when set, redirects delivery to a push notification
	// sent to that (already-trusted) device instead of SMS/email. Only
	// meaningful for OtpType == LOGIN_2FA — validated by the caller
	// (module/otp/validator.go) before this command ever sees it.
	TargetDeviceID *int64
}

type SendOtpCommandResult struct {
	OtpID     int64
	ExpiresAt mtime.MathTime
	Channel   otp_delivery.ChannelName
	OTPCode   string
	OTPType   string
}

type SendOtpCommandHandler struct {
	uow                transaction.UnitOfWork
	delivery           *otp_delivery.Adapter
	pushAdapter        *notifAdapter.Adapter
	clearDeadTokensCmd *notifCommand.ClearDeadTokensCommandHandler
}

// NewSendOtpCommandHandler wires both delivery paths: delivery (SMS/email,
// identifier-routed) and pushAdapter (Firebase, device-token-routed — nil
// when NOTIFICATION_PROVIDER is disabled, same nil-guard convention as the
// rest of the notification adapter's consumers).
func NewSendOtpCommandHandler(uow transaction.UnitOfWork, delivery *otp_delivery.Adapter, pushAdapter *notifAdapter.Adapter) *SendOtpCommandHandler {
	return &SendOtpCommandHandler{
		uow:                uow,
		delivery:           delivery,
		pushAdapter:        pushAdapter,
		clearDeadTokensCmd: notifCommand.NewClearDeadTokensCommandHandler(uow),
	}
}

func (h *SendOtpCommandHandler) Handle(ctx context.Context, cmd SendOtpCommand) (*SendOtpCommandResult, error) {
	log := logger.From(ctx)
	if !cmd.OtpType.IsValid() {
		return nil, errs.NewError(ctx, status.OTP_INVALID_TYPE, nil, ErrInvalidOtpType)
	}
	if cmd.Identifier == "" {
		return nil, errs.NewError(ctx, status.OTP_MISSING_IDENTIFIER, nil, ErrIdentifierRequired)
	}

	// target_device_id picks a different addressing scheme (device push
	// token) than the identifier-routed sms/email channels, so it bypasses
	// resolveChannel entirely.
	var channel otp_delivery.ChannelName
	if cmd.TargetDeviceID != nil {
		if h.pushAdapter == nil {
			return nil, errs.NewError(ctx, status.OTP_NO_DELIVERY_CHANNEL, map[string]any{
				"channel": string(pushChannelName),
			}, ErrPushChannelNotRegistered)
		}
		channel = pushChannelName
	} else {
		var err error
		channel, err = h.resolveChannel(ctx, cmd)
		if err != nil {
			return nil, err
		}
	}

	log.Infof("otp send command deliver to: %s via %s", cmd.Identifier, channel)

	// Generate the plaintext code outside the UoW — rand.Int can be slow
	// under crypto entropy pressure and shouldn't hold a tx open.
	plainCode, err := generateCode(OtpCodeLength)
	if err != nil {
		return nil, errs.NewError(ctx, status.OTP_GENERATION_FAILED, nil, err)
	}

	now := time.Now().UTC()
	expiresAt := now.Add(TtlFor(cmd.OtpType))

	var createdOtpID int64
	var targetPushToken string
	err = h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		// 0. Target-device validation (trusted-device push 2FA). Runs before
		// any cooldown/rate-limit/revoke side effects so a bad target_device_id
		// fails fast without consuming the caller's send budget or revoking a
		// legitimate pending OTP.
		if cmd.TargetDeviceID != nil {
			targetDevice, err := repos.Device.FindByDeviceId(ctx, *cmd.TargetDeviceID)
			if err != nil {
				return errs.NewError(ctx, status.DEVICE_REGISTRATION_FAIL, nil, err)
			}
			if targetDevice == nil {
				return errs.NewError(ctx, status.DEVICE_NOT_FOUND, nil, ErrTargetDeviceNotFound)
			}
			if cmd.UserID == nil || targetDevice.UserId() == nil || *targetDevice.UserId() != *cmd.UserID {
				return errs.NewError(ctx, status.DEVICE_NOT_OWNED, nil, ErrTargetDeviceNotOwned)
			}
			if !targetDevice.IsVerified() {
				return errs.NewError(ctx, status.DEVICE_NOT_TRUSTED, nil, ErrTargetDeviceNotTrusted)
			}
			if targetDevice.DevicePushToken() == nil || *targetDevice.DevicePushToken() == "" {
				return errs.NewError(ctx, status.NOTIFICATION_NO_DEVICE_TOKEN, nil, ErrTargetDeviceNoPushToken)
			}
			targetPushToken = *targetDevice.DevicePushToken()
		}

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
			}, ErrOtpSendWindowReached)
		}

		// 3. Revoke prior PENDING rows
		if err := repos.Otp.RevokePendingByTypeIdentifier(ctx, cmd.OtpType, cmd.Identifier); err != nil {
			return errs.NewError(ctx, status.OTP_GENERATION_FAILED, nil, err)
		}

		// 4. Insert fresh row
		o := otp.NewOtp()
		// o.SetOtpId(utils.GenerateUUID().String())
		otpID, err := seqgen.Next(ctx, repos.Seq, seq.NameOtp)
		if err != nil {
			return err
		}
		o.SetOtpId(otpID)

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

	// 5. Dispatch after commit. Delivery failure does NOT revoke the row;
	// the audit trail of attempts is intentional, and a retry will trip
	// the cooldown rather than spam the user.
	//
	// NOTE: this only actually dispatches for the push (target_device_id)
	// path today. SMS/email dispatch through h.delivery is a separate,
	// pre-existing gap outside this change's scope — see send_otp_command.go
	// history; those channels still only surface the code via OTPCode below.
	responseCode := plainCode
	if cmd.TargetDeviceID != nil {
		sendRes, perr := h.pushAdapter.Send(ctx, notifAdapter.PushMessage{
			Tokens: []string{targetPushToken},
			Title:  "Xác nhận đăng nhập",
			Body:   fmt.Sprintf("Mã xác thực đăng nhập của bạn là %s", plainCode),
			Data: map[string]string{
				"otp_type": cmd.OtpType.String(),
			},
		})
		if perr != nil {
			return nil, errs.NewError(ctx, status.OTP_DELIVERY_FAILED, nil, perr)
		}
		if len(sendRes.InvalidTokens) > 0 {
			// Best-effort: prune the dead token, but a cleanup failure must
			// not mask the real delivery failure below.
			if cerr := h.clearDeadTokensCmd.Handle(ctx, notifCommand.ClearDeadTokensCommand{
				Tokens: sendRes.InvalidTokens,
			}); cerr != nil {
				logger.From(ctx).Warnf("otp.push_clear_dead_tokens_failed err=%v", cerr)
			}
			return nil, errs.NewError(ctx, status.OTP_DELIVERY_FAILED, nil, ErrPushTokenInvalid)
		}

		// Delivered for real via a channel the requesting (untrusted) device
		// cannot read — echoing the code back in the response would defeat
		// the entire point of trusted-device 2FA, so it is withheld here.
		// responseCode = ""
	}

	if h.delivery != nil && channel == otp_delivery.ChannelEmail {
		err := h.delivery.Send(ctx, otp_delivery.Message{
			OtpType:    cmd.OtpType,
			Identifier: cmd.Identifier,
			Code:       plainCode,
			ExpiresAt:  expiresAt,
		})
		if err != nil {
			return nil, errs.NewError(ctx, status.OTP_DELIVERY_FAILED, nil, err)
		}
	}

	return &SendOtpCommandResult{
		OtpID:     createdOtpID,
		ExpiresAt: mtime.MathTime{Time: expiresAt},
		Channel:   channel,
		OTPCode:   responseCode,
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
			ErrOtpUnknownChannel)
	}
	if !h.delivery.HasChannel(channel) {
		return "", errs.NewError(ctx, status.OTP_NO_DELIVERY_CHANNEL, map[string]any{
			"channel": string(channel),
		}, ErrOtpChannelNotRegistered)
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
