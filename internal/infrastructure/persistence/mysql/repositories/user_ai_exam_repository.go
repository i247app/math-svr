package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"math-ai.com/math-ai/internal/domain/exam"
	"math-ai.com/math-ai/internal/domain/shared/mtime"
	"math-ai.com/math-ai/internal/infrastructure/database"
	"math-ai.com/math-ai/internal/infrastructure/persistence/mysql/models"
	"math-ai.com/math-ai/internal/shared/enum"
	"math-ai.com/math-ai/internal/shared/pagination"
)

const (
	userAiExamTable = "ma_user_ai_exams"

	userAiExamColumns = `u.id, u.user_ai_exam_id, u.uid, u.profile_id, u.ai_exam_id, u.user_exam_id, u.shuffle_map,
		u.req_exam_type, u.req_grade, u.req_level,
		u.res_total_questions, u.res_correct_number, u.res_skipped_number, u.res_score_percentage,
		u.started_dt, u.submitted_dt,
		u.note, u.user_ai_exam_status, u.status,
		u.create_id, u.create_dt, u.modify_id, u.modify_dt`

	userAiExamActiveWhere = `u.status IN (?) AND u.deleted_dt IS NULL`
)

func userAiExamActiveArgs() []any {
	return []any{enum.StatusActive}
}

type UserAiExamRepository struct {
	db database.Executor
}

func NewUserAiExamRepository(db database.Executor) exam.IUserAiExamRepository {
	return &UserAiExamRepository{db: db}
}

func scanUserAiExam(s database.RowScanner) (*models.UserAiExamModel, error) {
	var m models.UserAiExamModel
	if err := s.Scan(&m.Id, &m.UserAiExamId, &m.UserId, &m.ProfileId, &m.AiExamId, &m.UserExamId, &m.ShuffleMap,
		&m.ReqExamType, &m.ReqGrade, &m.ReqLevel,
		&m.ResTotalQuestions, &m.ResCorrectNumber, &m.ResSkippedNumber, &m.ResScorePercentage,
		&m.StartedDt, &m.SubmittedDt,
		&m.Note, &m.UserAiExamStatus, &m.Status,
		&m.CreateId, &m.CreateDt, &m.ModifyId, &m.ModifyDt); err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *UserAiExamRepository) findOneBy(ctx context.Context, where string, args ...any) (*exam.UserAiExam, error) {
	fullArgs := slices.Concat(args, userAiExamActiveArgs())
	query := `SELECT ` + userAiExamColumns + ` FROM ` + userAiExamTable + ` u WHERE (` +
		where + `) AND ` + userAiExamActiveWhere

	m, err := scanUserAiExam(r.db.QueryRow(ctx, query, fullArgs...))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("user ai exam repo find (%s): %w", where, err)
	}
	return ModelToDomainUserAiExam(m), nil
}

func (r *UserAiExamRepository) FindByUserAiExamId(ctx context.Context, userAiExamId int64) (*exam.UserAiExam, error) {
	return r.findOneBy(ctx, "u.user_ai_exam_id = ?", userAiExamId)
}

func (r *UserAiExamRepository) findBareById(ctx context.Context, id int64) (*exam.UserAiExam, error) {
	args := slices.Concat([]any{id}, userAiExamActiveArgs())
	query := `SELECT ` + userAiExamColumns + ` FROM ` + userAiExamTable + ` u WHERE (u.id = ?) AND ` +
		userAiExamActiveWhere

	m, err := scanUserAiExam(r.db.QueryRow(ctx, query, args...))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("user ai exam repo find bare by id: %w", err)
	}
	return ModelToDomainUserAiExam(m), nil
}

// ListAttempts returns one child's attempt history, newest first. Rows in
// every state are returned, including IN_PROGRESS ones: an unfinished exam
// is exactly what the "you left this open" surface needs, and the caller
// filters by status when it wants only those.
func (r *UserAiExamRepository) ListAttempts(ctx context.Context, filter exam.ListAttemptsFilter, page, limit int64) ([]*exam.UserAiExam, *pagination.Pagination, error) {
	if limit <= 0 {
		limit = pagination.DefaultPageSize
	}
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * limit

	filterWhere, filterArgs := buildUserAiExamFilterClause(filter)

	countArgs := slices.Concat(filterArgs, userAiExamActiveArgs())
	countQuery := `SELECT COUNT(*) FROM ` + userAiExamTable + ` u` +
		whereActive(filterWhere, userAiExamActiveWhere)

	var total int64
	if err := r.db.QueryRow(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, nil, fmt.Errorf("user ai exam repo count: %w", err)
	}

	args := slices.Concat(filterArgs, userAiExamActiveArgs(), []any{limit, offset})
	query := `SELECT ` + userAiExamColumns + ` FROM ` + userAiExamTable + ` u` +
		whereActive(filterWhere, userAiExamActiveWhere) + ` ORDER BY u.id DESC LIMIT ? OFFSET ?`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, nil, fmt.Errorf("user ai exam repo list: %w", err)
	}
	defer rows.Close()

	var attempts []*exam.UserAiExam
	for rows.Next() {
		m, err := scanUserAiExam(rows)
		if err != nil {
			return nil, nil, fmt.Errorf("user ai exam repo scan row: %w", err)
		}
		attempts = append(attempts, ModelToDomainUserAiExam(m))
	}
	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("user ai exam repo rows iteration: %w", err)
	}
	return attempts, pagination.NewPagination(page, limit, total), nil
}

