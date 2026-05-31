package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"math-ai.com/math-ai/internal/domain/device"
	mtime "math-ai.com/math-ai/internal/domain/shared/time"
	"math-ai.com/math-ai/internal/infrastructure/database"
	"math-ai.com/math-ai/internal/infrastructure/persistence/mysql/models"
	"math-ai.com/math-ai/internal/shared/enum"
)

const (
	deviceTable = "ma_devices"

	deviceColumns = `d.id, d.device_id, d.user_id, d.device_uuid, d.device_name,
		d.device_push_token, d.is_verified, d.note, d.device_status, d.status,
		d.create_id, d.create_dt, d.modify_id, d.modify_dt`

	deviceActiveWhere = `d.status IN (?) AND d.deleted_dt IS NULL`
)

func deviceActiveArgs() []any {
	return []any{enum.StatusActive}
}

type DeviceRepository struct {
	db database.Executor
}

func NewDeviceRepository(db database.Executor) device.IRepository {
	return &DeviceRepository{db: db}
}

func scanDevice(s database.RowScanner) (*models.DeviceModel, error) {
	var m models.DeviceModel
	if err := s.Scan(&m.Id, &m.DeviceId, &m.UserId, &m.DeviceUUID, &m.DeviceName,
		&m.DevicePushToken, &m.IsVerified, &m.Note, &m.DeviceStatus, &m.Status,
		&m.CreateId, &m.CreateDt, &m.ModifyId, &m.ModifyDt); err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *DeviceRepository) findOneBy(ctx context.Context, where string, args ...any) (*device.Device, error) {
	fullArgs := append(deviceActiveArgs(), args...)
	query := `SELECT ` + deviceColumns + ` FROM ` + deviceTable + ` d WHERE ` +
		deviceActiveWhere + ` AND (` + where + `)`

	m, err := scanDevice(r.db.QueryRow(ctx, query, fullArgs...))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("device repo find (%s): %w", where, err)
	}
	return ModelToDomainDevice(m), nil
}

func (r *DeviceRepository) findBareById(ctx context.Context, id int64) (*device.Device, error) {
	args := append(deviceActiveArgs(), id)
	query := `SELECT ` + deviceColumns + ` FROM ` + deviceTable + ` d WHERE ` +
		deviceActiveWhere + ` AND d.id = ?`

	m, err := scanDevice(r.db.QueryRow(ctx, query, args...))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("device repo find bare by id: %w", err)
	}
	return ModelToDomainDevice(m), nil
}

func (r *DeviceRepository) FindByDeviceId(ctx context.Context, deviceId int64) (*device.Device, error) {
	return r.findOneBy(ctx, "d.device_id = ?", deviceId)
}

// FindByUserDevice looks up the registration for the (user, device_uuid) pair.
// device_uuid is the client-supplied stable identifier (installation id, IDFV,
// etc.); device_id is our own UUID for the row.
func (r *DeviceRepository) FindByUserDevice(ctx context.Context, userId int64, deviceUUID string) (*device.Device, error) {
	return r.findOneBy(ctx, "d.user_id = ? AND d.device_uuid = ?", userId, deviceUUID)
}

func (r *DeviceRepository) ListByUserId(ctx context.Context, userId int64) ([]*device.Device, error) {
	args := append(deviceActiveArgs(), userId)
	query := `SELECT ` + deviceColumns + ` FROM ` + deviceTable + ` d WHERE ` +
		deviceActiveWhere + ` AND d.user_id = ? ORDER BY d.id DESC`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("device repo list by user id: %w", err)
	}
	defer rows.Close()

	var devices []*device.Device
	for rows.Next() {
		m, err := scanDevice(rows)
		if err != nil {
			return nil, fmt.Errorf("device repo scan row: %w", err)
		}
		devices = append(devices, ModelToDomainDevice(m))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("device repo rows iteration: %w", err)
	}
	return devices, nil
}

