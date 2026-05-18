package device

import (
	"context"

	"github.com/google/uuid"
	"math-ai.com/math-ai/internal/shared/enum"
)

type IRepository interface {
	FindByDeviceId(ctx context.Context, deviceId uuid.UUID) (*Device, error)
	FindByUserDevice(ctx context.Context, userId uuid.UUID, deviceUUID string) (*Device, error)
	ListByUserId(ctx context.Context, userId uuid.UUID) ([]*Device, error)
	Create(ctx context.Context, device *Device) (*Device, error)
	Update(ctx context.Context, device *Device) error
	MarkVerified(ctx context.Context, deviceId uuid.UUID, isVerified bool) error
	MarkStatusByDeviceId(ctx context.Context, deviceId uuid.UUID, status enum.DeviceStatusType) error
	SoftDeleteByDeviceId(ctx context.Context, deviceId uuid.UUID) error
}
