package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"math-ai.com/math-ai/internal/domain/quiz"
	mtime "math-ai.com/math-ai/internal/domain/shared/time"
	"math-ai.com/math-ai/internal/infrastructure/database"
	"math-ai.com/math-ai/internal/infrastructure/persistence/mysql/models"
	"math-ai.com/math-ai/internal/shared/enum"
	"math-ai.com/math-ai/internal/shared/pagination"
	"math-ai.com/math-ai/internal/shared/utils"
)

const (
	quizTable = "ma_quizzes"

	quizColumns = `q.id, q.quiz_id, q.user_id, q.profile_id, q.type,
		q.questions, q.answers, q.ai_review, q.ai_detect_grade,
		q.total_questions, q.correct_number, q.score_percentage,
		q.previous_quiz_id, q.note, q.quiz_status, q.status,
		q.create_id, q.create_dt, q.modify_id, q.modify_dt`

	quizActiveWhere = `q.status IN (?) AND q.deleted_dt IS NULL`
)

func quizActiveArgs() []any {
	return []any{enum.StatusActive}
}

type QuizRepository struct {
	db database.Executor
}

func NewQuizRepository(db database.Executor) quiz.IRepository {
	return &QuizRepository{db: db}
}

func scanQuiz(s database.RowScanner) (*models.QuizModel, error) {
	var m models.QuizModel
	if err := s.Scan(&m.Id, &m.QuizId, &m.UserId, &m.ProfileId, &m.QuizType,
		&m.Questions, &m.Answers, &m.AIReview, &m.AIDetectGrade,
		&m.TotalQuestions, &m.CorrectNumber, &m.ScorePercentage,
		&m.PreviousQuizId, &m.Note, &m.QuizStatus, &m.Status,
		&m.CreateId, &m.CreateDt, &m.ModifyId, &m.ModifyDt); err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *QuizRepository) findOneBy(ctx context.Context, where string, args ...any) (*quiz.Quiz, error) {
	fullArgs := append(quizActiveArgs(), args...)
	query := `SELECT ` + quizColumns + ` FROM ` + quizTable + ` q WHERE ` +
		quizActiveWhere + ` AND (` + where + `)`

	m, err := scanQuiz(r.db.QueryRow(ctx, query, fullArgs...))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("quiz repo find (%s): %w", where, err)
	}
	return ModelToDomainQuiz(m), nil
}

func (r *QuizRepository) findBareById(ctx context.Context, id int64) (*quiz.Quiz, error) {
	args := append(quizActiveArgs(), id)
	query := `SELECT ` + quizColumns + ` FROM ` + quizTable + ` q WHERE ` +
		quizActiveWhere + ` AND q.id = ?`

	m, err := scanQuiz(r.db.QueryRow(ctx, query, args...))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("quiz repo find bare by id: %w", err)
	}
	return ModelToDomainQuiz(m), nil
}

func (r *QuizRepository) FindByQuizId(ctx context.Context, quizId string) (*quiz.Quiz, error) {
	return r.findOneBy(ctx, "q.quiz_id = ?", quizId)
}

// ListQuizzes returns the quiz history (newest first), with pagination,
// filtered by any combination of profile_id and user_id. Listing surfaces
// both GENERATED (unsubmitted) and SUBMITTED rows so the client can
// resume an in-flight quiz.
func (r *QuizRepository) ListQuizzes(ctx context.Context, filter quiz.ListQuizzesFilter, page, limit int64) ([]*quiz.Quiz, *pagination.Pagination, error) {
	if limit <= 0 {
		limit = pagination.DefaultPageSize
	}
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * limit

	filterWhere, filterArgs := buildQuizListFilterClause(filter)

	countArgs := append(quizActiveArgs(), filterArgs...)
	countQuery := `SELECT COUNT(*) FROM ` + quizTable + ` q WHERE ` +
		quizActiveWhere + filterWhere

	var total int64
	if err := r.db.QueryRow(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, nil, fmt.Errorf("quiz repo count: %w", err)
	}

	args := append(quizActiveArgs(), filterArgs...)
	args = append(args, limit, offset)
	query := `SELECT ` + quizColumns + ` FROM ` + quizTable + ` q WHERE ` +
		quizActiveWhere + filterWhere + ` ORDER BY q.id DESC LIMIT ? OFFSET ?`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, nil, fmt.Errorf("quiz repo list: %w", err)
	}
	defer rows.Close()

	var quizzes []*quiz.Quiz
	for rows.Next() {
		m, err := scanQuiz(rows)
		if err != nil {
			return nil, nil, fmt.Errorf("quiz repo scan row: %w", err)
		}
		quizzes = append(quizzes, ModelToDomainQuiz(m))
	}
	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("quiz repo rows iteration: %w", err)
	}
	return quizzes, pagination.NewPagination(page, limit, total), nil
}

// buildQuizListFilterClause appends an "AND q.profile_id = ?" / "AND
// q.user_id = ?" pair for each non-nil field, keeping placeholder
// ordering in lockstep with the returned args slice.
func buildQuizListFilterClause(filter quiz.ListQuizzesFilter) (string, []any) {
	var (
		clause string
		args   []any
	)
	if filter.ProfileID != nil {
		clause += ` AND q.profile_id = ?`
		args = append(args, *filter.ProfileID)
	}
	if filter.UserID != nil {
		clause += ` AND q.user_id = ?`
		args = append(args, *filter.UserID)
	}
	return clause, args
}

