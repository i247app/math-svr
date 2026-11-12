package device

import (
	"time"
)

// Device is a (user, device_uuid) registration. The is_verified flag is the
// trust gate that auth/login consults to decide whether 2FA is required.
//
// The schema allows uid to be NULL (pre-binding registrations), so it is
// modeled as a *string. In practice the device row is created during login,
// at which point uid is always known — but the domain stays faithful to
// the column nullability.
type Device struct {
	deviceId        int64
	uid             *int64
	deviceUUID      string
	deviceName      string
	platform        string
	devicePushToken *string
	isVerified      bool
	trustDt         time.Time
	rptFlg          *string
	kwords          *string
	note            *string
	deviceStatus    *string
	status          string
	createId        *int64
	createDt        time.Time
	modifyId        *int64
	modifyDt        time.Time
}

func NewDevice() *Device {
	return &Device{}
}

func (d *Device) DeviceId() int64                 { return d.deviceId }
func (d *Device) SetDeviceId(deviceId int64)      { d.deviceId = deviceId }
func (d *Device) Uid() *int64                     { return d.uid }
func (d *Device) SetUid(uid *int64)               { d.uid = uid }
func (d *Device) DeviceUUID() string              { return d.deviceUUID }
func (d *Device) SetDeviceUUID(deviceUUID string) { d.deviceUUID = deviceUUID }
func (d *Device) DeviceName() string              { return d.deviceName }
func (d *Device) SetDeviceName(deviceName string) { d.deviceName = deviceName }

// Platform is the client platform (enum.PlatformType, e.g. "IOS") this
// device row was registered from. Set once at creation — see BuildDevice
// (login) and MarkDeviceVerifiedCommandHandler (2FA/registration auto-create).
func (d *Device) Platform() string            { return d.platform }
func (d *Device) SetPlatform(platform string) { d.platform = platform }
func (d *Device) DevicePushToken() *string {
	return d.devicePushToken
}

func (d *Device) SetDevicePushToken(devicePushToken *string) {
	d.devicePushToken = devicePushToken
}

func (d *Device) IsVerified() bool {
	return d.isVerified
}

func (d *Device) SetIsVerified(isVerified bool) {
	d.isVerified = isVerified
}

func (d *Device) TrustDt() time.Time {
	return d.trustDt
}

func (d *Device) SetTrustDt(trustDt time.Time) {
	d.trustDt = trustDt
}

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

func (d *Device) Note() *string {
	return d.note
}

func (d *Device) SetNote(note *string) {
	d.note = note
}

// RptFlg is the client-side report flag (rpt_flg). Nil means "not reported".
func (d *Device) RptFlg() *string {
	return d.rptFlg
}

func (d *Device) SetRptFlg(rptFlg *string) {
	d.rptFlg = rptFlg
}

// Kwords holds search keywords (kwords) for a future text-search index.
func (d *Device) Kwords() *string {
	return d.kwords
}

func (d *Device) SetKwords(kwords *string) {
	d.kwords = kwords
}

func (d *Device) DeviceStatus() *string {
	return d.deviceStatus
}

func (d *Device) SetDeviceStatus(deviceStatus *string) {
	d.deviceStatus = deviceStatus
}

func (d *Device) Status() string {
	return d.status
}

func (d *Device) SetStatus(status string) {
	d.status = status
}

func (d *Device) CreateId() *int64 {
	return d.createId
}

func (d *Device) SetCreateId(createId *int64) {
	d.createId = createId
}

func (d *Device) CreateDt() time.Time {
	return d.createDt
}

func (d *Device) SetCreateDt(createDt time.Time) {
	d.createDt = createDt
}

func (d *Device) ModifyId() *int64 {
	return d.modifyId
}

func (d *Device) SetModifyId(modifyId *int64) {
	d.modifyId = modifyId
}

func (d *Device) ModifyDt() time.Time {
	return d.modifyDt
}

func (d *Device) SetModifyDt(modifyDt time.Time) {
	d.modifyDt = modifyDt
}
