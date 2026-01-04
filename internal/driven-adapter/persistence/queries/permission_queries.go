package queries

// PermissionQueries contains all SQL queries for permission repository
type PermissionQueries struct{}

// Column lists
const (
	PermissionSelectColumns = `id, name, description, http_method, endpoint_path, resource, action,
		status, create_id, create_dt, modify_id, modify_dt`
)

// Base queries
const (
	PermissionBaseSelect = `SELECT ` + PermissionSelectColumns + `
		FROM permissions
		WHERE deleted_dt IS NULL`

	PermissionBaseCount = `SELECT COUNT(*)
		FROM permissions
		WHERE deleted_dt IS NULL`
)

// Find queries
const (
	PermissionFindByID = `SELECT ` + PermissionSelectColumns + `
		FROM permissions
		WHERE id = ? AND deleted_dt IS NULL`

	PermissionFindByName = `SELECT ` + PermissionSelectColumns + `
		FROM permissions
		WHERE name = ? AND deleted_dt IS NULL`

	PermissionFindByEndpoint = `SELECT ` + PermissionSelectColumns + `
		FROM permissions
		WHERE http_method = ? AND endpoint_path = ? AND deleted_dt IS NULL`

	PermissionFindByResource = `SELECT ` + PermissionSelectColumns + `
		FROM permissions
		WHERE resource = ? AND deleted_dt IS NULL
		ORDER BY name ASC`

	PermissionFindByRoleID = `SELECT DISTINCT p.id, p.name, p.description, p.http_method, p.endpoint_path, p.resource, p.action,
		p.status, p.create_id, p.create_dt, p.modify_id, p.modify_dt
		FROM permissions p
		INNER JOIN role_permissions rp ON p.id = rp.permission_id
		WHERE rp.role_id = ? AND p.deleted_dt IS NULL AND rp.deleted_dt IS NULL
		ORDER BY p.resource ASC, p.name ASC`
)

// Mutation queries
const (
	PermissionInsert = `INSERT INTO permissions (id, name, description, http_method, endpoint_path, resource, action, status, create_dt, modify_dt)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	PermissionUpdate = `UPDATE permissions
		SET name = COALESCE(?, name),
			description = COALESCE(?, description),
			http_method = COALESCE(?, http_method),
			endpoint_path = COALESCE(?, endpoint_path),
			resource = COALESCE(?, resource),
			action = COALESCE(?, action),
			status = COALESCE(?, status),
			modify_dt = ?
		WHERE id = ? AND deleted_dt IS NULL`

	PermissionSoftDelete = `UPDATE permissions
		SET deleted_dt = ?,
			modify_dt = ?
		WHERE id = ? AND deleted_dt IS NULL`

	PermissionForceDelete = `DELETE FROM permissions
		WHERE id = ?`
)
