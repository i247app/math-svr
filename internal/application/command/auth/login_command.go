package command

import (
	"context"
	"strings"
	"time"

	"math-ai.com/math-ai/internal/application/command/shared/seqgen"
	"math-ai.com/math-ai/internal/application/transaction"
	"math-ai.com/math-ai/internal/domain/device"
	"math-ai.com/math-ai/internal/domain/login"
	"math-ai.com/math-ai/internal/domain/loginlog"
	"math-ai.com/math-ai/internal/domain/seq"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/domain/user"
	"math-ai.com/math-ai/internal/shared/enum"
	"math-ai.com/math-ai/internal/shared/utils"
)

type LoginCommand struct {
	LoginName       string
	DeviceUUID      string
	DeviceName      string
	Platform        string
	IPAddress       string
	DevicePushToken string

	// Password is optional. Empty keeps the OTP-only login exactly as it
	// was. When set it must match the account's ma_logins credential —
	// an account with no password set fails too — and only then does the
	// usual device-trust flow run, so an untrusted device still has to
	// pass OTP afterwards.
	Password string
}

// LoginCommandResult communicates one of two outcomes:
//
//   - TwoFactorRequired=true → device is new or not yet trusted; the caller
//     must run the 2FA challenge. No session is created. UID and DeviceID
//     are populated so the 2FA endpoint can target the right device.
//
//   - TwoFactorRequired=false → device is trusted; a fresh login_log row was
//     inserted (any prior active session for the same device was revoked).
//     LoginLogID identifies the new session.
type LoginCommandResult struct {
	UID             int64
	DeviceID        int64
	LoginLogID      int64
	IsTrustedDevice bool
}

type LoginCommandHandler struct {
	uow                transaction.UnitOfWork
	hasher             login.PasswordHasher
	trustDeviceTTLDays int
}

