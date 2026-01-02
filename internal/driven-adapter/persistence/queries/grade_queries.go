package queries

// GradeQueries contains all SQL queries for grade repository
type GradeQueries struct{}

// Column lists
const (
	GradeSelectColumns = `g.id,
		COALESCE(gt.label, g.label) AS label,
		COALESCE(gt.description, g.discription) AS description,
		g.image_key,
		g.status,
		g.display_order,
		g.create_id,
		g.create_dt,
		g.modify_id,
		g.modify_dt`

	GradeSelectColumnsNoTranslation = `id, label, discription, image_key, status, display_order,
		create_id, create_dt, modify_id, modify_dt`
)

// List queries - no translation
const (
	GradeListQueryNoTranslation = `SELECT ` + GradeSelectColumnsNoTranslation + `
		FROM ma_grades
		WHERE deleted_dt IS NULL`

	GradeListCountQueryNoTranslation = `SELECT COUNT(*)
		FROM ma_grades
		WHERE deleted_dt IS NULL`
)

// List queries with translation
const (
	GradeListQuery = `SELECT ` + GradeSelectColumns + `
		FROM ma_grades g
		LEFT JOIN ma_grade_translations gt ON g.id = gt.grade_id AND gt.language = ?
		WHERE g.deleted_dt IS NULL`

	GradeListCountQuery = `SELECT COUNT(*)
		FROM ma_grades g
		LEFT JOIN ma_grade_translations gt ON g.id = gt.grade_id AND gt.language = ?
		WHERE g.deleted_dt IS NULL`
)

// Find queries - no translation (direct lookup)
const (
	GradeFindByID = `SELECT ` + GradeSelectColumnsNoTranslation + `
		FROM ma_grades
		WHERE id = ? AND deleted_dt IS NULL`

	GradeFindByLabel = `SELECT ` + GradeSelectColumnsNoTranslation + `
		FROM ma_grades
		WHERE label = ? AND deleted_dt IS NULL`
)

// Find queries with translation
const (
	GradeFindByIDWithTranslation = `SELECT ` + GradeSelectColumns + `
		FROM ma_grades g
		LEFT JOIN ma_grade_translations gt ON g.id = gt.grade_id AND gt.language = ?
		WHERE g.id = ? AND g.deleted_dt IS NULL`

	GradeFindByLabelWithTranslation = `SELECT ` + GradeSelectColumns + `
		FROM ma_grades g
		LEFT JOIN ma_grade_translations gt ON g.id = gt.grade_id AND gt.language = ?
		WHERE (g.label = ? OR gt.label = ?) AND g.deleted_dt IS NULL`
)

// Mutation queries
const (
	GradeInsert = `INSERT INTO ma_grades (id, label, discription, image_key, grade_status, display_order, create_dt, modify_dt)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	GradeTranslationInsert = `INSERT INTO ma_grade_translations (id, grade_id, language, label, description, note, create_dt, modify_dt)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	GradeUpdate = `UPDATE ma_grades
		SET label = COALESCE(?, label),
			discription = COALESCE(?, discription),
			image_key = COALESCE(?, image_key),
			grade_status = COALESCE(?, grade_status),
			note = COALESCE(?, note),
			display_order = COALESCE(?, display_order),
			modify_dt = ?
		WHERE id = ? AND deleted_dt IS NULL`

	GradeTranslationUpdate = `UPDATE ma_grade_translations
		SET label = COALESCE(?, label),
			description = COALESCE(?, description),
			note = COALESCE(?, note),
			modify_dt = ?
		WHERE grade_id = ? AND language = ? AND deleted_dt IS NULL`

	GradeDelete = `UPDATE ma_grades
		SET deleted_dt = ?, modify_dt = ?
		WHERE id = ? AND deleted_dt IS NULL`

	GradeTranslationDeleteByGradeID = `UPDATE ma_grade_translations
		SET deleted_dt = ?, modify_dt = ?
		WHERE grade_id = ? AND deleted_dt IS NULL`

	GradeForceDelete = `DELETE FROM grades WHERE id = ?`

	GradeTranslationForceDeleteByGradeID = `DELETE FROM ma_grade_translations WHERE grade_id = ?`
)

// BuildListQueryWithSearch adds search condition to list query
// Supports searching by label and description in translated or default values
func (GradeQueries) BuildListQueryWithSearch(language, searchTerm string) (string, string, []interface{}) {
	searchCondition := ` AND (COALESCE(gt.label, g.label) LIKE ? OR COALESCE(gt.description, g.discription) LIKE ?)`
	searchPattern := "%" + searchTerm + "%"

	query := GradeListQuery + searchCondition
	countQuery := GradeListCountQuery + searchCondition

	// Args: language for JOIN, search pattern x2 for LIKE
	args := []interface{}{language, searchPattern, searchPattern}

	return query, countQuery, args
}
