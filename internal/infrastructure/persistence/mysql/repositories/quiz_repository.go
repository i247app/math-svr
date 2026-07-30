package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"math-ai.com/math-ai/internal/domain/quiz"
	"math-ai.com/math-ai/internal/domain/shared/mtime"
	"math-ai.com/math-ai/internal/infrastructure/database"
	"math-ai.com/math-ai/internal/infrastructure/persistence/mysql/models"
	"math-ai.com/math-ai/internal/shared/enum"
	"math-ai.com/math-ai/internal/shared/pagination"
)

const (
	quizTable = "ma_quizzes"

	quizColumns = `q.id, q.quiz_id, q.user_id, q.profile_id, q.purpose, q.type_of_quiz, q.title, q.short_text,
		q.questions, q.answers, q.ai_review, q.assessment_grade,
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
	if err := s.Scan(&m.Id, &m.QuizId, &m.UserId, &m.ProfileId, &m.Purpose, &m.TypeOfQuiz, &m.Title, &m.ShortText,
		&m.Questions, &m.Answers, &m.AIReview, &m.AssessmentGrade,
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

func (r *QuizRepository) FindByQuizId(ctx context.Context, quizId int64) (*quiz.Quiz, error) {
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
	if filter.Purpose != nil && *filter.Purpose != "" {
		clause += ` AND q.purpose = ?`
		args = append(args, *filter.Purpose)
	}
	return clause, args
}

func (r *QuizRepository) Create(ctx context.Context, q *quiz.Quiz) (*quiz.Quiz, error) {
	query := `
		INSERT INTO ` + quizTable + `
			(quiz_id, user_id, profile_id, purpose, type_of_quiz, title, short_text, questions, answers,
			 ai_review, assessment_grade, previous_quiz_id, note, quiz_status, create_dt, modify_dt)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := r.db.Exec(ctx, query,
		q.QuizId(), q.UserId(), q.ProfileId(), q.Purpose(), q.TypeOfQuiz(), q.Title(), q.ShortText(),
		q.Questions(), q.Answers(), q.AIReview(), q.AssessmentGrade(),
		q.PreviousQuizId(), q.Note(), q.QuizStatus(), mtime.Now().Time, mtime.Now().Time)
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
// + AI grading + score counts in one shot. assessment_grade is *string
// because PRACTICE quizzes don't predict a grade.
func (r *QuizRepository) UpdateAnswersAndGrading(ctx context.Context, quizId int64,
	answers string, grading quiz.GradingUpdate, quizStatus string) error {
	query := `
		UPDATE ` + quizTable + `
		SET answers          = COALESCE(?, answers),
			ai_review        = COALESCE(?, ai_review),
			assessment_grade = COALESCE(?, assessment_grade),
			total_questions  = COALESCE(?, total_questions),
			correct_number   = COALESCE(?, correct_number),
			score_percentage = COALESCE(?, score_percentage),
			quiz_status      = ?,
			modify_dt        = ?
		WHERE quiz_id = ?
	`
	if _, err := r.db.Exec(ctx, query,
		answers, grading.AIReview, grading.AssessmentGrade,
		grading.TotalQuestions, grading.CorrectNumber, grading.ScorePercentage,
		quizStatus, mtime.Now().Time, quizId); err != nil {
		return fmt.Errorf("quiz repo update answers and grading: %w", err)
	}
	return nil
}

func (r *QuizRepository) SoftDelete(ctx context.Context, quizId int64) error {
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

func (r *QuizRepository) SoftDeleteByUserId(ctx context.Context, userId int64) error {
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

func (r *QuizRepository) ForceDeleteByUserId(ctx context.Context, userId int64) error {
	query := `DELETE FROM ` + quizTable + ` WHERE user_id = ?`
	if _, err := r.db.Exec(ctx, query, userId); err != nil {
		return fmt.Errorf("quiz repo force delete by user id: %w", err)
	}
	return nil
}

// progressPointColumns is the narrow projection for the analytics chart —
// no LONGTEXT questions/answers. Order == scanProgressPoint's Scan order.
const progressPointColumns = `q.quiz_id, q.purpose, q.type_of_quiz, q.title, q.short_text,
	q.score_percentage, q.correct_number, q.total_questions, q.modify_dt`

func scanProgressPoint(s database.RowScanner) (*quiz.ProgressPoint, error) {
	var (
		p          quiz.ProgressPoint
		typeOfQuiz *string
		title      *string
		shortText  *string
		correct    *int64
		total      *int64
		modifyDt   time.Time
	)
	if err := s.Scan(&p.QuizId, &p.Purpose, &typeOfQuiz, &title, &shortText,
		&p.ScorePercentage, &correct, &total, &modifyDt); err != nil {
		return nil, err
	}
	p.TypeOfQuiz = typeOfQuiz
	p.Title = title
	p.ShortText = shortText
	p.CorrectNumber = correct
	p.TotalQuestions = total
	p.CompletedDt = mtime.NewTime(modifyDt)
	return &p, nil
}

// ListProgressPoints returns completed (SUBMITTED, graded) quizzes for one
// profile, newest first, capped at params.Limit. The caller reverses to
// chronological order. Only the columns the chart needs are selected.
func (r *QuizRepository) ListProgressPoints(ctx context.Context, params quiz.ProgressPointsParams) ([]*quiz.ProgressPoint, error) {
	where := quizActiveWhere +
		` AND q.quiz_status = ? AND q.profile_id = ? AND q.score_percentage IS NOT NULL`
	args := append(quizActiveArgs(), string(enum.QuizStatusTypeSubmitted), params.ProfileID)

	if params.Purpose != nil && *params.Purpose != "" {
		where += ` AND q.purpose = ?`
		args = append(args, *params.Purpose)
	}
	if params.From != nil && !params.From.IsZero() {
		where += ` AND q.modify_dt >= ?`
		args = append(args, params.From.Time)
	}
	if params.To != nil && !params.To.IsZero() {
		where += ` AND q.modify_dt <= ?`
		args = append(args, params.To.Time)
	}
	if params.CompletedBefore != nil && !params.CompletedBefore.IsZero() {
		where += ` AND q.modify_dt < ?`
		args = append(args, params.CompletedBefore.Time)
	}

	limit := params.Limit
	if limit <= 0 {
		limit = 10
	}
	args = append(args, limit)

	query := `SELECT ` + progressPointColumns + ` FROM ` + quizTable + ` q WHERE ` +
		where + ` ORDER BY q.modify_dt DESC, q.quiz_id DESC LIMIT ?`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("quiz repo list progress points: %w", err)
	}
	defer rows.Close()

	var points []*quiz.ProgressPoint
	for rows.Next() {
		p, err := scanProgressPoint(rows)
		if err != nil {
			return nil, fmt.Errorf("quiz repo scan progress point: %w", err)
		}
		points = append(points, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("quiz repo progress points iteration: %w", err)
	}
	return points, nil
}

func ModelToDomainQuiz(m *models.QuizModel) *quiz.Quiz {
	q := quiz.NewQuiz()
	q.SetId(m.Id)
	q.SetQuizId(m.QuizId)
	q.SetUserId(m.UserId)
	q.SetProfileId(m.ProfileId)
	q.SetPurpose(m.Purpose)
	q.SetTypeOfQuiz(m.TypeOfQuiz)
	q.SetTitle(m.Title)
	q.SetShortText(m.ShortText)
	q.SetQuestions(m.Questions)
	q.SetAnswers(m.Answers)
	q.SetAIReview(m.AIReview)
	q.SetAssessmentGrade(m.AssessmentGrade)
	q.SetTotalQuestions(m.TotalQuestions)
	q.SetCorrectNumber(m.CorrectNumber)
	q.SetScorePercentage(m.ScorePercentage)
	q.SetPreviousQuizId(m.PreviousQuizId)
	q.SetNote(m.Note)
	q.SetQuizStatus(m.QuizStatus)
	q.SetStatus(m.Status)
	q.SetCreateId(m.CreateId)
	q.SetCreateDt(mtime.MathTime{Time: m.CreateDt})
	q.SetModifyId(m.ModifyId)
	q.SetModifyDt(mtime.MathTime{Time: m.ModifyDt})
	return q
}
