package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"time"

	"math-ai.com/math-ai/internal/domain/exam"
	"math-ai.com/math-ai/internal/domain/shared/mtime"
	"math-ai.com/math-ai/internal/infrastructure/database"
	"math-ai.com/math-ai/internal/infrastructure/persistence/mysql/models"
	"math-ai.com/math-ai/internal/shared/enum"
)

const (
	userExamTable = "ma_user_exams"

	userExamColumns = `e.id, e.user_exam_id, e.user_id, e.profile_id, e.req_exam_type,
		e.res_total_questions, e.res_correct_number, e.res_skipped_number, e.res_score_percentage,
		e.res_review, e.current_grade, e.current_level, e.last_submitted_dt, e.ended_dt,
		e.note, e.user_exam_status, e.status,
		e.create_id, e.create_dt, e.modify_id, e.modify_dt`

	userExamActiveWhere = `e.status = ? AND (e.user_exam_status IS NULL OR e.user_exam_status != ?) AND e.deleted_dt IS NULL`
)

func userExamActiveArgs() []any {
	return []any{enum.StatusActive, string(enum.UserExamStatusDeleted)}
}

type UserExamRepository struct {
	db database.Executor
}

func NewUserExamRepository(db database.Executor) exam.IUserExamRepository {
	return &UserExamRepository{db: db}
}

func scanUserExam(s database.RowScanner) (*models.UserExamModel, error) {
	var m models.UserExamModel
	if err := s.Scan(&m.Id, &m.UserExamId, &m.UserId, &m.ProfileId, &m.ReqExamType,
		&m.ResTotalQuestions, &m.ResCorrectNumber, &m.ResSkippedNumber, &m.ResScorePercentage,
		&m.ResReview, &m.CurrentGrade, &m.CurrentLevel, &m.LastSubmittedDt, &m.EndedDt,
		&m.Note, &m.UserExamStatus, &m.Status,
		&m.CreateId, &m.CreateDt, &m.ModifyId, &m.ModifyDt); err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *UserExamRepository) findOneBy(ctx context.Context, where string, args ...any) (*exam.UserExam, error) {
	fullArgs := append(userExamActiveArgs(), args...)
	query := `SELECT ` + userExamColumns + ` FROM ` + userExamTable + ` e WHERE ` +
		userExamActiveWhere + ` AND (` + where + `)`

	m, err := scanUserExam(r.db.QueryRow(ctx, query, fullArgs...))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("user exam repo find (%s): %w", where, err)
	}
	return ModelToDomainUserExam(m), nil
}

// FindByUserExamId reads one row of a journey.
func (r *UserExamRepository) FindByUserExamId(ctx context.Context, userExamId int64) (*exam.UserExam, error) {
	return r.findOneBy(ctx, "e.user_exam_id = ?", userExamId)
}

// FindByUserExamIdAndType reads one row of a journey. user_exam_id alone
// is not a key here — the ASSESSMENT row and the PRACTICE row of one
// journey share it — so the type is part of every by-id read.
func (r *UserExamRepository) FindByUserExamIdAndType(ctx context.Context, userExamId int64, examType string) (*exam.UserExam, error) {
	return r.findOneBy(ctx, "e.user_exam_id = ? AND e.req_exam_type = ?", userExamId, examType)
}

// FindActiveByUserProfileType reads the open journey. uk_active_journey
// guarantees there is at most one, so no ORDER BY is needed to pick.
func (r *UserExamRepository) FindActiveByUserProfileType(ctx context.Context, userId, profileId int64, examType string) (*exam.UserExam, error) {
	return r.findOneBy(ctx,
		"e.user_id = ? AND e.profile_id = ? AND e.req_exam_type = ? AND e.user_exam_status = ?",
		userId, profileId, examType, string(enum.UserExamStatusActive))
}

