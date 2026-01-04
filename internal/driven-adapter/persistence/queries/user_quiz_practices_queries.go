package queries

// UserQuizPracticesQueries contains all SQL queries for user quiz practices repository
type UserQuizPracticesQueries struct{}

// Column lists
const (
	UserQuizPracticesSelectColumns = `id, uid, questions, answers, ai_review, status,
		create_id, create_dt, modify_id, modify_dt`
)

// Find queries
const (
	UserQuizPracticesFindByID = `SELECT ` + UserQuizPracticesSelectColumns + `
		FROM ma_user_quiz_practices
		WHERE id = ? AND deleted_dt IS NULL`

	UserQuizPracticesFindByUID = `SELECT ` + UserQuizPracticesSelectColumns + `
		FROM ma_user_quiz_practices
		WHERE uid = ? AND deleted_dt IS NULL
		ORDER BY create_dt DESC
		LIMIT 1`
)

// Mutation queries
const (
	UserQuizPracticesInsert = `INSERT INTO ma_user_quiz_practices (id, uid, questions, answers, ai_review, status, create_dt, modify_dt)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	UserQuizPracticesUpdate = `UPDATE ma_user_quiz_practices
		SET questions = COALESCE(?, questions),
			answers = COALESCE(?, answers),
			ai_review = COALESCE(?, ai_review),
			modify_dt = ?
		WHERE id = ? AND deleted_dt IS NULL`

	UserQuizPracticesSoftDelete = `UPDATE ma_user_quiz_practices
		SET deleted_dt = ?,
			modify_dt = ?
		WHERE id = ? AND deleted_dt IS NULL`

	UserQuizPracticesSoftDeleteByUID = `UPDATE ma_user_quiz_practices
		SET deleted_dt = ?,
			modify_dt = ?
		WHERE uid = ? AND deleted_dt IS NULL`

	UserQuizPracticesForceDelete = `DELETE FROM ma_user_quiz_practices
		WHERE id = ?`

	UserQuizPracticesForceDeleteByUID = `DELETE FROM ma_user_quiz_practices
		WHERE uid = ?`
)
