package queries

// DeviceQueries contains all SQL queries for device repository
type DeviceQueries struct{}

// Column lists
const (
	DeviceSelectColumns = `id, uid, device_uuid, device_name, device_push_token, is_verified`
)

// Find queries
const (
	DeviceFindByDeviceUUID = `SELECT ` + DeviceSelectColumns + `
		FROM ma_devices
		WHERE device_uuid = ? AND status = ?`

	DeviceFindByUIDAndDeviceUUID = `SELECT ` + DeviceSelectColumns + `
		FROM ma_devices
		WHERE uid = ? AND device_uuid = ? AND status = ?`

	DeviceCheckTrustedByUID = `SELECT is_verified
		FROM ma_devices
		WHERE uid = ? AND device_uuid = ? AND status = ?`
)

// Mutation queries
const (
	DeviceInsert = `INSERT INTO ma_devices (id, uid, device_uuid, device_name, device_push_token, is_verified, status, create_dt, modify_dt)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	DeviceUpdate = `UPDATE ma_devices
		SET device_name = COALESCE(?, device_name),
			device_push_token = COALESCE(?, device_push_token),
			is_verified = COALESCE(?, is_verified),
			modify_dt = ?
		WHERE uid = ? AND device_uuid = ? AND status = ?`

	DeviceMarkVerified = `UPDATE ma_devices
		SET is_verified = ?,
			modify_dt = ?
		WHERE uid = ? AND device_uuid = ? AND status = ?`

	DeviceSoftDeleteByUID = `UPDATE ma_devices
		SET deleted_dt = ?,
			modify_dt = ?
		WHERE uid = ? AND deleted_dt IS NULL`

	DeviceForceDeleteByUID = `DELETE FROM ma_devices
		WHERE uid = ?`
)