// FindLatestCompletedByUserProfileType reads the journey a new one
// inherits from. Ordered by ended_dt, not create_dt: the row that closed
// most recently is the freshest measurement, whichever opened first.
func (r *UserExamRepository) FindLatestCompletedByUserProfileType(ctx context.Context, userId, profileId int64, examType string) (*exam.UserExam, error) {
	args := append(userExamActiveArgs(), userId, profileId, examType, string(enum.UserExamStatusComplete))
	query := `SELECT ` + userExamColumns + ` FROM ` + userExamTable + ` e WHERE ` +
		userExamActiveWhere +
		` AND e.user_id = ? AND e.profile_id = ? AND e.req_exam_type = ? AND e.user_exam_status = ?` +
		` ORDER BY e.ended_dt DESC, e.id DESC LIMIT 1`

	m, err := scanUserExam(r.db.QueryRow(ctx, query, args...))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("user exam repo find latest completed: %w", err)
	}
	return ModelToDomainUserExam(m), nil
}

// ListByUserProfile returns a child's journeys, newest first inside each
// exam type, so the open journey (if any) is always the first row of its
// group and the ended ones follow as history.
func (r *UserExamRepository) ListByUserProfile(ctx context.Context, userId, profileId int64, filter exam.ListJourneysFilter) ([]*exam.UserExam, error) {
	where := userExamActiveWhere + ` AND e.user_id = ? AND e.profile_id = ?`
	args := append(userExamActiveArgs(), userId, profileId)

	if filter.ExamType != nil && *filter.ExamType != "" {
		where += ` AND e.req_exam_type = ?`
		args = append(args, *filter.ExamType)
	}
	if filter.Status != nil && *filter.Status != "" {
		where += ` AND e.user_exam_status = ?`
		args = append(args, *filter.Status)
	}

	query := `SELECT ` + userExamColumns + ` FROM ` + userExamTable + ` e WHERE ` + where +
		` ORDER BY e.create_dt DESC, e.id DESC`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("user exam repo list by user/profile: %w", err)
	}
	defer rows.Close()

	var out []*exam.UserExam
	for rows.Next() {
		m, err := scanUserExam(rows)
		if err != nil {
			return nil, fmt.Errorf("user exam repo scan row: %w", err)
		}
		out = append(out, ModelToDomainUserExam(m))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("user exam repo rows iteration: %w", err)
	}
	return out, nil
}

