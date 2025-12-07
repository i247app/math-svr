package di

import (
	"context"
	"database/sql"

	domain "math-ai.com/math-ai/internal/core/domain/device"
)

type IDeviceRepository interface {
	GetDeviceByDeviceUUID(ctx context.Context, deviceUUID string) (*domain.Device, error)
	GetDeviceByUIDAnDeviceUUID(ctx context.Context, uid string, deviceUUID string) (*domain.Device, error)
	CheckTrustedDeviceByUID(ctx context.Context, uid string, deviceUUID string) (bool, error)
	StoreDevice(ctx context.Context, tx *sql.Tx, device *domain.Device) error
	UpdateDevice(ctx context.Context, device *domain.Device) error
	MarkVerifiedDeviceByUIDAndDeviceUUID(ctx context.Context, uid string, deviceUUID string) error
	DeleteDeviceByUID(ctx context.Context, tx *sql.Tx, uid string) error
	ForceDeleteDeviceByUID(ctx context.Context, tx *sql.Tx, uid string) error
}
