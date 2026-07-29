package enum

import "strings"

// PlatformType is the client platform a device row was registered from.
type PlatformType string

const (
	PlatformTypeIOS     PlatformType = "IOS"
	PlatformTypeAndroid PlatformType = "ANDROID"
	PlatformTypeWeb     PlatformType = "WEB"
	// PlatformTypeUnknown covers legacy rows (backfilled) and requests where
	// the client omitted or sent an unrecognized platform. Deliberately not
	// a validation failure — platform is descriptive metadata, not a gate.
	PlatformTypeUnknown PlatformType = "UNKNOWN"
)

func (p PlatformType) String() string {
	return string(p)
}

func (p PlatformType) IsValid() bool {
	switch p {
	case PlatformTypeIOS, PlatformTypeAndroid, PlatformTypeWeb, PlatformTypeUnknown:
		return true
	default:
		return false
	}
}

// ParsePlatformType normalizes a client-reported platform string (the
// metadata object carries it lowercase, e.g. "ios"/"android"/"web") into the
// canonical enum. Empty or unrecognized input maps to PlatformTypeUnknown
// rather than erroring, so an older client or an unexpected value never
// blocks login/registration.
func ParsePlatformType(raw string) PlatformType {
	switch strings.ToUpper(strings.TrimSpace(raw)) {
	case string(PlatformTypeIOS):
		return PlatformTypeIOS
	case string(PlatformTypeAndroid):
		return PlatformTypeAndroid
	case string(PlatformTypeWeb):
		return PlatformTypeWeb
	default:
		return PlatformTypeUnknown
	}
}
