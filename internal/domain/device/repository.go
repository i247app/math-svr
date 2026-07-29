package device

import (
	"context"

	"math-ai.com/math-ai/internal/shared/enum"
)

// ListDevicesParams narrows the per-user device listing. UserID scopes
// ownership (required); IsVerified is a tri-state filter — nil means "no
// filter" (current behavior, backward compatible), non-nil restricts to
// devices whose is_verified matches exactly. Mirrors
// profile.ListProfilesParams.IsDefault.
type ListDevicesParams struct {
	UserID     int64
	IsVerified *bool
}

type IRepository interface {
	FindByDeviceId(ctx context.Context, deviceId int64) (*Device, error)
	FindByUserDevice(ctx context.Context, userId int64, deviceUUID string) (*Device, error)
	ListByUserId(ctx context.Context, params *ListDevicesParams) ([]*Device, error)
	Create(ctx context.Context, device *Device) (*Device, error)
	Update(ctx context.Context, device *Device) error
	MarkVerifiedByUserDevice(ctx context.Context, userId int64, deviceUUID string, isVerified bool) error
	MarkStatusByDeviceId(ctx context.Context, deviceId int64, status enum.DeviceStatusType) error
	SoftDeleteByDeviceId(ctx context.Context, deviceId int64) error
	// ClearPushTokens nulls device_push_token for any device whose token is
	// in tokens. Called after a push send when FCM reports those tokens as
	// dead (unregistered / invalid) so they are not retried.
	ClearPushTokens(ctx context.Context, tokens []string) error
}
