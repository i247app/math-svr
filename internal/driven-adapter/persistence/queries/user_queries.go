package queries

// UserQueries contains all SQL queries for user repository
// Organizing queries here makes them easier to:
// - Review and optimize
// - Test independently
// - Version control
// - Document
type UserQueries struct{}

// Column lists
const (
	UserSelectColumns = `u.id, u.name, u.phone, u.email, u.avatar_key, u.dob,
		u.role_id, u.status,
		u.create_id, u.create_dt, u.modify_id, u.modify_dt,
		r.name as role_name`
)

// Base queries - reusable fragments
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
	UserFindByID = UserBaseSelect + ` AND u.id = ?`

	UserFindByEmail = UserBaseSelect + ` AND u.email = ?`

	UserGetByLoginName = `SELECT ` + UserSelectColumns + `, l.hash_pass
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

// Mutation queries
const (
	UserInsert = `INSERT INTO ma_users (id, name, phone, email, avatar_key, dob, role_id, status, create_dt, modify_dt)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	UserDelete = `UPDATE ma_users
		SET deleted_dt = ?, modify_dt = ?
		WHERE id = ?`

	UserForceDelete = `DELETE FROM ma_users WHERE id = ?`
)

// Alias queries
const (
	UserAliasInsert = `INSERT INTO ma_aliases (id, uid, aka, status, create_dt, modify_dt)
		VALUES (?, ?, ?, ?, ?, ?)`

	UserAliasDelete = `UPDATE ma_aliases
		SET deleted_dt = ?, modify_dt = ?
		WHERE uid = ? AND deleted_dt IS NULL`

	UserAliasForceDelete = `DELETE FROM ma_aliases WHERE uid = ?`
)

// Helper methods to build dynamic queries

// BuildListQueryWithSearch adds search condition to list query
func (UserQueries) BuildListQueryWithSearch(searchTerm string) (string, string, []interface{}) {
	searchCondition := ` AND (u.name LIKE ? OR u.email LIKE ?)`
	searchPattern := "%" + searchTerm + "%"
	args := []interface{}{searchPattern, searchPattern}

	query := UserListQuery + searchCondition
	countQuery := UserListCountQuery + searchCondition

	return query, countQuery, args
}
