package metadata

import (
	"context"
)

type contextKey string

const (
	// metadataContextKey is the key used to store RequestMetadata in context
	metadataContextKey contextKey = "request_metadata"
)

// WithMetadata adds RequestMetadata to the context
func WithMetadata(ctx context.Context, metadata *RequestMetadata) context.Context {
	return context.WithValue(ctx, metadataContextKey, metadata)
}

// FromContext extracts RequestMetadata from the context
// Returns a new empty RequestMetadata if not found in context
func FromContext(ctx context.Context) *RequestMetadata {
	if metadata, ok := ctx.Value(metadataContextKey).(*RequestMetadata); ok {
		return metadata
	}
	// Return empty metadata if not found (safe default)
	return NewRequestMetadata()
}

// GetTraceID is a convenience function to get the trace ID from context
func GetTraceID(ctx context.Context) string {
	return FromContext(ctx).TraceID
}

// GetRequestID is a convenience function to get the request ID from context
func GetRequestID(ctx context.Context) string {
	return FromContext(ctx).RequestID
}

// GetLocale is a convenience function to get the user's locale from context
func GetLocale(ctx context.Context) string {
	return FromContext(ctx).UserContext.Locale
}

// GetTimezone is a convenience function to get the user's timezone from context
func GetTimezone(ctx context.Context) string {
	return FromContext(ctx).UserContext.Timezone
}

// GetLanguage is a convenience function to get the user's language from context
func GetLanguage(ctx context.Context) string {
	return FromContext(ctx).UserContext.Language
}

// GetPlatform is a convenience function to get the client platform from context
func GetPlatform(ctx context.Context) string {
	return FromContext(ctx).ClientInfo.Platform
}

// GetAppVersion is a convenience function to get the client app version from context
func GetAppVersion(ctx context.Context) string {
	return FromContext(ctx).ClientInfo.AppVersion
}

// GetDeviceModel is a convenience function to get the client device model from context
func GetDeviceModel(ctx context.Context) string {
	return FromContext(ctx).ClientInfo.DeviceModel
}

// GetDeviceID is a convenience function to get the client device ID from context
func GetDeviceID(ctx context.Context) string {
	return FromContext(ctx).ClientInfo.DeviceID
}

// GetDeviceName is a convenience function to get the client device name from context
func GetDeviceName(ctx context.Context) string {
	return FromContext(ctx).ClientInfo.DeviceName
}

// GetDevicePushToken is a convenience function to get the client device push token from context
func GetDevicePushToken(ctx context.Context) string {
	return FromContext(ctx).ClientInfo.DevicePushToken
}

// GetIPAddress is a convenience function to get the client IP address from context
func GetIPAddress(ctx context.Context) string {
	return FromContext(ctx).ClientInfo.IPAddress
}
