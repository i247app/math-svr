package metadata

import (
	"math-ai.com/math-ai/internal/domain/shared/mtime"
	"math-ai.com/math-ai/internal/shared/enum"
)

type ClientLanguage string

const (
	ClientLanguageViVN ClientLanguage = "vi-VN"
	ClientLanguageEnEN ClientLanguage = "en-EN"
	ClientLanguageVi   ClientLanguage = "vi"
	ClientLanguageEn   ClientLanguage = "en"
)

func (c ClientLanguage) String() string {
	return string(c)
}

// ToEnumLanguage maps the client-reported language onto the internal
// enum. Both the short ("vi"/"en") and long ("vi-VN"/"en-EN") forms are
// accepted because the mobile app sends the short code. Unknown / empty
// values fall back to Vietnamese — the project default (see CLAUDE.md).
func (c ClientLanguage) ToEnumLanguage() enum.LanguageType {
	switch c {
	case ClientLanguageViVN, ClientLanguageVi:
		return enum.LanguageTypeVietnamese
	case ClientLanguageEnEN, ClientLanguageEn:
		return enum.LanguageTypeEnglish
	}
	return enum.LanguageTypeVietnamese
}

// RequestMetadata is the `metadata` object every client request carries in
// its body (JSON object, or a `metadata` form field for multipart uploads).
// It is parsed by MetadataMiddleware into the request context.
//
// JSON tags mirror the mobile client's payload exactly — do not rename a tag
// without coordinating with the app team.
type RequestMetadata struct {
	// Request tracking
	TraceID   string `json:"trace_id,omitempty"`
	RequestID string `json:"request_id,omitempty"`

	// App / build
	AppVersion string `json:"app_version,omitempty"` // formatted string, e.g. "Version: 1.0.0 + 12"
	Version    string `json:"version,omitempty"`     // semantic version, e.g. "1.0.0"
	Build      string `json:"build,omitempty"`       // build number, e.g. "12"

	// Device
	Platform        string `json:"platform,omitempty"`          // "ios" | "android" | "web"
	ModelName       string `json:"model_name,omitempty"`        // manufacturer / make, e.g. "google", "Apple"
	OSVersion       string `json:"system_version,omitempty"`    // OS version, e.g. "35", "iOS 16.0"
	DeviceID        string `json:"device_uuid,omitempty"`       // unique device identifier
	DeviceName      string `json:"device_name,omitempty"`       // e.g. "Pixel 8"
	DevicePushToken string `json:"device_push_token,omitempty"` // FCM / APNS push token

	// Locale / negotiation
	Language       ClientLanguage `json:"language,omitempty"`        // "vi" | "en" | "vi-VN" | "en-EN"
	AcceptLanguage string         `json:"accept_language,omitempty"` // e.g. "vi"
	Accept         string         `json:"accept,omitempty"`          // e.g. "application/json"
	ContentType    string         `json:"content_type,omitempty"`    // e.g. "application/json"

	// Network — client-reported; the server also derives the real peer IP in
	// the logging middleware. Treat this as advisory, not authoritative.
	IPAddress string `json:"ip_address,omitempty"`

	// Authorization carries the session token ("Bearer <jwt>") the client
	// reports in metadata. Informational only — session resolution uses the
	// real Authorization header / cookie via GexSessionMiddleware, not this
	// field. Should be redacted in request logs.
	Authorization string `json:"authorization,omitempty"`

	// User context
	UserContext UserContext `json:"user_context,omitempty"`

	// Timestamp when the request was created on the client
	Timestamp string `json:"timestamp,omitempty"`
}

// UserContext contains user-specific context information
type UserContext struct {
	Locale   string `json:"locale,omitempty"`   // User's locale (e.g., "en-US", "vi-VN")
	Timezone string `json:"timezone,omitempty"` // User's timezone (e.g., "America/New_York", "Asia/Ho_Chi_Minh")
	Language string `json:"language,omitempty"` // Preferred language (e.g., "en", "vn")
}

// NewRequestMetadata creates a new RequestMetadata instance with default values
func NewRequestMetadata() *RequestMetadata {
	return &RequestMetadata{
		Timestamp:   mtime.Now().String(),
		UserContext: UserContext{},
	}
}
