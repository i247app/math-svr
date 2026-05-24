package device

import (
	"context"

	"math-ai.com/math-ai/internal/shared/enum"
)

type IRepository interface {
	FindByDeviceId(ctx context.Context, deviceId string) (*Device, error)
	FindByUserDevice(ctx context.Context, userId string, deviceUUID string) (*Device, error)
	ListByUserId(ctx context.Context, userId string) ([]*Device, error)
	Create(ctx context.Context, device *Device) (*Device, error)
	Update(ctx context.Context, device *Device) error
	MarkVerified(ctx context.Context, deviceId string, isVerified bool) error
	MarkStatusByDeviceId(ctx context.Context, deviceId string, status enum.DeviceStatusType) error
	SoftDeleteByDeviceId(ctx context.Context, deviceId string) error
}
