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
	return FromContext(ctx).Platform
}

// GetAppVersion is a convenience function to get the client app version from context
func GetAppVersion(ctx context.Context) string {
	return FromContext(ctx).AppVersion
}

// GetModelName is a convenience function to get the client device
// manufacturer / make (e.g. "google", "Apple") from context.
func GetModelName(ctx context.Context) string {
	return FromContext(ctx).ModelName
}

// GetDeviceModel is retained for backward compatibility; the client now
// reports the make via model_name, so this returns the same value.
func GetDeviceModel(ctx context.Context) string {
	return FromContext(ctx).ModelName
}

// GetOSVersion is a convenience function to get the client OS version
// (system_version) from context.
func GetOSVersion(ctx context.Context) string {
	return FromContext(ctx).OSVersion
}

// GetVersion returns the semantic app version (e.g. "1.0.0").
func GetVersion(ctx context.Context) string {
	return FromContext(ctx).Version
}

// GetBuild returns the app build number (e.g. "12").
func GetBuild(ctx context.Context) string {
	return FromContext(ctx).Build
}

// GetAcceptLanguage returns the client's accept_language value.
func GetAcceptLanguage(ctx context.Context) string {
	return FromContext(ctx).AcceptLanguage
}

// GetAuthorization returns the raw "Bearer <jwt>" the client supplied in
// metadata. Handlers normally rely on the resolved session instead.
func GetAuthorization(ctx context.Context) string {
	return FromContext(ctx).Authorization
}

// GetDeviceID is a convenience function to get the client device ID from context
func GetDeviceID(ctx context.Context) string {
	return FromContext(ctx).DeviceID
}

// GetDeviceName is a convenience function to get the client device name from context
func GetDeviceName(ctx context.Context) string {
	return FromContext(ctx).DeviceName
}

// GetDevicePushToken is a convenience function to get the client device push token from context
func GetDevicePushToken(ctx context.Context) string {
	return FromContext(ctx).DevicePushToken
}

// GetIPAddress is a convenience function to get the client IP address from context
func GetIPAddress(ctx context.Context) string {
	return FromContext(ctx).IPAddress
}

// GetClientLanguage is a convenience function to get the client language from context
func GetClientLanguage(ctx context.Context) ClientLanguage {
	return FromContext(ctx).Language
}
