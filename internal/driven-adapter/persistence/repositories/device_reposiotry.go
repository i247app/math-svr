package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"math-ai.com/math-ai/internal/core/di/repositories"
	domain "math-ai.com/math-ai/internal/core/domain/device"
	"math-ai.com/math-ai/internal/driven-adapter/persistence/models"
	"math-ai.com/math-ai/internal/shared/constant/enum"
	"math-ai.com/math-ai/internal/shared/db"
)

type deviceRepository struct {
	db db.IDatabase
}

func NewDeviceRepository(db db.IDatabase) repositories.IDeviceRepository {
	return &deviceRepository{
		db: db,
	}
}

func (r *deviceRepository) GetDeviceByDeviceUUID(ctx context.Context, deviceUUID string) (*domain.Device, error) {
	query := `
		SELECT id, uid, device_uuid, device_name, device_push_token, is_verified
		FROM devices
		WHERE device_uuid = ? AND status = ?
	`

	result := r.db.QueryRow(ctx, nil, query, deviceUUID, enum.StatusActive)
	var d models.DeviceModel
	err := result.Scan(&d.ID, &d.UID, &d.DeviceUuid, &d.DeviceName, &d.DevicePushToken, &d.IsVerified)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	device := domain.BuildDeviceDomainFromModel(&d)

	return device, nil
}

func (r *deviceRepository) GetDeviceByUIDAnDeviceUUID(ctx context.Context, uid string, deviceUUID string) (*domain.Device, error) {
	query := `
		SELECT id, uid, device_uuid, device_name, device_push_token, is_verified
		FROM devices
		WHERE uid = ? AND device_uuid = ? AND status = ?
	`

	result := r.db.QueryRow(ctx, nil, query, uid, deviceUUID, enum.StatusActive)
	var d models.DeviceModel
	err := result.Scan(&d.ID, &d.UID, &d.DeviceUuid, &d.DeviceName, &d.DevicePushToken, &d.IsVerified)
	if err != nil {
		log.Println("Error scanning device:", err)
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	device := domain.BuildDeviceDomainFromModel(&d)

	return device, nil
}

func (r *deviceRepository) CheckTrustedDeviceByUID(ctx context.Context, uid string, deviceUUID string) (bool, error) {
	query := `
		SELECT is_verified
		FROM devices
		WHERE uid = ? AND device_uuid = ? AND status = ?
	`

	result := r.db.QueryRow(ctx, nil, query, uid, deviceUUID, enum.StatusActive)
	var isVerified bool
	err := result.Scan(&isVerified)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	return isVerified, nil
}

func (r *deviceRepository) StoreDevice(ctx context.Context, tx *sql.Tx, device *domain.Device) error {
	query := `
		INSERT INTO devices (id, uid, device_uuid, device_name, device_push_token, is_verified, status)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.Exec(ctx, tx, query,
		device.ID(),
		device.UID(),
		device.DeviceUuid(),
		device.DeviceName(),
		device.DevicePushToken(),
		device.IsVerified(),
		enum.StatusActive,
	)

	return err
}

func (r *deviceRepository) UpdateDevice(ctx context.Context, device *domain.Device) error {
	query := `
		UPDATE devices
		SET uid = COALESCE(?, uid),
			device_uuid = COALESCE(?, device_uuid),
			device_name = COALESCE(?, device_name),
			device_push_token = COALESCE(?, device_push_token),
			is_verified = COALESCE(?, is_verified)
		WHERE id = ? AND status = ?
	`

	_, err := r.db.Exec(ctx, nil, query,
		device.UID(),
		device.DeviceUuid(),
		device.DeviceName(),
		device.DevicePushToken(),
		device.IsVerified(),
		device.ID(),
		enum.StatusActive,
	)

	return err
}

func (r *deviceRepository) MarkVerifiedDeviceByUIDAndDeviceUUID(ctx context.Context, uid string, deviceUUID string) error {
	query := `
		UPDATE devices
		SET is_verified = ?
		WHERE uid = ? AND device_uuid = ? AND status = ?
	`

	_, err := r.db.Exec(ctx, nil, query,
		true,
		uid,
		deviceUUID,
		enum.StatusActive,
	)

	return err
}

func (r *deviceRepository) DeleteDeviceByUID(ctx context.Context, tx *sql.Tx, uid string) error {
	query := `
		UPDATE devices
		SET deleted_dt = ?
		WHERE uid = ? AND deleted_dt IS NULL
	`
	_, err := r.db.Exec(ctx, tx, query, time.Now().UTC(), uid)
	if err != nil {
		return fmt.Errorf("failed to delete user logins: %v", err)
	}
	return nil
}

func (r *deviceRepository) ForceDeleteDeviceByUID(ctx context.Context, tx *sql.Tx, uid string) error {
	query := `
		DELETE FROM devices
		WHERE uid = ?
	`
	_, err := r.db.Exec(ctx, tx, query, uid)
	if err != nil {
		return fmt.Errorf("failed to force delete user devices: %v", err)
	}
	return nil
}