// buildUserAiExamFilterClause keeps placeholder order in lockstep with the
// returned args slice.
func buildUserAiExamFilterClause(filter exam.ListAttemptsFilter) (string, []any) {
	var (
		clause string
		args   []any
	)
	if filter.ProfileID != 0 {
		clause += ` AND u.profile_id = ?`
		args = append(args, filter.ProfileID)
	}
	if filter.ExamType != nil && *filter.ExamType != "" {
		clause += ` AND u.req_exam_type = ?`
		args = append(args, *filter.ExamType)
	}
	if filter.Status != nil && *filter.Status != "" {
		clause += ` AND u.user_ai_exam_status = ?`
		args = append(args, *filter.Status)
	}
	if filter.UserExamID != nil && *filter.UserExamID != 0 {
		clause += ` AND u.user_exam_id = ?`
		args = append(args, *filter.UserExamID)
	}
	return clause, args
}

// ListInProgressByProfile reads ix_profile_status_started: one child, one
// status, in the order the sittings were handed out.
func (r *UserAiExamRepository) ListInProgressByProfile(ctx context.Context, profileId int64) ([]*exam.UserAiExam, error) {
	args := slices.Concat([]any{profileId, string(enum.UserAiExamStatusInProgress)}, userAiExamActiveArgs())
	query := `SELECT ` + userAiExamColumns + ` FROM ` + userAiExamTable + ` u WHERE ` +
		`(u.profile_id = ? AND u.user_ai_exam_status = ?) AND ` + userAiExamActiveWhere +
		` ORDER BY u.started_dt ASC, u.id ASC`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("user ai exam repo list in progress: %w", err)
	}
	defer rows.Close()

	var out []*exam.UserAiExam
	for rows.Next() {
		m, err := scanUserAiExam(rows)
		if err != nil {
			return nil, fmt.Errorf("user ai exam repo scan row: %w", err)
		}
		out = append(out, ModelToDomainUserAiExam(m))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("user ai exam repo rows iteration: %w", err)
	}
	return out, nil
}

// FindLatestSubmittedByUserExamId walks ix_user_exam_submitted backwards:
// the newest submitted_dt of the journey, id as the tie-break so two
// sittings submitted in the same microsecond still order deterministically.
func (r *UserAiExamRepository) FindLatestSubmittedByUserExamId(ctx context.Context, userExamId int64) (*exam.UserAiExam, error) {
	args := slices.Concat([]any{userExamId, string(enum.UserAiExamStatusSubmitted)}, userAiExamActiveArgs())
	query := `SELECT ` + userAiExamColumns + ` FROM ` + userAiExamTable + ` u WHERE ` +
		`(u.user_exam_id = ? AND u.user_ai_exam_status = ?) AND ` + userAiExamActiveWhere +
		` ORDER BY u.submitted_dt DESC, u.id DESC LIMIT 1`

	m, err := scanUserAiExam(r.db.QueryRow(ctx, query, args...))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("user ai exam repo find latest submitted: %w", err)
	}
	return ModelToDomainUserAiExam(m), nil
}

// ListRecentByProfileGrade reads a child's newest sittings at one grade.
// It walks ix_profile_status_started's profile prefix and filters on
// req_grade; the candidate set per child is small enough that a
// dedicated index is not worth its write cost.
func (r *UserAiExamRepository) ListRecentByProfileGrade(ctx context.Context, profileId int64, grade int, limit int) ([]*exam.UserAiExam, error) {
	if limit <= 0 {
		return nil, nil
	}
	args := slices.Concat([]any{profileId, grade}, userAiExamActiveArgs(), []any{limit})
	query := `SELECT ` + userAiExamColumns + ` FROM ` + userAiExamTable + ` u WHERE ` +
		`(u.profile_id = ? AND u.req_grade = ?) AND ` + userAiExamActiveWhere +
		` ORDER BY u.started_dt DESC, u.id DESC LIMIT ?`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("user ai exam repo list recent by grade: %w", err)
	}
	defer rows.Close()

	var out []*exam.UserAiExam
	for rows.Next() {
		m, err := scanUserAiExam(rows)
		if err != nil {
			return nil, fmt.Errorf("user ai exam repo scan row: %w", err)
		}
		out = append(out, ModelToDomainUserAiExam(m))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("user ai exam repo rows iteration: %w", err)
	}
	return out, nil
}

