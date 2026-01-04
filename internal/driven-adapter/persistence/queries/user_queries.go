package queries

// UserQueries contains all SQL queries for user repository
type UserQueries struct{}

// Column lists
const (
	UserSelectColumns = `u.id, u.name, u.phone, u.email, u.avatar_key, u.dob,
		u.role_id, u.status,
		u.create_id, u.create_dt, u.modify_id, u.modify_dt,
		r.name as role_name`

	UserSelectColumnsWithPassword = `u.id, u.name, u.phone, u.email, u.avatar_key, u.dob,
		u.role_id, u.status, l.hash_pass,
		u.create_id, u.create_dt, u.modify_id, u.modify_dt,
		r.name as role_name`
)

// Base queries - with role join
const (
	UserBaseSelect = `SELECT ` + UserSelectColumns + `
		FROM ma_users u
		LEFT JOIN roles r ON u.role_id = r.id AND r.deleted_dt IS NULL
		WHERE u.deleted_dt IS NULL`

	UserCountBase = `SELECT COUNT(*)
		FROM ma_users u
		WHERE u.deleted_dt IS NULL`
)

// Find queries
const (
	UserFindByID = `SELECT ` + UserSelectColumns + `
		FROM ma_users u
		LEFT JOIN roles r ON u.role_id = r.id AND r.deleted_dt IS NULL
		WHERE u.id = ? AND u.deleted_dt IS NULL`

	UserFindByEmail = `SELECT ` + UserSelectColumns + `
		FROM ma_users u
		LEFT JOIN roles r ON u.role_id = r.id AND r.deleted_dt IS NULL
		WHERE u.email = ? AND u.deleted_dt IS NULL`

	UserGetByLoginName = `SELECT ` + UserSelectColumnsWithPassword + `
		FROM ma_users u
		JOIN ma_aliases a ON u.id = a.uid
		JOIN ma_logins l ON u.id = l.uid
		LEFT JOIN roles r ON u.role_id = r.id AND r.deleted_dt IS NULL
		WHERE a.aka = ? AND u.deleted_dt IS NULL AND a.deleted_dt IS NULL AND l.deleted_dt IS NULL`
)

// List queries
const (
	UserListQuery = UserBaseSelect

	UserListCountQuery = UserCountBase
)

// User mutation queries
const (
	UserInsert = `INSERT INTO ma_users (id, name, phone, email, avatar_key, dob, role_id, status, create_dt, modify_dt)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	UserUpdate = `UPDATE ma_users
		SET name = COALESCE(?, name),
			phone = COALESCE(?, phone),
			email = COALESCE(?, email),
			dob = COALESCE(?, dob),
			role_id = COALESCE(?, role_id),
			avatar_key = COALESCE(?, avatar_key),
			modify_dt = ?
		WHERE id = ? AND deleted_dt IS NULL`

	UserSoftDelete = `UPDATE ma_users
		SET deleted_dt = ?,
			modify_dt = ?
		WHERE id = ?`

	UserForceDelete = `DELETE FROM ma_users
		WHERE id = ?`
)

// Alias mutation queries
const (
	AliasInsert = `INSERT INTO ma_aliases (id, uid, aka, status, create_dt, modify_dt)
		VALUES (?, ?, ?, ?, ?, ?)`

	AliasSoftDeleteByUID = `UPDATE ma_aliases
		SET deleted_dt = ?,
			modify_dt = ?
		WHERE uid = ? AND deleted_dt IS NULL`

	AliasForceDeleteByUID = `DELETE FROM ma_aliases
		WHERE uid = ?`
)

// BuildListQueryWithSearch adds search condition to list query
// Supports searching by name and email
func (UserQueries) BuildListQueryWithSearch(searchTerm string) (string, string, []interface{}) {
	searchCondition := ` AND (u.name LIKE ? OR u.email LIKE ?)`
	searchPattern := "%" + searchTerm + "%"

	query := UserListQuery + searchCondition
	countQuery := UserListCountQuery + searchCondition

	// Args: search pattern x2 for LIKE
	args := []interface{}{searchPattern, searchPattern}

	return query, countQuery, args
}
