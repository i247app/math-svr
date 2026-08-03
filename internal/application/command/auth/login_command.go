package command

import (
	"context"
	"strings"
	"time"

	"math-ai.com/math-ai/internal/application/command/shared/seqgen"
	"math-ai.com/math-ai/internal/application/transaction"
	"math-ai.com/math-ai/internal/domain/device"
	"math-ai.com/math-ai/internal/domain/loginlog"
	"math-ai.com/math-ai/internal/domain/seq"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/enum"
)

type LoginCommand struct {
	LoginName       string
	DeviceUUID      string
	DeviceName      string
	Platform        string
	IPAddress       string
	DevicePushToken string
}

// LoginCommandResult communicates one of two outcomes:
//
//   - TwoFactorRequired=true → device is new or not yet trusted; the caller
//     must run the 2FA challenge. No session is created. UserID and DeviceID
//     are populated so the 2FA endpoint can target the right device.
//
//   - TwoFactorRequired=false → device is trusted; a fresh login_log row was
//     inserted (any prior active session for the same device was revoked).
//     LoginLogID identifies the new session.
type LoginCommandResult struct {
	UserID          int64
	DeviceID        int64
	LoginLogID      int64
	IsTrustedDevice bool
}

type LoginCommandHandler struct {
	uow                transaction.UnitOfWork
	trustDeviceTTLDays int
}

func NewLoginCommandHandler(uow transaction.UnitOfWork, trustDeviceTTLDays int) *LoginCommandHandler {
	return &LoginCommandHandler{uow: uow, trustDeviceTTLDays: trustDeviceTTLDays}
}

