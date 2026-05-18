package enum

type OtpType string

const (
	OtpTypeLogin2FA       OtpType = "LOGIN_2FA"
	OtpTypeRegister       OtpType = "REGISTER"
	OtpTypeForgotPassword OtpType = "FORGOT_PASSWORD"
	OtpTypeChangePassword OtpType = "CHANGE_PASSWORD"
	OtpTypeVerifyEmail    OtpType = "VERIFY_EMAIL"
	OtpTypeVerifyPhone    OtpType = "VERIFY_PHONE"
)

func (t OtpType) String() string {
	return string(t)
}

func (t OtpType) IsValid() bool {
	switch t {
	case OtpTypeLogin2FA, OtpTypeRegister, OtpTypeForgotPassword,
		OtpTypeChangePassword, OtpTypeVerifyEmail, OtpTypeVerifyPhone:
		return true
	default:
		return false
	}
}

type OtpStatusType string

const (
	OtpStatusTypePending  OtpStatusType = "PENDING"
	OtpStatusTypeVerified OtpStatusType = "VERIFIED"
	OtpStatusTypeExpired  OtpStatusType = "EXPIRED"
	OtpStatusTypeRevoked  OtpStatusType = "REVOKED"
)

func (s OtpStatusType) String() string {
	return string(s)
}

func (s OtpStatusType) IsValid() bool {
	switch s {
	case OtpStatusTypePending, OtpStatusTypeVerified, OtpStatusTypeExpired, OtpStatusTypeRevoked:
		return true
	default:
		return false
	}
}

// OtpChannel selects how a code is delivered. Empty = auto-detect from
// identifier shape (email if it contains "@", else SMS).
type OtpChannel string

const (
	OtpChannelAuto  OtpChannel = ""
	OtpChannelSMS   OtpChannel = "SMS"
	OtpChannelEmail OtpChannel = "EMAIL"
)

func (c OtpChannel) String() string {
	return string(c)
}

func (c OtpChannel) IsValid() bool {
	switch c {
	case OtpChannelAuto, OtpChannelSMS, OtpChannelEmail:
		return true
	default:
		return false
	}
}
