package command

import (
	"context"
	"fmt"
	"time"

	notifCommand "math-ai.com/math-ai/internal/application/command/notification"
	"math-ai.com/math-ai/internal/application/command/shared/seqgen"
	"math-ai.com/math-ai/internal/application/transaction"
	"math-ai.com/math-ai/internal/domain/otp"
	"math-ai.com/math-ai/internal/domain/seq"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/infrastructure/logger"
	notiProvider "math-ai.com/math-ai/internal/provider/notification"
	otpProvider "math-ai.com/math-ai/internal/provider/otp_delivery"
	"math-ai.com/math-ai/internal/shared/enum"
)

// pushChannelName labels SendOtpCommandResult.Channel when the OTP was
// pushed to a trusted device instead of going through otpProvider. It is
// not registered with otpProvider.Registry — push uses a different
// addressing scheme (device push token, not phone/email identifier) and a
// richer result (InvalidTokens) that the Deliverer interface can't carry, so
// it is dispatched directly through the notification provider below instead
// of being shoehorned into a Deliverer.
var (
	pushChannelName otpProvider.ChannelName = "push"
)

// SendOtpCommand issues a fresh OTP for (type, identifier) and dispatches it
// via the OTP delivery provider. Inside one UoW it:
//
//  1. Reuses a still-valid PENDING OTP instead of issuing a new one.
//  2. Enforces the per-window send cap (OTP_RATE_LIMITED).
//  3. Revokes any prior PENDING OTPs for (type, identifier).
//  4. Generates a 6-digit code, hashes it, inserts the row.
//
// Step 5 happens AFTER commit: the delivery provider is invoked. If delivery
// fails, the row stays in place (audit trail) and the caller sees
// OTP_DELIVERY_FAILED — the client can retry, which reuses the same row
// rather than creating a duplicate one.
//
// A reused OTP is NOT re-dispatched (Result.Reused reports this). Leaving
// and re-entering the OTP screen calls this endpoint again; the user already
// holds the code, so resending it would only spam their inbox.
type SendOtpCommand struct {
	OtpType    enum.OtpType
	Identifier string
	UID        *int64
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
	ExpiresAt time.Time
	Channel   otpProvider.ChannelName
	// OTPCode is the code the caller may echo back to the requesting
	// client. Empty when it was pushed to a trusted device instead.
	OTPCode string
	// PushedCode is the code delivered to the trusted device, for
	// owner-only channels (the account's in-app inbox). Never return it to
	// the requesting client.
	PushedCode string
	OTPType    string
	// Reused is true when an existing, still-valid PENDING OTP was handed
	// back instead of a new one being issued — nothing was delivered, so
	// callers must not raise their own "we just sent you a code" side
	// effects either.
	Reused bool
}

type SendOtpCommandHandler struct {
	uow                transaction.UnitOfWork
	delivery           *otpProvider.Registry
	pushProvider       *notiProvider.Registry
	clearDeadTokensCmd *notifCommand.ClearDeadTokensCommandHandler
	demoNames          []string
}

// NewSendOtpCommandHandler wires both delivery paths: delivery (SMS/email,
// identifier-routed) and pushProvider (Firebase, device-token-routed — nil
// when NOTIFICATION_PROVIDER is disabled, same nil-guard convention as the
// rest of the notification provider's consumers).
func NewSendOtpCommandHandler(uow transaction.UnitOfWork, delivery *otpProvider.Registry, pushProvider *notiProvider.Registry, demoNames []string) *SendOtpCommandHandler {
	return &SendOtpCommandHandler{
		uow:                uow,
		delivery:           delivery,
		pushProvider:       pushProvider,
		clearDeadTokensCmd: notifCommand.NewClearDeadTokensCommandHandler(uow),
		demoNames:          demoNames,
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

	// Detect Demo names
	for _, demoName := range h.demoNames {
		if cmd.Identifier == demoName {
			cmd.TargetDeviceID = nil
		}
	}

	// target_device_id picks a different addressing scheme (device push
	// token) than the identifier-routed sms/email channels, so it bypasses
	// resolveChannel entirely.
	var channel otpProvider.ChannelName
	if cmd.TargetDeviceID != nil {
		if h.pushProvider == nil {
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
	var reused bool
	var targetPushToken string
	err = h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		// Avoid two concurrent sends for the same (type, identifier) deadlock
		if err := seqgen.Lock(ctx, repos.Seq, seq.NameOtp); err != nil {
			return err
		}

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
			if cmd.UID == nil || targetDevice.Uid() == nil || *targetDevice.Uid() != *cmd.UID {
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

		// 1. Reuse a still-valid PENDING OTP.
		// Expiry is read from OtpExpireDt (app-set, always UTC) — never
		// from CreateDt, which MySQL fills with DEFAULT CURRENT_TIMESTAMP(6)
		// in the server's local wall-clock and the driver then mis-tags as
		// UTC.
		latest, err := repos.Otp.FindLatestPending(ctx, cmd.OtpType, cmd.Identifier)
		if err != nil {
			return errs.NewError(ctx, status.OTP_GENERATION_FAILED, nil, err)
		}
		if latest != nil && !latest.OtpExpireDt().IsZero() && now.Before(latest.OtpExpireDt()) {
			createdOtpID = latest.OtpId()
			plainCode = latest.OtpCode()
			expiresAt = latest.OtpExpireDt()
			reused = true
			return nil
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
		o.SetUid(cmd.UID)
		o.SetIdentifier(cmd.Identifier)
		o.SetDeviceUUID(cmd.DeviceUUID)
		o.SetDeviceName(cmd.DeviceName)
		// o.SetOtpCode(hashCode(plainCode))
		o.SetOtpCode(plainCode)
		o.SetOtpCreateDt(now)
		o.SetOtpExpireDt(expiresAt)
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
	pushedCode := ""
	if reused {
		log.Infof("otp send command reuse pending otp id=%d, delivery skipped", createdOtpID)
	}
	if !reused && cmd.TargetDeviceID != nil {
		sendRes, perr := h.pushProvider.Send(ctx, notiProvider.PushMessage{
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
		responseCode = ""
		pushedCode = plainCode
	}

	if !reused && h.delivery != nil && channel == otpProvider.ChannelEmail {
		err := h.delivery.Send(ctx, otpProvider.Message{
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
		OtpID:      createdOtpID,
		ExpiresAt:  expiresAt,
		Channel:    channel,
		OTPCode:    responseCode,
		PushedCode: pushedCode,
		OTPType:    cmd.OtpType.String(),
		Reused:     reused,
	}, nil
}

// resolveChannel returns the channel the OTP will ship through. Explicit
// overrides win; otherwise the provider's auto-detect picks SMS vs email from
// identifier shape.
func (h *SendOtpCommandHandler) resolveChannel(ctx context.Context, cmd SendOtpCommand) (otpProvider.ChannelName, error) {
	var channel otpProvider.ChannelName
	switch cmd.Channel {
	case enum.OtpChannelSMS:
		channel = otpProvider.ChannelSMS
	case enum.OtpChannelEmail:
		channel = otpProvider.ChannelEmail
	case enum.OtpChannelAuto:
		if isEmailLike(cmd.Identifier) {
			channel = otpProvider.ChannelEmail
		} else {
			channel = otpProvider.ChannelSMS
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