// ListByUserAiExamIds hydrates a batch of attempts, oldest first. The IN
// list is built from the id count so the query stays parameterised.
func (r *UserAiExamRepository) ListByUserAiExamIds(ctx context.Context, userAiExamIds []int64) ([]*exam.UserAiExam, error) {
	if len(userAiExamIds) == 0 {
		return nil, nil
	}

	placeholders := make([]string, 0, len(userAiExamIds))
	args := make([]any, 0, len(userAiExamIds)+1)
	for _, id := range userAiExamIds {
		placeholders = append(placeholders, "?")
		args = append(args, id)
	}
	args = append(args, userAiExamActiveArgs()...)

	query := `SELECT ` + userAiExamColumns + ` FROM ` + userAiExamTable + ` u WHERE (u.user_ai_exam_id IN (` +
		strings.Join(placeholders, ", ") + `)) AND ` + userAiExamActiveWhere +
		` ORDER BY u.user_ai_exam_id ASC`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("user ai exam repo list by ids: %w", err)
	}
	defer rows.Close()

	var out []*exam.UserAiExam
	for rows.Next() {
		m, err := scanUserAiExam(rows)
		if err != nil {
			return nil, fmt.Errorf("user ai exam repo scan row: %w", err)
		}
		out = append(out, ModelToDomainUserAiExam(m))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("user ai exam repo rows iteration: %w", err)
	}
	return out, nil
}

