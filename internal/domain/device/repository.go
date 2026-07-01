package device

import (
	"context"

	"math-ai.com/math-ai/internal/shared/enum"
)

type IRepository interface {
	FindByDeviceId(ctx context.Context, deviceId int64) (*Device, error)
	FindByUserDevice(ctx context.Context, userId int64, deviceUUID string) (*Device, error)
	ListByUserId(ctx context.Context, userId int64) ([]*Device, error)
	Create(ctx context.Context, device *Device) (*Device, error)
	Update(ctx context.Context, device *Device) error
	MarkVerified(ctx context.Context, deviceId int64, isVerified bool) error
	MarkStatusByDeviceId(ctx context.Context, deviceId int64, status enum.DeviceStatusType) error
	SoftDeleteByDeviceId(ctx context.Context, deviceId int64) error
	// ClearPushTokens nulls device_push_token for any device whose token is
	// in tokens. Called after a push send when FCM reports those tokens as
	// dead (unregistered / invalid) so they are not retried.
	ClearPushTokens(ctx context.Context, tokens []string) error
}