// MarkStatus ends a journey. The WHERE clause carries ACTIVE as the
// expected state: two marks racing on the same journey both read ACTIVE,
// and this is what makes the loser update zero rows and get
// exam.ErrJourneyNotActive rather than silently "ending" it twice with a
// different status. A PRACTICE row is never ACTIVE and so is never hit.
//
// Flipping user_exam_status also flips the generated active_key to NULL,
// which is what frees the (user, profile, type) slots for the next journey.
func (r *UserExamRepository) MarkStatus(ctx context.Context, userExamId int64, newStatus string, endedDt mtime.MathTime) error {
	query := `
		UPDATE ` + userExamTable + `
		SET user_exam_status = ?,
			ended_dt         = ?,
			modify_dt        = ?
		WHERE user_exam_id = ? AND user_exam_status = ?
	`
	ended := endedDt.Time
	if endedDt.IsZero() {
		ended = mtime.Now().Time
	}

	result, err := r.db.Exec(ctx, query,
		newStatus, ended, mtime.Now().Time,
		userExamId, string(enum.UserExamStatusActive))
	if err != nil {
		return fmt.Errorf("user exam repo mark status: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("user exam repo mark status rows affected: %w", err)
	}
	if affected == 0 {
		return exam.ErrJourneyNotActive
	}
	return nil
}

// Reopen puts an ended journey back in play and clears ended_dt so the
// row reads as open again. The PRACTICE row is left as it is: it is born
// COMPLETE and must not be dragged to ACTIVE, where it would contend for
// the (user, profile, PRACTICE) slot and read as a journey of its own.
//
// Two guards. The WHERE clause carries the ended states, so a reopen
// racing a mark matches zero rows and gets ErrJourneyNotEnded rather than
// re-opening something that just changed. And flipping user_exam_status
// regenerates active_key, so if another journey of the type is open the
// UPDATE trips uk_active_journey and comes back as ErrJourneyConflict —
// the database, not the caller's earlier read, is what holds "one open
// journey at a time".
func (r *UserExamRepository) Reopen(ctx context.Context, userExamId int64) error {
	query := `
		UPDATE ` + userExamTable + `
		SET user_exam_status = ?,
			ended_dt         = NULL,
			modify_dt        = ?
		WHERE user_exam_id = ? AND req_exam_type <> ? AND user_exam_status IN (?, ?)
	`
	result, err := r.db.Exec(ctx, query,
		string(enum.UserExamStatusActive), mtime.Now().Time,
		userExamId, string(enum.ExamTypePractice), string(enum.UserExamStatusComplete), string(enum.UserExamStatusCancel))
	if err != nil {
		if isDuplicateEntry(err) {
			return exam.ErrJourneyConflict
		}
		return fmt.Errorf("user exam repo reopen: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("user exam repo reopen rows affected: %w", err)
	}
	if affected == 0 {
		return exam.ErrJourneyNotEnded
	}
	return nil
}

// SetCurrent records where the child is working, as the client stated it
// on a hand-out. Each value is COALESCEd: a nil leaves the column as it
// is, so a request that names only the grade does not blank the level.
// Only the open row of the journey's own type is touched — a PRACTICE
// row has no placement of its own.
func (r *UserExamRepository) SetCurrent(ctx context.Context, userExamId int64, examType string, grade, level *int) error {
	if grade == nil && level == nil {
		return nil
	}
	query := `
		UPDATE ` + userExamTable + `
		SET current_grade = COALESCE(?, current_grade),
			current_level = COALESCE(?, current_level),
			modify_dt     = ?
		WHERE user_exam_id = ? AND req_exam_type = ? AND user_exam_status = ?
	`
	result, err := r.db.Exec(ctx, query, grade, level, mtime.Now().Time,
		userExamId, examType, string(enum.UserExamStatusActive))
	if err != nil {
		return fmt.Errorf("user exam repo set current: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("user exam repo set current rows affected: %w", err)
	}
	if affected == 0 {
		return exam.ErrJourneyNotActive
	}
	return nil
}

// Create opens a journey row. It is a plain INSERT on purpose: the
// previous INSERT ... ON DUPLICATE KEY UPDATE quietly redirected the write
// into whatever row collided on ANY unique key — and when the schema
// drifted and the triple key came back without its active_key column,
// that row was an already-ended journey. Now a collision is reported, not
// absorbed.
//
// Two unique keys back this. uk_active_journey makes "at most one open
// row per (user, profile, type)" hold under concurrency: two first-ever
// submits both try to INSERT, one wins, the other gets ErrJourneyConflict
// and folds into the winner. uk_journey_type makes (user_exam_id, type)
// unique, which is what lets a PRACTICE row reuse its journey's id
// without ever doubling up. The caller decides the id: a fresh one for an
// ASSESSMENT row, the journey's own for a PRACTICE row.
func (r *UserExamRepository) Create(ctx context.Context, e *exam.UserExam, delta exam.StatsDelta) error {
	query := `
		INSERT INTO ` + userExamTable + `
			(user_exam_id, user_id, profile_id, req_exam_type,
			 res_total_questions, res_correct_number, res_skipped_number, res_score_percentage,
			 res_review, current_grade, current_level, last_submitted_dt,
			 user_exam_status, create_id, create_dt, modify_dt)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	var percentage *int
	if delta.TotalQuestions > 0 {
		// Same rounding as the ROUND() in Accumulate, so the first fold and
		// every later one agree on the percentage.
		p := int(math.Round(float64(delta.CorrectNumber) / float64(delta.TotalQuestions) * 100))
		percentage = &p
	}

	now := mtime.Now().Time
	lastSubmitted := nullableTime(e.LastSubmittedDt())

	// A journey's own row opens ACTIVE; a PRACTICE row is born in its
	// journey's state — COMPLETE — and the caller says which by setting it.
	rowStatus := string(enum.UserExamStatusActive)
	if s := e.UserExamStatus(); s != nil && *s != "" {
		rowStatus = *s
	}

	if _, err := r.db.Exec(ctx, query,
		e.UserExamId(), e.UserId(), e.ProfileId(), e.ReqExamType(),
		delta.TotalQuestions, delta.CorrectNumber, delta.SkippedNumber, percentage,
		e.ResReview(), e.CurrentGrade(), e.CurrentLevel(), lastSubmitted,
		rowStatus, e.CreateId(), now, now); err != nil {
		if isDuplicateEntry(err) {
			return exam.ErrJourneyConflict
		}
		return fmt.Errorf("user exam repo create: %w", err)
	}
	return nil
}

// Accumulate folds a sitting into a journey row. Counters ADD; the
// review OVERWRITES. current_grade / current_level are NOT touched: they
// are what the client stated at hand-out (see SetCurrent), not something
// a result moves.
//
// The percentage is the subtle assignment: it must come from the NEW
// totals, so it is computed from the two columns updated just above it.
// MySQL evaluates SET assignments left to right, which is what lets line
// four read what lines one and two wrote — reorder them and the
// percentage silently lags one submission.
//
// The WHERE clause carries the status the caller expects. A row whose
// state moved between the caller's read and this write matches zero rows
// and gets ErrJourneyNotActive, rather than having a sitting folded into
// the wrong place.
func (r *UserExamRepository) Accumulate(ctx context.Context, userExamId int64, examType, expectedStatus string, e *exam.UserExam, delta exam.StatsDelta) error {
	query := `
		UPDATE ` + userExamTable + `
		SET res_total_questions  = res_total_questions + ?,
			res_correct_number   = res_correct_number  + ?,
			res_skipped_number   = res_skipped_number  + ?,
			res_score_percentage = IF(res_total_questions > 0,
				ROUND(res_correct_number * 100 / res_total_questions), NULL),
			res_review           = ?,
			last_submitted_dt    = ?,
			modify_dt            = ?
		WHERE user_exam_id = ? AND req_exam_type = ? AND user_exam_status = ?
	`
	lastSubmitted := nullableTime(e.LastSubmittedDt())

	result, err := r.db.Exec(ctx, query,
		delta.TotalQuestions, delta.CorrectNumber, delta.SkippedNumber,
		e.ResReview(), lastSubmitted, mtime.Now().Time,
		userExamId, examType, expectedStatus)
	if err != nil {
		return fmt.Errorf("user exam repo accumulate: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("user exam repo accumulate rows affected: %w", err)
	}
	if affected == 0 {
		return exam.ErrJourneyNotActive
	}
	return nil
}

// nullableTime maps the zero MathTime to SQL NULL. A journey opened at
// hand-out has never been submitted to, and its last_submitted_dt must
// read as "never", not as year one.
func nullableTime(mt mtime.MathTime) *time.Time {
	if mt.IsZero() {
		return nil
	}
	return &mt.Time
}

func ModelToDomainUserExam(m *models.UserExamModel) *exam.UserExam {
	e := exam.NewUserExam()
	e.SetId(m.Id)
	e.SetUserExamId(m.UserExamId)
	e.SetUserId(m.UserId)
	e.SetProfileId(m.ProfileId)
	e.SetReqExamType(m.ReqExamType)
	e.SetResTotalQuestions(m.ResTotalQuestions)
	e.SetResCorrectNumber(m.ResCorrectNumber)
	e.SetResSkippedNumber(m.ResSkippedNumber)
	e.SetResScorePercentage(m.ResScorePercentage)
	e.SetResReview(m.ResReview)
	e.SetCurrentGrade(m.CurrentGrade)
	e.SetCurrentLevel(m.CurrentLevel)
	e.SetLastSubmittedDt(mtime.MathTimeFromPtr(m.LastSubmittedDt))
	e.SetEndedDt(mtime.MathTimeFromPtr(m.EndedDt))
	e.SetNote(m.Note)
	e.SetUserExamStatus(m.UserExamStatus)
	e.SetStatus(m.Status)
	e.SetCreateId(m.CreateId)
	e.SetCreateDt(mtime.MathTime{Time: m.CreateDt})
	e.SetModifyId(m.ModifyId)
	e.SetModifyDt(mtime.MathTime{Time: m.ModifyDt})
	return e
}