func (r *UserAiExamRepository) Create(ctx context.Context, a *exam.UserAiExam) (*exam.UserAiExam, error) {
	query := `
		INSERT INTO ` + userAiExamTable + `
			(user_ai_exam_id, uid, profile_id, ai_exam_id, user_exam_id, shuffle_map,
			 req_exam_type, req_grade, req_level,
			 started_dt, note, user_ai_exam_status, create_id, create_dt, modify_dt)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	now := mtime.Now().Time
	startedDt := mtime.MathTimePtrToTime(a.StartedDt().Ptr())

	result, err := r.db.Exec(ctx, query,
		a.UserAiExamId(), a.UserId(), a.ProfileId(), a.AiExamId(), a.UserExamId(), a.ShuffleMap(),
		a.ReqExamType(), a.ReqGrade(), a.ReqLevel(),
		startedDt, a.Note(), a.UserAiExamStatus(), a.CreateId(), now, now)
	if err != nil {
		return nil, fmt.Errorf("user ai exam repo create: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("user ai exam repo last insert id: %w", err)
	}
	return r.findBareById(ctx, id)
}

// MarkSubmitted performs the attempt's single state transition and writes
// the sitting's result in the same statement.
//
// The WHERE clause carries the expected state. Two submits racing on one
// attempt both read IN_PROGRESS before either writes, so checking in the
// command is not enough — here the loser matches zero rows and gets
// exam.ErrAttemptNotInProgress instead of silently double-counting the
// answers into the lifetime totals.
//
// user_exam_id is COALESCEd: a nil result leaves what the hand-out wrote,
// a value pins the sitting to the journey it actually folded into.
func (r *UserAiExamRepository) MarkSubmitted(ctx context.Context, userAiExamId int64, res exam.AttemptResult) error {
	query := `
		UPDATE ` + userAiExamTable + `
		SET res_total_questions  = ?,
			res_correct_number   = ?,
			res_skipped_number   = ?,
			res_score_percentage = ?,
			submitted_dt         = ?,
			user_exam_id         = COALESCE(?, user_exam_id),
			user_ai_exam_status  = ?,
			modify_dt            = ?
		WHERE user_ai_exam_id = ? AND user_ai_exam_status = ?
	`
	submittedDt := res.SubmittedDt.Time
	if res.SubmittedDt.IsZero() {
		submittedDt = mtime.Now().Time
	}

	result, err := r.db.Exec(ctx, query,
		res.TotalQuestions, res.CorrectNumber, res.SkippedNumber, res.ScorePercentage,
		submittedDt, res.UserExamId, string(enum.UserAiExamStatusSubmitted), mtime.Now().Time,
		userAiExamId, string(enum.UserAiExamStatusInProgress))
	if err != nil {
		return fmt.Errorf("user ai exam repo mark submitted: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("user ai exam repo mark submitted rows affected: %w", err)
	}
	if affected == 0 {
		return exam.ErrAttemptNotInProgress
	}
	return nil
}

// progressPointColumnsExam is the narrow projection for the analytics
// chart. Order == scanExamProgressPoint's Scan order.
const progressPointColumnsExam = `u.user_ai_exam_id, u.ai_exam_id, u.req_exam_type, u.req_grade,
	u.res_score_percentage, u.res_correct_number, u.res_total_questions, u.submitted_dt`

func scanExamProgressPoint(s database.RowScanner) (*exam.ProgressPoint, error) {
	var (
		p           exam.ProgressPoint
		correct     *int64
		total       *int64
		submittedDt *time.Time
	)
	if err := s.Scan(&p.UserAiExamId, &p.AiExamId, &p.ExamType, &p.Grade,
		&p.ScorePercentage, &correct, &total, &submittedDt); err != nil {
		return nil, err
	}
	p.CorrectNumber = correct
	p.TotalQuestions = total
	p.CompletedDt = mtime.MathTimeFromPtr(submittedDt)
	return &p, nil
}

// ListProgressPoints returns submitted, scored attempts for one profile,
// newest first, capped at params.Limit. The caller reverses into
// chronological order.
func (r *UserAiExamRepository) ListProgressPoints(ctx context.Context, params exam.ProgressPointsParams) ([]*exam.ProgressPoint, error) {
	where := `u.user_ai_exam_status = ? AND u.profile_id = ? AND u.res_score_percentage IS NOT NULL`
	args := []any{string(enum.UserAiExamStatusSubmitted), params.ProfileID}

	if params.ExamType != nil && *params.ExamType != "" {
		where += ` AND u.req_exam_type = ?`
		args = append(args, *params.ExamType)
	}
	if params.UserExamID != nil && *params.UserExamID != 0 {
		where += ` AND u.user_exam_id = ?`
		args = append(args, *params.UserExamID)
	}
	if params.From != nil && !params.From.IsZero() {
		where += ` AND u.submitted_dt >= ?`
		args = append(args, params.From.Time)
	}
	if params.To != nil && !params.To.IsZero() {
		where += ` AND u.submitted_dt <= ?`
		args = append(args, params.To.Time)
	}
	if params.CompletedBefore != nil && !params.CompletedBefore.IsZero() {
		where += ` AND u.submitted_dt < ?`
		args = append(args, params.CompletedBefore.Time)
	}

	limit := params.Limit
	if limit <= 0 {
		limit = 10
	}
	args = append(args, userAiExamActiveArgs()...)
	args = append(args, limit)

	query := `SELECT ` + progressPointColumnsExam + ` FROM ` + userAiExamTable + ` u WHERE (` +
		where + `) AND ` + userAiExamActiveWhere + ` ORDER BY u.submitted_dt DESC, u.user_ai_exam_id DESC LIMIT ?`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("user ai exam repo list progress points: %w", err)
	}
	defer rows.Close()

	var points []*exam.ProgressPoint
	for rows.Next() {
		p, err := scanExamProgressPoint(rows)
		if err != nil {
			return nil, fmt.Errorf("user ai exam repo scan progress point: %w", err)
		}
		points = append(points, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("user ai exam repo progress points iteration: %w", err)
	}
	return points, nil
}

func ModelToDomainUserAiExam(m *models.UserAiExamModel) *exam.UserAiExam {
	a := exam.NewUserAiExam()
	a.SetId(m.Id)
	a.SetUserAiExamId(m.UserAiExamId)
	a.SetUserId(m.UserId)
	a.SetProfileId(m.ProfileId)
	a.SetAiExamId(m.AiExamId)
	a.SetUserExamId(m.UserExamId)
	a.SetShuffleMap(m.ShuffleMap)
	a.SetReqExamType(m.ReqExamType)
	a.SetReqGrade(m.ReqGrade)
	a.SetReqLevel(m.ReqLevel)
	a.SetResTotalQuestions(m.ResTotalQuestions)
	a.SetResCorrectNumber(m.ResCorrectNumber)
	a.SetResSkippedNumber(m.ResSkippedNumber)
	a.SetResScorePercentage(m.ResScorePercentage)
	a.SetStartedDt(mtime.MathTimeFromPtr(m.StartedDt))
	a.SetSubmittedDt(mtime.MathTimeFromPtr(m.SubmittedDt))
	a.SetNote(m.Note)
	a.SetUserAiExamStatus(m.UserAiExamStatus)
	a.SetStatus(m.Status)
	a.SetCreateId(m.CreateId)
	a.SetCreateDt(mtime.MathTime{Time: m.CreateDt})
	a.SetModifyId(m.ModifyId)
	a.SetModifyDt(mtime.MathTime{Time: m.ModifyDt})
	return a
}