// Handle resolves the user by phone, ensures a device registration exists for
// (user, device_uuid), and branches on the device's trust state:
//
//   - First-sight or untrusted device → return TwoFactorRequired=true and
//     stop. The device row is persisted (or already existed) so the future
//     2FA endpoint can flip is_verified once the challenge passes.
//   - Trusted device → revoke any prior active session for the same device
//     and insert a fresh login_log row carrying the new opaque token.
//
// The entire sequence runs inside one UoW so partial failures can't leak
// state (e.g. an orphan device row paired with no session).
func (h *LoginCommandHandler) Handle(ctx context.Context, cmd LoginCommand) (*LoginCommandResult, error) {
	var result *LoginCommandResult

	err := h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		u, err := repos.User.FindByLoginName(ctx, cmd.LoginName)
		if err != nil {
			return errs.NewError(ctx, status.AUTH_LOGIN_FAILED, nil, err)
		}

		if u == nil {
			return nil
		}

		result = &LoginCommandResult{
			UserID:          u.UserId(),
			IsTrustedDevice: true,
		}
		// return nil

		d, err := ensureDevice(ctx, repos, u.UserId(), cmd)
		if err != nil {
			return err
		}

		result.DeviceID = d.DeviceId()
		result.IsTrustedDevice = d.IsVerified() && !d.IsTrustExpired(h.trustDeviceTTLDays, time.Now().UTC())

		if result.IsTrustedDevice {
			err := repos.Device.MarkVerifiedByUserDevice(ctx, u.UserId(), cmd.DeviceUUID, true)
			if err != nil {
				return errs.NewError(ctx, status.DEVICE_REGISTRATION_FAIL, nil, err)
			}
		}

		// if !d.IsVerified() {
		// 	result = &LoginCommandResult{
		// 		UserID:            u.UserId(),
		// 		DeviceID:          d.DeviceId(),
		// 		TwoFactorRequired: d.IsVerified(),
		// 	}
		// 	return nil
		// }

		// if err := repos.LoginLog.MarkStatusByUserDevice(
		// 	ctx, u.UserId(), cmd.DeviceUUID, enum.LoginLogStatusTypeRevoked,
		// ); err != nil {
		// 	return errs.NewError(ctx, status.AUTH_LOGIN_FAILED, nil, err)
		// }

		// ll := BuildLoginLog(u.UserId(), cmd)
		// created, err := repos.LoginLog.Create(ctx, ll)
		// if err != nil {
		// 	return errs.NewError(ctx, status.AUTH_LOGIN_FAILED, nil, err)
		// }

		// result = &LoginCommandResult{
		// 	UserID:     u.UserId(),
		// 	DeviceID:   d.DeviceId(),
		// 	LoginLogID: created.LoginLogId(),
		// }

		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// ensureDevice returns the (user, device_uuid) registration, creating a fresh
// untrusted row on first sight. New rows always start with is_verified=false
// so the very first login from a previously-unseen device is forced through
// 2FA. Push-token freshening for known devices happens here too — a returning
// device may have rotated its token since the last login.
func ensureDevice(
	ctx context.Context,
	repos transaction.Repositories,
	userId int64,
	cmd LoginCommand,
) (*device.Device, error) {
	existing, err := repos.Device.FindByUserDevice(ctx, userId, cmd.DeviceUUID)
	if err != nil {
		return nil, errs.NewError(ctx, status.DEVICE_REGISTRATION_FAIL, nil, err)
	}
	if existing != nil {
		// Refresh the push token if the client sent a new value. Leaving
		// it stale would mean push notifications silently fail.
		if cmd.DevicePushToken != "" {
			current := ""
			if existing.DevicePushToken() != nil {
				current = *existing.DevicePushToken()
			}
			if current != cmd.DevicePushToken {
				patch := device.NewDevice()
				patch.SetDeviceId(existing.DeviceId())
				token := cmd.DevicePushToken
				patch.SetDevicePushToken(&token)
				if err := repos.Device.Update(ctx, patch); err != nil {
					return nil, errs.NewError(ctx, status.DEVICE_REGISTRATION_FAIL, nil, err)
				}
			}
		}
		return existing, nil
	}

	deviceBuild := BuildDevice(userId, cmd)
	deviceID, err := seqgen.Next(ctx, repos.Seq, seq.NameDevice)
	if err != nil {
		return nil, err
	}
	deviceBuild.SetDeviceId(deviceID)

	created, err := repos.Device.Create(ctx, deviceBuild)
	if err != nil {
		return nil, errs.NewError(ctx, status.DEVICE_REGISTRATION_FAIL, nil, err)
	}
	return created, nil
}

func BuildDevice(userId int64, cmd LoginCommand) *device.Device {
	d := device.NewDevice()
	d.SetUserId(&userId)
	d.SetDeviceUUID(cmd.DeviceUUID)
	d.SetDeviceName(cmd.DeviceName)
	d.SetPlatform(enum.ParsePlatformType(cmd.Platform).String())
	if cmd.DevicePushToken != "" {
		token := cmd.DevicePushToken
		d.SetDevicePushToken(&token)
	}
	d.SetIsVerified(false)
	active := enum.DeviceStatusTypeActive.String()
	d.SetDeviceStatus(&active)
	d.SetStatus(enum.StatusActive.String())

	return d
}

func BuildLoginLog(userId int64, cmd LoginCommand) *loginlog.LoginLog {
	ll := loginlog.NewLoginLog()
	ll.SetUserId(userId)
	ll.SetDeviceUUID(cmd.DeviceUUID)
	ll.SetIpAddress(cmd.IPAddress)
	ll.SetToken(cmd.DevicePushToken)
	active := enum.LoginLogStatusTypeActive.String()
	ll.SetLoginLogStatus(&active)
	ll.SetStatus(enum.StatusActive.String())
	return ll
}

func ValidateLoginCommand(ctx context.Context, cmd LoginCommand) error {
	if strings.TrimSpace(cmd.DeviceUUID) == "" {
		return errs.NewError(ctx, status.AUTH_MISSING_DEVICE_UUID, nil, ErrDeviceUUIDRequired)
	}
	if strings.TrimSpace(cmd.IPAddress) == "" {
		return errs.NewError(ctx, status.AUTH_MISSING_IP_ADDRESS, nil, ErrIPAddressRequired)
	}
	if strings.TrimSpace(cmd.DevicePushToken) == "" {
		return errs.NewError(ctx, status.AUTH_MISSING_DEVICE_PUSH_TOKEN, nil, ErrDevicePushTokenRequired)
	}
	return nil
}
