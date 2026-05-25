package otp

import (
	mtime "math-ai.com/math-ai/internal/domain/shared/time"
)

// Otp is a one-time password issuance. The plaintext code never lives in the
// domain — only the hash that the repository persists to ma_otps.otp_code.
//
// userId / deviceUUID / deviceName are nullable because some flows (REGISTER,
// FORGOT_PASSWORD) don't have a known user or device when the OTP is issued.
// identifier (phone or email) is always present and is the lookup key on
// verify.
type Otp struct {
	id           int64
	otpId        string
	otpType      string
	userId       *string
	identifier   string
	deviceUUID   *string
	deviceName   *string
	otpCode      string
	otpCreateDt  mtime.MathTime
	otpExpireDt  mtime.MathTime
	attemptCount int
	note         *string
	otpStatus    *string
	status       string
	createId     *string
	createDt     mtime.MathTime
	modifyId     *string
	modifyDt     mtime.MathTime
}

func NewOtp() *Otp { return &Otp{} }

func (o *Otp) Id() int64                       { return o.id }
func (o *Otp) SetId(id int64)                  { o.id = id }
func (o *Otp) OtpId() string                   { return o.otpId }
func (o *Otp) SetOtpId(v string)               { o.otpId = v }
func (o *Otp) OtpType() string                 { return o.otpType }
func (o *Otp) SetOtpType(v string)             { o.otpType = v }
func (o *Otp) UserId() *string                 { return o.userId }
func (o *Otp) SetUserId(v *string)             { o.userId = v }
func (o *Otp) Identifier() string              { return o.identifier }
func (o *Otp) SetIdentifier(v string)          { o.identifier = v }
func (o *Otp) DeviceUUID() *string             { return o.deviceUUID }
func (o *Otp) SetDeviceUUID(v *string)         { o.deviceUUID = v }
func (o *Otp) DeviceName() *string             { return o.deviceName }
func (o *Otp) SetDeviceName(v *string)         { o.deviceName = v }
func (o *Otp) OtpCode() string                 { return o.otpCode }
func (o *Otp) SetOtpCode(v string)             { o.otpCode = v }
func (o *Otp) OtpCreateDt() mtime.MathTime     { return o.otpCreateDt }
func (o *Otp) SetOtpCreateDt(v mtime.MathTime) { o.otpCreateDt = v }
func (o *Otp) OtpExpireDt() mtime.MathTime     { return o.otpExpireDt }
func (o *Otp) SetOtpExpireDt(v mtime.MathTime) { o.otpExpireDt = v }
func (o *Otp) AttemptCount() int               { return o.attemptCount }
func (o *Otp) SetAttemptCount(v int)           { o.attemptCount = v }
func (o *Otp) Note() *string                   { return o.note }
func (o *Otp) SetNote(v *string)               { o.note = v }
func (o *Otp) OtpStatus() *string              { return o.otpStatus }
func (o *Otp) SetOtpStatus(v *string)          { o.otpStatus = v }
func (o *Otp) Status() string                  { return o.status }
func (o *Otp) SetStatus(v string)              { o.status = v }
func (o *Otp) CreateId() *string               { return o.createId }
func (o *Otp) SetCreateId(v *string)           { o.createId = v }
func (o *Otp) CreateDt() mtime.MathTime        { return o.createDt }
func (o *Otp) SetCreateDt(v mtime.MathTime)    { o.createDt = v }
func (o *Otp) ModifyId() *string               { return o.modifyId }
func (o *Otp) SetModifyId(v *string)           { o.modifyId = v }
func (o *Otp) ModifyDt() mtime.MathTime        { return o.modifyDt }
func (o *Otp) SetModifyDt(v mtime.MathTime)    { o.modifyDt = v }