func (r *DeviceRepository) Create(ctx context.Context, d *device.Device) (*device.Device, error) {
	query := `
		INSERT INTO ` + deviceTable + `
			(device_id, user_id, device_uuid, device_name, device_push_token,
			 is_verified, note, device_status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := r.db.Exec(ctx, query,
		d.DeviceId(), d.UserId(), d.DeviceUUID(), d.DeviceName(), d.DevicePushToken(),
		d.IsVerified(), d.Note(), d.DeviceStatus())
	if err != nil {
		return nil, fmt.Errorf("device repo create: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("device repo last insert id: %w", err)
	}
	return r.findBareById(ctx, id)
}

// Update mutates the row in place. COALESCE keeps nil fields untouched so
// callers can patch a single attribute (push token, name) without re-supplying
// everything.
func (r *DeviceRepository) Update(ctx context.Context, d *device.Device) error {
	query := `
		UPDATE ` + deviceTable + `
		SET device_name       = COALESCE(?, device_name),
			device_push_token = COALESCE(?, device_push_token),
			note              = COALESCE(?, note)
		WHERE device_id = ?
	`

	var nameArg any
	if d.DeviceName() != "" {
		nameArg = d.DeviceName()
	}

	if _, err := r.db.Exec(ctx, query,
		nameArg, d.DevicePushToken(), d.Note(), d.DeviceId()); err != nil {
		return fmt.Errorf("device repo update: %w", err)
	}
	return nil
}

// MarkVerified flips is_verified. Called by the (future) 2FA flow on success,
// or by revoke to un-trust the device.
func (r *DeviceRepository) MarkVerified(ctx context.Context, deviceId int64, isVerified bool) error {
	query := `
		UPDATE ` + deviceTable + `
		SET is_verified = ?,
			modify_dt   = ?
		WHERE device_id = ?
	`
	if _, err := r.db.Exec(ctx, query, isVerified, mtime.Now().Time, deviceId); err != nil {
		return fmt.Errorf("device repo mark verified: %w", err)
	}
	return nil
}

func (r *DeviceRepository) MarkStatusByDeviceId(ctx context.Context, deviceId int64, st enum.DeviceStatusType) error {
	query := `
		UPDATE ` + deviceTable + `
		SET device_status = ?,
			modify_dt     = ?
		WHERE device_id = ?
	`
	if _, err := r.db.Exec(ctx, query, st, mtime.Now().Time, deviceId); err != nil {
		return fmt.Errorf("device repo mark status: %w", err)
	}
	return nil
}

func (r *DeviceRepository) SoftDeleteByDeviceId(ctx context.Context, deviceId int64) error {
	query := `
		UPDATE ` + deviceTable + `
		SET device_status = ?,
			status        = ?,
			deleted_dt    = ?
		WHERE device_id = ?
	`
	if _, err := r.db.Exec(ctx, query,
		enum.DeviceStatusTypeDeleted, enum.StatusInactive, mtime.Now().Time, deviceId); err != nil {
		return fmt.Errorf("device repo soft delete: %w", err)
	}
	return nil
}

func ModelToDomainDevice(m *models.DeviceModel) *device.Device {
	d := device.NewDevice()
	d.SetId(m.Id)
	d.SetDeviceId(m.DeviceId)
	d.SetUserId(m.UserId)
	d.SetDeviceUUID(m.DeviceUUID)
	d.SetDeviceName(m.DeviceName)
	d.SetDevicePushToken(m.DevicePushToken)
	d.SetIsVerified(m.IsVerified)
	d.SetNote(m.Note)
	d.SetDeviceStatus(m.DeviceStatus)
	d.SetStatus(m.Status)
	d.SetCreateId(m.CreateId)
	d.SetCreateDt(mtime.MathTime{Time: m.CreateDt})
	d.SetModifyId(m.ModifyId)
	d.SetModifyDt(mtime.MathTime{Time: m.ModifyDt})
	return d
}
