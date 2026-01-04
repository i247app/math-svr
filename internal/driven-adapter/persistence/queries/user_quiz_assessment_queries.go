package queries

// UserQuizAssessmentQueries contains all SQL queries for user quiz assessment repository
type UserQuizAssessmentQueries struct{}

// Column lists
const (
	UserQuizAssessmentSelectColumns = `id, uid, questions, answers, ai_review, ai_detect_grade, status,
		create_id, create_dt, modify_id, modify_dt`
)

// Base queries
const (
	UserQuizAssessmentBaseSelectByUID = `SELECT ` + UserQuizAssessmentSelectColumns + `
		FROM ma_user_quiz_assessments
		WHERE uid = ? AND deleted_dt IS NULL`

	UserQuizAssessmentCountByUID = `SELECT COUNT(*)
		FROM ma_user_quiz_assessments
		WHERE uid = ? AND deleted_dt IS NULL`
)

// Find queries
const (
	UserQuizAssessmentFindByID = `SELECT ` + UserQuizAssessmentSelectColumns + `
		FROM ma_user_quiz_assessments
		WHERE id = ? AND deleted_dt IS NULL`
)

// Mutation queries
const (
	UserQuizAssessmentInsert = `INSERT INTO ma_user_quiz_assessments (id, uid, questions, answers, ai_review, ai_detect_grade, status, create_dt, modify_dt)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	UserQuizAssessmentUpdate = `UPDATE ma_user_quiz_assessments
		SET questions = COALESCE(?, questions),
			answers = COALESCE(?, answers),
			ai_review = COALESCE(?, ai_review),
			ai_detect_grade = COALESCE(?, ai_detect_grade),
			modify_dt = ?
		WHERE id = ? AND deleted_dt IS NULL`

	UserQuizAssessmentSoftDelete = `UPDATE ma_user_quiz_assessments
		SET deleted_dt = ?,
			modify_dt = ?
		WHERE id = ? AND deleted_dt IS NULL`

	UserQuizAssessmentForceDelete = `DELETE FROM ma_user_quiz_assessments
		WHERE id = ?`
)
