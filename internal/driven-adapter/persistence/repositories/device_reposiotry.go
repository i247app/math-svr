package repositories

import (
	"context"
	"database/sql"
	"fmt"

	di "math-ai.com/math-ai/internal/core/di/repositories"
	domain "math-ai.com/math-ai/internal/core/domain/device"
	"math-ai.com/math-ai/internal/driven-adapter/persistence/models"
	"math-ai.com/math-ai/internal/driven-adapter/persistence/queries"
	"math-ai.com/math-ai/internal/shared/constant/enum"
	"math-ai.com/math-ai/internal/shared/db"
	mathtime "math-ai.com/math-ai/internal/shared/utils/time"
)

const (
	deviceTableName = "ma_devices"
)

type deviceRepository struct {
	BaseRepository // Embed BaseRepository for common operations
}

func NewDeviceRepository(database db.IDatabase) di.IDeviceRepository {
	return &deviceRepository{
		BaseRepository: NewBaseRepository(database),
	}
}

// scanDevice is a reusable helper method to scan device data from a row
func (r *deviceRepository) scanDevice(scanner Scanner) (*domain.Device, error) {
	var d models.DeviceModel
	err := scanner.Scan(&d.ID, &d.UID, &d.DeviceUuid, &d.DeviceName, &d.DevicePushToken, &d.IsVerified)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("scan error: %v", err)
	}

	return domain.BuildDeviceDomainFromModel(&d), nil
}

// GetDeviceByDeviceUUID retrieves a device by device UUID.
func (r *deviceRepository) GetDeviceByDeviceUUID(ctx context.Context, deviceUUID string) (*domain.Device, error) {
	row := r.db.QueryRow(ctx, nil, queries.DeviceFindByDeviceUUID, deviceUUID, enum.StatusActive)
	return r.scanDevice(row)
}

// GetDeviceByUIDAnDeviceUUID retrieves a device by UID and device UUID.
func (r *deviceRepository) GetDeviceByUIDAnDeviceUUID(ctx context.Context, uid string, deviceUUID string) (*domain.Device, error) {
	row := r.db.QueryRow(ctx, nil, queries.DeviceFindByUIDAndDeviceUUID, uid, deviceUUID, enum.StatusActive)
	return r.scanDevice(row)
}

// CheckTrustedDeviceByUID checks if a device is verified by UID and device UUID.
// Now uses query constants from queries package
func (r *deviceRepository) CheckTrustedDeviceByUID(ctx context.Context, uid string, deviceUUID string) (bool, error) {
	var isVerified bool
	row := r.db.QueryRow(ctx, nil, queries.DeviceCheckTrustedByUID, uid, deviceUUID, enum.StatusActive)
	err := row.Scan(&isVerified)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("scan error: %v", err)
	}

	return isVerified, nil
}

// StoreDevice inserts a new device into the database.
func (r *deviceRepository) StoreDevice(ctx context.Context, tx *sql.Tx, device *domain.Device) error {
	_, err := r.db.Exec(ctx, tx, queries.DeviceInsert,
		device.ID(),
		device.UID(),
		device.DeviceUuid(),
		device.DeviceName(),
		device.DevicePushToken(),
		device.IsVerified(),
		enum.StatusActive,
		mathtime.Now(),
		mathtime.Now(),
	)

	return err
}

// UpdateDevice modifies an existing device in the database.
func (r *deviceRepository) UpdateDevice(ctx context.Context, tx *sql.Tx, device *domain.Device) error {
	_, err := r.db.Exec(ctx, tx, queries.DeviceUpdate,
		device.DeviceName(),
		device.DevicePushToken(),
		device.IsVerified(),
		mathtime.Now(),
		device.UID(),
		device.DeviceUuid(),
		device.Status(),
	)
	if err != nil {
		return fmt.Errorf("failed to create grade: %v", err)
	}

	return nil
}

// MarkVerifiedDeviceByUIDAndDeviceUUID marks a device as verified.
func (r *deviceRepository) MarkVerifiedDeviceByUIDAndDeviceUUID(ctx context.Context, uid string, deviceUUID string) error {
	_, err := r.db.Exec(ctx, nil, queries.DeviceMarkVerified,
		true,
		mathtime.Now(),
		uid,
		deviceUUID,
		enum.StatusActive,
	)

	return err
}

// DeleteDeviceByUID soft deletes devices by UID.
func (r *deviceRepository) DeleteDeviceByUID(ctx context.Context, tx *sql.Tx, uid string) error {
	_, err := r.db.Exec(ctx, tx, queries.DeviceSoftDeleteByUID, mathtime.Now(), mathtime.Now(), uid)
	if err != nil {
		return fmt.Errorf("failed to delete user devices: %v", err)
	}
	return nil
}

// ForceDeleteDeviceByUID permanently deletes devices by UID.
func (r *deviceRepository) ForceDeleteDeviceByUID(ctx context.Context, tx *sql.Tx, uid string) error {
	_, err := r.db.Exec(ctx, tx, queries.DeviceForceDeleteByUID, uid)
	if err != nil {
		return fmt.Errorf("failed to force delete user devices: %v", err)
	}
	return nil
}