func (r *QuizRepository) Create(ctx context.Context, q *quiz.Quiz) (*quiz.Quiz, error) {
	query := `
		INSERT INTO ` + quizTable + `
			(quiz_id, user_id, profile_id, type, questions, answers,
			 ai_review, ai_detect_grade, previous_quiz_id, note, quiz_status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := r.db.Exec(ctx, query,
		q.QuizId(), q.UserId(), q.ProfileId(), q.QuizType(),
		q.Questions(), q.Answers(), q.AIReview(), q.AIDetectGrade(),
		q.PreviousQuizId(), q.Note(), q.QuizStatus())
	if err != nil {
		return nil, fmt.Errorf("quiz repo create: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("quiz repo last insert id: %w", err)
	}
	return r.findBareById(ctx, id)
}

// UpdateAnswersAndGrading is the only mutator on a quiz row. Submission
// is the single state transition: GENERATED → SUBMITTED, writing answers
// + AI grading + score counts in one shot. ai_detect_grade is *string
// because PRACTICE quizzes don't predict a grade.
func (r *QuizRepository) UpdateAnswersAndGrading(ctx context.Context, quizId string,
	answers string, grading quiz.GradingUpdate, quizStatus string) error {
	query := `
		UPDATE ` + quizTable + `
		SET answers          = ?,
			ai_review        = ?,
			ai_detect_grade  = ?,
			total_questions  = ?,
			correct_number   = ?,
			score_percentage = ?,
			quiz_status      = ?,
			modify_dt        = ?
		WHERE quiz_id = ?
	`
	if _, err := r.db.Exec(ctx, query,
		answers, grading.AIReview, grading.AIDetectGrade,
		grading.TotalQuestions, grading.CorrectNumber, grading.ScorePercentage,
		quizStatus, mtime.Now().Time, quizId); err != nil {
		return fmt.Errorf("quiz repo update answers and grading: %w", err)
	}
	return nil
}

func (r *QuizRepository) SoftDelete(ctx context.Context, quizId string) error {
	query := `
		UPDATE ` + quizTable + `
		SET quiz_status = ?,
			status      = ?,
			deleted_dt  = ?
		WHERE quiz_id = ?
	`
	if _, err := r.db.Exec(ctx, query,
		enum.QuizStatusTypeDeleted, enum.StatusInactive, mtime.Now().Time, quizId); err != nil {
		return fmt.Errorf("quiz repo soft delete: %w", err)
	}
	return nil
}

func (r *QuizRepository) SoftDeleteByUserId(ctx context.Context, userId string) error {
	query := `
		UPDATE ` + quizTable + `
		SET quiz_status = ?,
			status      = ?,
			deleted_dt  = ?
		WHERE user_id = ?
	`
	if _, err := r.db.Exec(ctx, query,
		enum.QuizStatusTypeDeleted, enum.StatusInactive, mtime.Now().Time, userId); err != nil {
		return fmt.Errorf("quiz repo soft delete by user id: %w", err)
	}
	return nil
}

func (r *QuizRepository) ForceDeleteByUserId(ctx context.Context, userId string) error {
	query := `DELETE FROM ` + quizTable + ` WHERE user_id = ?`
	if _, err := r.db.Exec(ctx, query, userId); err != nil {
		return fmt.Errorf("quiz repo force delete by user id: %w", err)
	}
	return nil
}

func ModelToDomainQuiz(m *models.QuizModel) *quiz.Quiz {
	quizId, err := utils.StringToUUID(m.QuizId)
	if err != nil {
		return nil
	}

	userId, err := utils.PtrStringToUUID(m.UserId)
	if err != nil {
		return nil
	}

	profileId, err := utils.PtrStringToUUID(m.ProfileId)
	if err != nil {
		return nil
	}

	previousQuizId, err := utils.PtrStringToUUID(m.PreviousQuizId)
	if err != nil {
		return nil
	}

	createId, err := utils.PtrStringToUUID(m.CreateId)
	if err != nil {
		return nil
	}

	modifyId, err := utils.PtrStringToUUID(m.ModifyId)
	if err != nil {
		return nil
	}

	q := quiz.NewQuiz()
	q.SetId(m.Id)
	q.SetQuizId(quizId)
	q.SetUserId(&userId)
	q.SetProfileId(&profileId)
	q.SetQuizType(m.QuizType)
	q.SetQuestions(m.Questions)
	q.SetAnswers(m.Answers)
	q.SetAIReview(m.AIReview)
	q.SetAIDetectGrade(m.AIDetectGrade)
	q.SetTotalQuestions(m.TotalQuestions)
	q.SetCorrectNumber(m.CorrectNumber)
	q.SetScorePercentage(m.ScorePercentage)
	q.SetPreviousQuizId(&previousQuizId)
	q.SetNote(m.Note)
	q.SetQuizStatus(m.QuizStatus)
	q.SetStatus(m.Status)
	q.SetCreateId(&createId)
	q.SetCreateDt(mtime.MathTime{Time: m.CreateDt})
	q.SetModifyId(&modifyId)
	q.SetModifyDt(mtime.MathTime{Time: m.ModifyDt})
	return q
}