func NewLoginCommandHandler(uow transaction.UnitOfWork, hasher login.PasswordHasher, trustDeviceTTLDays int) *LoginCommandHandler {
	return &LoginCommandHandler{uow: uow, hasher: hasher, trustDeviceTTLDays: trustDeviceTTLDays}
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
	if cmd.Password != "" {
		if err := h.verifyPassword(ctx, cmd); err != nil {
			return nil, err
		}
	}

	var result *LoginCommandResult

	err := h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		u, err := resolveUser(ctx, repos, cmd.LoginName)
		if err != nil || u == nil {
			return err
		}

		result = &LoginCommandResult{
			UID:             u.Uid(),
			IsTrustedDevice: true,
		}
		// return nil

		d, err := ensureDevice(ctx, repos, u.Uid(), cmd)
		if err != nil {
			return err
		}

		result.DeviceID = d.DeviceId()
		result.IsTrustedDevice = d.IsVerified() && !d.IsTrustExpired(h.trustDeviceTTLDays, time.Now().UTC())

		if result.IsTrustedDevice {
			err := repos.Device.MarkVerifiedByUserDevice(ctx, u.Uid(), cmd.DeviceUUID, true)
			if err != nil {
				return errs.NewError(ctx, status.DEVICE_REGISTRATION_FAIL, nil, err)
			}

			ll, err := repos.LoginLog.FindActiveByUserDevice(ctx, u.Uid(), cmd.DeviceUUID)
			if err != nil {
				return errs.NewError(ctx, status.AUTH_LOGIN_FAILED, nil, err)
			}

			if ll != nil {
				result.LoginLogID = ll.LoginLogId()
				if err := repos.LoginLog.MarkStatusByUserDevice(ctx, u.Uid(), cmd.DeviceUUID, enum.LoginLogStatusTypeActive); err != nil {
					return errs.NewError(ctx, status.AUTH_LOGIN_FAILED, nil, err)
				}
			} else {
				loginLogId, err := seqgen.Next(ctx, repos.Seq, seq.NameLoginLog)
				if err != nil {
					return err
				}
				ll := BuildLoginLog(u.Uid(), cmd)
				ll.SetLoginLogId(loginLogId)
				_, err = repos.LoginLog.Create(ctx, ll)
				if err != nil {
					return errs.NewError(ctx, status.AUTH_LOGIN_FAILED, nil, err)
				}
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// verifyPassword checks cmd.Password against the account's stored hash.
// It reads in its own short transaction and compares after it has
// committed: bcrypt is deliberately slow, and comparing inside the main
// transaction would hold a pooled connection for every guess.
//
// An unknown login name passes through untouched — the main flow then
// returns its usual "no account" nil, so a password does not change what
// the endpoint reveals about which names exist.
func (h *LoginCommandHandler) verifyPassword(ctx context.Context, cmd LoginCommand) error {
	var (
		found bool
		hash  string
	)

	handler := func(ctx context.Context, repos transaction.Repositories) error {
		u, err := resolveUser(ctx, repos, cmd.LoginName)
		if err != nil || u == nil {
			return err
		}
		found = true
		cred, err := repos.Login.FindByUid(ctx, u.Uid())
		if err != nil {
			return errs.NewError(ctx, status.AUTH_LOGIN_FAILED, nil, err)
		}
		if cred != nil && cred.Upw() != nil {
			hash = *cred.Upw()
		}
		return nil
	}

	err := h.uow.Do(ctx, handler)
	if err != nil || !found {
		return err
	}

	// No password on file is the same answer as a wrong one: the caller
	// asked to be checked against a credential this account does not have.
	if hash == "" || h.hasher.Compare(cmd.Password, hash) != nil {
		return errs.NewError(ctx, status.AUTH_INVALID_CREDENTIALS, nil, ErrInvalidCredentials)
	}
	return nil
}

// resolveUser maps a login name to the account it signs in to. The name
// (phone or email) is resolved through the alias registry first, then the
// user row — the user repo never JOINs ma_aliases. Either lookup missing
// means "no such account": (nil, nil), so the caller keeps the
// enumeration-safe response.
//
// A guest is registered in ma_aliases under their device_uuid, so that
// string is a resolvable login name — and anyone who learns it could
// otherwise start a login against someone else's guest account. A guest
// has no login of their own until they register (which converts the row
// to USER), so they resolve to no account at all, with the same nil.
func resolveUser(ctx context.Context, repos transaction.Repositories, loginName string) (*user.User, error) {
	alias, err := repos.Alias.FindByAka(ctx, loginName)
	if err != nil {
		return nil, errs.NewError(ctx, status.AUTH_LOGIN_FAILED, nil, err)
	}
	if alias == nil {
		return nil, nil
	}

	u, err := repos.User.FindByUid(ctx, alias.Uid())
	if err != nil {
		return nil, errs.NewError(ctx, status.AUTH_LOGIN_FAILED, nil, err)
	}
	if u == nil || enum.IdentityCodeType(utils.DerefString(u.IdentityCode())).IsGuest() {
		return nil, nil
	}
	return u, nil
}

// ensureDevice returns the (user, device_uuid) registration, creating a fresh
// untrusted row on first sight. New rows always start with is_verified=false
// so the very first login from a previously-unseen device is forced through
// 2FA. Push-token freshening for known devices happens here too — a returning
// device may have rotated its token since the last login.
func ensureDevice(
	ctx context.Context,
	repos transaction.Repositories,
	uid int64,
	cmd LoginCommand,
) (*device.Device, error) {
	existing, err := repos.Device.FindByUserDevice(ctx, uid, cmd.DeviceUUID)
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

	deviceBuild := BuildDevice(uid, cmd)
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

func BuildDevice(uid int64, cmd LoginCommand) *device.Device {
	d := device.NewDevice()
	d.SetUid(&uid)
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

func BuildLoginLog(uid int64, cmd LoginCommand) *loginlog.LoginLog {
	ll := loginlog.NewLoginLog()
	ll.SetUid(uid)
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
