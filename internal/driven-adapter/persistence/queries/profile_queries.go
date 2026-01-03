package queries

// ProfileQueries contains all SQL queries for profile repository
type ProfileQueries struct{}

// Column lists
const (
	ProfileSelectColumns = `p.id, p.uid, u.name, u.email, u.phone, u.avatar_key, u.dob, g.label, s.name, p.status,
		p.create_id, p.create_dt, p.modify_id, p.modify_dt`
)

// Find queries
const (
	ProfileFindByID = `SELECT ` + ProfileSelectColumns + `
		FROM ma_profiles p
		INNER JOIN ma_users u ON p.uid = u.id
		INNER JOIN ma_semesters s ON p.semester_id = s.id
		INNER JOIN ma_grades g ON p.grade_id = g.id
		WHERE p.id = ? AND p.deleted_dt IS NULL AND u.deleted_dt IS NULL`

	ProfileFindByUID = `SELECT ` + ProfileSelectColumns + `
		FROM ma_profiles p
		INNER JOIN ma_users u ON p.uid = u.id
		INNER JOIN ma_semesters s ON p.semester_id = s.id
		INNER JOIN ma_grades g ON p.grade_id = g.id
		WHERE p.uid = ? AND p.deleted_dt IS NULL AND u.deleted_dt IS NULL`
)

// Mutation queries
const (
	ProfileInsert = `INSERT INTO ma_profiles (id, uid, grade_id, semester_id, status, create_dt, modify_dt)
		VALUES (?, ?, ?, ?, ?, ?, ?)`

	ProfileUpdate = `UPDATE ma_profiles
		SET grade_id = COALESCE(?, grade_id),
			semester_id = COALESCE(?, semester_id),
			status = COALESCE(?, status),
			modify_dt = ?
		WHERE uid = ? AND deleted_dt IS NULL`

	ProfileSoftDeleteByUID = `UPDATE ma_profiles
		SET deleted_dt = ?,
			modify_dt = ?
		WHERE uid = ? AND deleted_dt IS NULL`

	ProfileForceDeleteByUID = `DELETE FROM ma_profiles
		WHERE uid = ?`
)
