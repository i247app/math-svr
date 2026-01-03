package queries

// SemesterQueries contains all SQL queries for semester repository
type SemesterQueries struct{}

// Column lists
const (
	SemesterSelectColumns = `s.id,
		COALESCE(st.name, s.name) AS name,
		COALESCE(st.description, s.description) AS description,
		s.image_key,
		s.status,
		s.display_order,
		s.create_id,
		s.create_dt,
		s.modify_id,
		s.modify_dt`

	SemesterSelectColumnsNoTranslation = `id, name, description, image_key, status, display_order,
		create_id, create_dt, modify_id, modify_dt`
)

// Base queries - with translation support
const (
	SemesterBaseSelect = `SELECT ` + SemesterSelectColumns + `
		FROM ma_semesters s
		LEFT JOIN ma_semester_translations st ON s.id = st.semester_id AND st.language = ?
		WHERE s.deleted_dt IS NULL`

	SemesterCountBase = `SELECT COUNT(*)
		FROM ma_semesters s
		LEFT JOIN ma_semester_translations st ON s.id = st.semester_id AND st.language = ?
		WHERE s.deleted_dt IS NULL`
)

// Find queries - no translation (direct lookup)
const (
	SemesterFindByID = `SELECT ` + SemesterSelectColumnsNoTranslation + `
		FROM ma_semesters
		WHERE id = ? AND deleted_dt IS NULL`

	SemesterFindByName = `SELECT ` + SemesterSelectColumnsNoTranslation + `
		FROM ma_semesters
		WHERE name = ? AND deleted_dt IS NULL`
)

// Find queries with translation
const (
	SemesterFindByIDWithTranslation = `SELECT ` + SemesterSelectColumns + `
		FROM ma_semesters s
		LEFT JOIN ma_semester_translations st ON s.id = st.semester_id AND st.language = ?
		WHERE s.id = ? AND s.deleted_dt IS NULL`

	SemesterFindByNameWithTranslation = `SELECT ` + SemesterSelectColumns + `
		FROM ma_semesters s
		LEFT JOIN ma_semester_translations st ON s.id = st.semester_id AND st.language = ?
		WHERE s.name = ? AND s.deleted_dt IS NULL`
)

// List queries
const (
	SemesterListQuery = SemesterBaseSelect

	SemesterListCountQuery = SemesterCountBase
)

// Mutation queries
const (
	SemesterInsert = `INSERT INTO ma_semesters (id, name, description, image_key, status, display_order, create_dt, modify_dt)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	SemesterTranslationInsert = `INSERT INTO ma_semester_translations (id, semester_id, language, name, description, note, create_dt, modify_dt)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	SemesterUpdate = `UPDATE ma_semesters
		SET name = COALESCE(?, name),
			description = COALESCE(?, description),
			image_key = COALESCE(?, image_key),
			semester_status = COALESCE(?, semester_status),
			status = COALESCE(?, status),
			display_order = COALESCE(?, display_order),
			modify_dt = ?
		WHERE id = ? AND deleted_dt IS NULL`

	SemesterTranslationUpdate = `UPDATE ma_semester_translations
		SET name = COALESCE(?, name),
			description = COALESCE(?, description),
			note = COALESCE(?, note),
			st_status = COALESCE(?, st_status),
			status = COALESCE(?, status),
			modify_dt = ?
		WHERE semester_id = ? AND language = ? AND deleted_dt IS NULL`

	SemesterDelete = `UPDATE ma_semesters
		SET deleted_dt = ?, modify_dt = ?
		WHERE id = ? AND deleted_dt IS NULL`

	SemesterTranslationDeleteBySemesterID = `UPDATE ma_semester_translations
		SET deleted_dt = ?, modify_dt = ?
		WHERE semester_id = ? AND deleted_dt IS NULL`

	SemesterForceDelete = `DELETE FROM ma_semesters WHERE id = ?`

	SemesterTranslationForceDeleteBySemesterID = `DELETE FROM ma_semester_translations WHERE semester_id = ?`
)

// BuildListQueryWithSearch adds search condition to list query
// Supports searching by name and description in translated or default values
func (SemesterQueries) BuildListQueryWithSearch(language, searchTerm string) (string, string, []interface{}) {
	searchCondition := ` AND (COALESCE(st.name, s.name) LIKE ? OR COALESCE(st.description, s.description) LIKE ?)`
	searchPattern := "%" + searchTerm + "%"

	query := SemesterListQuery + searchCondition
	countQuery := SemesterListCountQuery + searchCondition

	// Args: language for JOIN, search pattern x2 for LIKE
	args := []interface{}{language, searchPattern, searchPattern}

	return query, countQuery, args
}
