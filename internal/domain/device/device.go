package device

import (
	"time"

	"math-ai.com/math-ai/internal/domain/shared/audit"
)

// Device is a (user, device_uuid) registration. The is_verified flag is the
// trust gate that auth/login consults to decide whether 2FA is required.
//
// The schema allows uid to be NULL (pre-binding registrations), so it is
// modeled as a *string. In practice the device row is created during login,
// at which point uid is always known — but the domain stays faithful to
// the column nullability.
type Device struct {
	audit.Audit

	deviceId        int64
	uid             *int64
	deviceUUID      string
	deviceName      string
	platform        string
	devicePushToken *string
	isVerified      bool
	trustDt         time.Time
	deviceStatus    *string
}

func NewDevice() *Device {
	return &Device{}
}

func (d *Device) DeviceId() int64            { return d.deviceId }
func (d *Device) SetDeviceId(deviceId int64) { d.deviceId = deviceId }

func (d *Device) UID() *int64       { return d.uid }
func (d *Device) SetUID(uid *int64) { d.uid = uid }

func (d *Device) DeviceUUID() string              { return d.deviceUUID }
func (d *Device) SetDeviceUUID(deviceUUID string) { d.deviceUUID = deviceUUID }

func (d *Device) DeviceName() string              { return d.deviceName }
func (d *Device) SetDeviceName(deviceName string) { d.deviceName = deviceName }

// Platform is the client platform (enum.PlatformType, e.g. "IOS") this
// device row was registered from. Set once at creation — see BuildDevice
// (login) and MarkDeviceVerifiedCommandHandler (2FA/registration auto-create).
func (d *Device) Platform() string            { return d.platform }
func (d *Device) SetPlatform(platform string) { d.platform = platform }

func (d *Device) DevicePushToken() *string                   { return d.devicePushToken }
func (d *Device) SetDevicePushToken(devicePushToken *string) { d.devicePushToken = devicePushToken }

func (d *Device) IsVerified() bool              { return d.isVerified }
func (d *Device) SetIsVerified(isVerified bool) { d.isVerified = isVerified }

func (d *Device) TrustDt() time.Time           { return d.trustDt }
func (d *Device) SetTrustDt(trustDt time.Time) { d.trustDt = trustDt }

func (d *Device) DeviceStatus() *string                { return d.deviceStatus }
func (d *Device) SetDeviceStatus(deviceStatus *string) { d.deviceStatus = deviceStatus }

// IsTrustExpired reports whether the device's trust grant has aged past
// ttlDays. ttlDays <= 0 disables the TTL (trust never expires by age — the
// device still has to be independently marked verified). A device that was
// never trusted (zero-value trustDt) is treated as expired so it falls back
// to the 2FA gate.
func (d *Device) IsTrustExpired(ttlDays int, now time.Time) bool {
	if ttlDays <= 0 {
		return false
	}
	if d.trustDt.IsZero() {
		return true
	}
	return now.After(d.trustDt.AddDate(0, 0, ttlDays))
}
