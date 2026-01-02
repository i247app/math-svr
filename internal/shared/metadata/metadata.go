package metadata

import (
	"time"
)

// RequestMetadata contains metadata information sent by clients with every request
type RequestMetadata struct {
	// Request tracking
	TraceID   string `json:"trace_id,omitempty"`
	RequestID string `json:"request_id,omitempty"`

	// Client information
	ClientInfo ClientInfo `json:"client_info,omitempty"`

	// User context
	UserContext UserContext `json:"user_context,omitempty"`

	// Timestamp when request was received
	Timestamp time.Time `json:"timestamp,omitempty"`
}

// ClientInfo contains information about the client making the request
type ClientInfo struct {
	AppVersion  string `json:"app_version,omitempty"`  // Client app version (e.g., "1.2.3")
	Platform    string `json:"platform,omitempty"`     // Platform (e.g., "ios", "android", "web")
	DeviceModel string `json:"device_model,omitempty"` // Device model (e.g., "iPhone 14", "Pixel 7")
	OSVersion   string `json:"os_version,omitempty"`   // OS version (e.g., "iOS 16.0", "Android 13")
	DeviceID    string `json:"device_id,omitempty"`    // Unique device identifier
	DeviceName  string `json:"device_name,omitempty"`  // Device name (e.g., "John's iPhone")
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
		Timestamp:   time.Now(),
		ClientInfo:  ClientInfo{},
		UserContext: UserContext{},
	}
}

// IsEmpty checks if the metadata is empty or not populated
func (m *RequestMetadata) IsEmpty() bool {
	return m.TraceID == "" && m.RequestID == "" && m.ClientInfo.Platform == ""
}
