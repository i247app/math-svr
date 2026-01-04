package queries

// AuthQueries contains all SQL queries for auth repository
type AuthQueries struct{}

// Column lists
const (
	LoginLogSelectColumns = `id, uid, ip_address, device_uuid, token`
)

// Find queries
const (
	LoginLogFindByUIDAndDeviceUUID = `SELECT ` + LoginLogSelectColumns + `
		FROM ma_login_logs
		WHERE uid = ? AND device_uuid = ?`
)

// Login mutation queries
const (
	LoginInsert = `INSERT INTO ma_logins (id, uid, hash_pass, status, create_dt, modify_dt)
		VALUES (?, ?, ?, ?, ?, ?)`

	LoginSoftDeleteByUID = `UPDATE ma_logins
		SET deleted_dt = ?,
			modify_dt = ?
		WHERE uid = ? AND deleted_dt IS NULL`

	LoginForceDeleteByUID = `DELETE FROM ma_logins
		WHERE uid = ?`
)

// Login log mutation queries
const (
	LoginLogInsert = `INSERT INTO ma_login_logs (id, uid, ip_address, device_uuid, token, status, create_dt, modify_dt)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	LoginLogUpdate = `UPDATE ma_login_logs
		SET token = COALESCE(?, token),
			status = COALESCE(?, status),
			modify_dt = ?
		WHERE uid = ? AND device_uuid = ?`

	LoginLogSoftDeleteByUID = `UPDATE ma_login_logs
		SET deleted_dt = ?,
			modify_dt = ?
		WHERE uid = ? AND deleted_dt IS NULL`

	LoginLogForceDeleteByUID = `DELETE FROM ma_login_logs
		WHERE uid = ?`
)
