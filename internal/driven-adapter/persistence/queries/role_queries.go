package queries

// RoleQueries contains all SQL queries for role repository
type RoleQueries struct{}

// Column lists
const (
	RoleSelectColumns = `id, name, code, description, parent_role_id, is_system_role,
		status, display_order, create_id, create_dt, modify_id, modify_dt`
)

// Base queries
const (
	RoleBaseSelect = `SELECT ` + RoleSelectColumns + `
		FROM roles
		WHERE deleted_dt IS NULL`

	RoleBaseCount = `SELECT COUNT(*)
		FROM roles
		WHERE deleted_dt IS NULL`
)

// Find queries
const (
	RoleFindByID = `SELECT ` + RoleSelectColumns + `
		FROM roles
		WHERE id = ? AND deleted_dt IS NULL`

	RoleFindByCode = `SELECT ` + RoleSelectColumns + `
		FROM roles
		WHERE code = ? AND deleted_dt IS NULL`
)

// Mutation queries
const (
	RoleInsert = `INSERT INTO roles (id, name, code, description, parent_role_id, is_system_role, status, display_order, create_dt, modify_dt)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	RoleUpdate = `UPDATE roles
		SET name = COALESCE(?, name),
			code = COALESCE(?, code),
			description = COALESCE(?, description),
			parent_role_id = COALESCE(?, parent_role_id),
			status = COALESCE(?, status),
			display_order = COALESCE(?, display_order),
			modify_dt = ?
		WHERE id = ? AND deleted_dt IS NULL`

	RoleSoftDelete = `UPDATE roles
		SET deleted_dt = ?,
			modify_dt = ?
		WHERE id = ? AND deleted_dt IS NULL AND is_system_role = FALSE`

	RoleForceDelete = `DELETE FROM roles
		WHERE id = ? AND is_system_role = FALSE`
)
