package queries

// RolePermissionQueries contains all SQL queries for role_permission repository
type RolePermissionQueries struct{}

// Column lists
const (
	RolePermissionSelectColumns = `id, role_id, permission_id, create_id, create_dt, modify_id, modify_dt`
)

// Find queries
const (
	RolePermissionFindByRoleID = `SELECT ` + RolePermissionSelectColumns + `
		FROM role_permissions
		WHERE role_id = ? AND deleted_dt IS NULL`

	RolePermissionFindByPermissionID = `SELECT ` + RolePermissionSelectColumns + `
		FROM role_permissions
		WHERE permission_id = ? AND deleted_dt IS NULL`
)

// Mutation queries
const (
	RolePermissionInsert = `INSERT INTO role_permissions (id, role_id, permission_id)
		VALUES (?, ?, ?)`

	RolePermissionRevoke = `DELETE FROM role_permissions
		WHERE role_id = ? AND permission_id = ?`

	RolePermissionRevokeAll = `DELETE FROM role_permissions
		WHERE role_id = ?`
)
