package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"slices"
	"time"

	"math-ai.com/math-ai/internal/domain/exam"
	"math-ai.com/math-ai/internal/domain/shared/mtime"
	"math-ai.com/math-ai/internal/infrastructure/database"
	"math-ai.com/math-ai/internal/infrastructure/persistence/mysql/models"
	"math-ai.com/math-ai/internal/shared/enum"
)

const (
	examSessionTable = "ma_exam_sessions"

	examSessionColumns = `e.esess_id, e.uid, e.profile_id, e.req_exam_type,
		e.res_total_questions, e.res_correct_number, e.res_skipped_number, e.res_score_percentage,
		e.res_review, e.current_grade, e.current_level, e.last_submitted_dt, e.ended_dt,
		e.rpt_flg, e.kwords, e.note, e.esess_status, e.status,
		e.create_id, e.create_dt, e.modify_id, e.modify_dt`

	examSessionActiveWhere = `e.status IN (?) AND e.deleted_dt IS NULL`
)

func examSessionActiveArgs() []any {
	return []any{enum.StatusActive}
}

type ExamSessionRepository struct {
	db database.Executor
}

func NewExamSessionRepository(db database.Executor) exam.IExamSessionRepository {
	return &ExamSessionRepository{db: db}
}

func scanExamSession(s database.RowScanner) (*models.ExamSessionModel, error) {
	var m models.ExamSessionModel
	if err := s.Scan(&m.EsessId, &m.Uid, &m.ProfileId, &m.ReqExamType,
		&m.ResTotalQuestions, &m.ResCorrectNumber, &m.ResSkippedNumber, &m.ResScorePercentage,
		&m.ResReview, &m.CurrentGrade, &m.CurrentLevel, &m.LastSubmittedDt, &m.EndedDt,
		&m.RptFlg, &m.Kwords, &m.Note, &m.EsessStatus, &m.Status,
		&m.CreateId, &m.CreateDt, &m.ModifyId, &m.ModifyDt); err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *ExamSessionRepository) findOneBy(ctx context.Context, where string, args ...any) (*exam.ExamSession, error) {
	fullArgs := slices.Concat(args, examSessionActiveArgs())
	query := `SELECT ` + examSessionColumns + ` FROM ` + examSessionTable + ` e WHERE (` +
		where + `) AND ` + examSessionActiveWhere

	m, err := scanExamSession(r.db.QueryRow(ctx, query, fullArgs...))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("user exam repo find (%s): %w", where, err)
	}
	return ModelToDomainExamSession(m), nil
}

// FindByEsessId reads the row that owns the journey's lifecycle — never
// its PRACTICE row, which shares the esess_id (see FindByEsessIdAndType).
// Without the type filter a journey that has been practised could come
// back as its COMPLETE PRACTICE row even while the journey itself is open.
func (r *ExamSessionRepository) FindByEsessId(ctx context.Context, esessId int64) (*exam.ExamSession, error) {
	return r.findOneBy(ctx, "e.esess_id = ? AND e.req_exam_type <> ?", esessId, string(enum.ExamTypePractice))
}

// FindByEsessIdAndType reads one row of a journey. esess_id alone
// is not a key here — the owning row and the PRACTICE row of one
// journey share it — so the type is part of every by-id read.
func (r *ExamSessionRepository) FindByEsessIdAndType(ctx context.Context, esessId int64, examType string) (*exam.ExamSession, error) {
	return r.findOneBy(ctx, "e.esess_id = ? AND e.req_exam_type = ?", esessId, examType)
}

// FindActiveByUserProfileType reads the open journey. uk_active_journey
// guarantees there is at most one, so no ORDER BY is needed to pick.
func (r *ExamSessionRepository) FindActiveByUserProfileType(ctx context.Context, uid, profileId int64, examType string) (*exam.ExamSession, error) {
	return r.findOneBy(ctx,
		"e.uid = ? AND e.profile_id = ? AND e.req_exam_type = ? AND e.esess_status = ?",
		uid, profileId, examType, string(enum.EsessStatusActive))
}

// FindLatestCompletedByUserProfileType reads the journey a new one
// inherits from. Ordered by ended_dt, not create_dt: the row that closed
// most recently is the freshest measurement, whichever opened first.
func (r *ExamSessionRepository) FindLatestCompletedByUserProfileType(ctx context.Context, uid, profileId int64, examType string) (*exam.ExamSession, error) {
	args := slices.Concat([]any{uid, profileId, examType, string(enum.EsessStatusComplete)}, examSessionActiveArgs())
	query := `SELECT ` + examSessionColumns + ` FROM ` + examSessionTable + ` e WHERE ` +
		`(e.uid = ? AND e.profile_id = ? AND e.req_exam_type = ? AND e.esess_status = ?)` +
		` AND ` + examSessionActiveWhere +
		` ORDER BY e.ended_dt DESC, e.esess_id DESC LIMIT 1`

	m, err := scanExamSession(r.db.QueryRow(ctx, query, args...))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("user exam repo find latest completed: %w", err)
	}
	return ModelToDomainExamSession(m), nil
}

// ListByUserProfile returns a child's journeys, newest first inside each
// exam type, so the open journey (if any) is always the first row of its
// group and the ended ones follow as history.
func (r *ExamSessionRepository) ListByUserProfile(ctx context.Context, uid, profileId int64, filter exam.ListJourneysFilter) ([]*exam.ExamSession, error) {
	where := `e.uid = ? AND e.profile_id = ?`
	args := []any{uid, profileId}

	if filter.ExamType != nil && *filter.ExamType != "" {
		where += ` AND e.req_exam_type = ?`
		args = append(args, *filter.ExamType)
	}
	if filter.Status != nil && *filter.Status != "" {
		where += ` AND e.esess_status = ?`
		args = append(args, *filter.Status)
	}

	args = append(args, examSessionActiveArgs()...)
	query := `SELECT ` + examSessionColumns + ` FROM ` + examSessionTable + ` e WHERE (` + where + `) AND ` +
		examSessionActiveWhere + ` ORDER BY e.create_dt DESC, e.esess_id DESC`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("user exam repo list by user/profile: %w", err)
	}
	defer rows.Close()

	var out []*exam.ExamSession
	for rows.Next() {
		m, err := scanExamSession(rows)
		if err != nil {
			return nil, fmt.Errorf("user exam repo scan row: %w", err)
		}
		out = append(out, ModelToDomainExamSession(m))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("user exam repo rows iteration: %w", err)
	}
	return out, nil
}

// ListProgressPoints returns scored journeys for one child, newest
// submission first, capped at params.Limit. The caller reverses into
// chronological order. With no exam type the PRACTICE rows are skipped:
// they share their journey's id and would put every journey on the
// chart twice.
// ReassignOwnerByProfile re-points every journey of one child at another
// account. Addressed by profile_id because the child is what moves; the
// rows themselves are unchanged apart from who owns them.
func (r *ExamSessionRepository) ReassignOwnerByProfile(ctx context.Context, profileId int64, newUid int64) error {
	query := `UPDATE ` + examSessionTable + ` SET uid = ?, modify_dt = ? WHERE profile_id = ?`
	if _, err := r.db.Exec(ctx, query, newUid, mtime.Now().Time, profileId); err != nil {
		return fmt.Errorf("user exam repo reassign owner: %w", err)
	}
	return nil
}

func (r *ExamSessionRepository) ListProgressPoints(ctx context.Context, params exam.JourneyProgressParams) ([]*exam.ExamSession, error) {
	where := `e.uid = ? AND e.profile_id = ? AND e.res_score_percentage IS NOT NULL AND e.last_submitted_dt IS NOT NULL`
	args := []any{params.UID, params.ProfileID}

	if params.ExamType != nil && *params.ExamType != "" {
		where += ` AND e.req_exam_type = ?`
		args = append(args, *params.ExamType)
	} else {
		where += ` AND e.req_exam_type <> ?`
		args = append(args, string(enum.ExamTypePractice))
	}
	if params.From != nil && !params.From.IsZero() {
		where += ` AND e.last_submitted_dt >= ?`
		args = append(args, params.From.Time)
	}
	if params.To != nil && !params.To.IsZero() {
		where += ` AND e.last_submitted_dt <= ?`
		args = append(args, params.To.Time)
	}
	if params.SubmittedBefore != nil && !params.SubmittedBefore.IsZero() {
		where += ` AND e.last_submitted_dt < ?`
		args = append(args, params.SubmittedBefore.Time)
	}

	limit := params.Limit
	if limit <= 0 {
		limit = 10
	}
	args = append(args, examSessionActiveArgs()...)
	args = append(args, limit)

	query := `SELECT ` + examSessionColumns + ` FROM ` + examSessionTable + ` e WHERE (` + where + `) AND ` +
		examSessionActiveWhere + ` ORDER BY e.last_submitted_dt DESC, e.esess_id DESC LIMIT ?`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("user exam repo list progress points: %w", err)
	}
	defer rows.Close()

	var out []*exam.ExamSession
	for rows.Next() {
		m, err := scanExamSession(rows)
		if err != nil {
			return nil, fmt.Errorf("user exam repo scan progress point: %w", err)
		}
		out = append(out, ModelToDomainExamSession(m))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("user exam repo progress points iteration: %w", err)
	}
	return out, nil
}

// MarkStatus ends a journey. The WHERE clause carries ACTIVE as the
// expected state: two marks racing on the same journey both read ACTIVE,
// and this is what makes the loser update zero rows and get
// exam.ErrJourneyNotActive rather than silently "ending" it twice with a
// different status. A PRACTICE row is never ACTIVE and so is never hit.
//
// Moving esess_status off ACTIVE takes the row out of uk_active_journey
// (its IF(esess_status = 'ACTIVE', 1, NULL) key part becomes NULL), which is
// what frees the (user, profile, type) slot for the next journey.
func (r *ExamSessionRepository) MarkStatus(ctx context.Context, esessId int64, newStatus string, endedDt mtime.MathTime) error {
	query := `
		UPDATE ` + examSessionTable + `
		SET esess_status = ?,
			ended_dt         = ?,
			modify_dt        = ?
		WHERE esess_id = ? AND esess_status = ?
	`
	ended := endedDt.Time
	if endedDt.IsZero() {
		ended = mtime.Now().Time
	}

	result, err := r.db.Exec(ctx, query,
		newStatus, ended, mtime.Now().Time,
		esessId, string(enum.EsessStatusActive))
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
// re-opening something that just changed. And moving esess_status back to
// ACTIVE puts the row back into uk_active_journey, so if another journey of
// the type is open the UPDATE trips the key and comes back as ErrJourneyConflict —
// the database, not the caller's earlier read, is what holds "one open
// journey at a time".
func (r *ExamSessionRepository) Reopen(ctx context.Context, esessId int64) error {
	query := `
		UPDATE ` + examSessionTable + `
		SET esess_status = ?,
			ended_dt         = NULL,
			modify_dt        = ?
		WHERE esess_id = ? AND req_exam_type <> ? AND esess_status IN (?, ?)
	`
	result, err := r.db.Exec(ctx, query,
		string(enum.EsessStatusActive), mtime.Now().Time,
		esessId, string(enum.ExamTypePractice), string(enum.EsessStatusComplete), string(enum.EsessStatusCancel))
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
func (r *ExamSessionRepository) SetCurrent(ctx context.Context, esessId int64, examType string, grade, level *int) error {
	if grade == nil && level == nil {
		return nil
	}
	query := `
		UPDATE ` + examSessionTable + `
		SET current_grade = COALESCE(?, current_grade),
			current_level = COALESCE(?, current_level),
			modify_dt     = ?
		WHERE esess_id = ? AND req_exam_type = ? AND esess_status = ?
	`
	result, err := r.db.Exec(ctx, query, grade, level, mtime.Now().Time,
		esessId, examType, string(enum.EsessStatusActive))
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
// drifted and the triple key came back without its ACTIVE-only key part,
// that row was an already-ended journey. Now a collision is reported, not
// absorbed.
//
// Two unique keys back this. uk_active_journey makes "at most one open
// row per (user, profile, type)" hold under concurrency: two first-ever
// submits both try to INSERT, one wins, the other gets ErrJourneyConflict
// and folds into the winner. The PRIMARY KEY is (esess_id, req_exam_type)
// — not esess_id alone (up/035) — which is what lets a PRACTICE row reuse
// its journey's id without ever doubling up. The caller decides the id: a fresh one for an
// ASSESSMENT or GRADE row, the journey's own for a PRACTICE row.
func (r *ExamSessionRepository) Create(ctx context.Context, e *exam.ExamSession, delta exam.StatsDelta) error {
	query := `
		INSERT INTO ` + examSessionTable + `
			(esess_id, uid, profile_id, req_exam_type,
			 res_total_questions, res_correct_number, res_skipped_number, res_score_percentage,
			 res_review, current_grade, current_level, last_submitted_dt,
			 esess_status, create_id, create_dt, modify_dt)
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
	rowStatus := string(enum.EsessStatusActive)
	if s := e.EsessStatus(); s != nil && *s != "" {
		rowStatus = *s
	}

	if _, err := r.db.Exec(ctx, query,
		e.EsessId(), e.Uid(), e.ProfileId(), e.ReqExamType(),
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
func (r *ExamSessionRepository) Accumulate(ctx context.Context, esessId int64, examType, expectedStatus string, e *exam.ExamSession, delta exam.StatsDelta) error {
	query := `
		UPDATE ` + examSessionTable + `
		SET res_total_questions  = res_total_questions + ?,
			res_correct_number   = res_correct_number  + ?,
			res_skipped_number   = res_skipped_number  + ?,
			res_score_percentage = IF(res_total_questions > 0,
				ROUND(res_correct_number * 100 / res_total_questions), NULL),
			res_review           = ?,
			last_submitted_dt    = ?,
			modify_dt            = ?
		WHERE esess_id = ? AND req_exam_type = ? AND esess_status = ?
	`
	lastSubmitted := nullableTime(e.LastSubmittedDt())

	result, err := r.db.Exec(ctx, query,
		delta.TotalQuestions, delta.CorrectNumber, delta.SkippedNumber,
		e.ResReview(), lastSubmitted, mtime.Now().Time,
		esessId, examType, expectedStatus)
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

func ModelToDomainExamSession(m *models.ExamSessionModel) *exam.ExamSession {
	e := exam.NewExamSession()
	e.SetEsessId(m.EsessId)
	e.SetUid(m.Uid)
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
	e.SetRptFlg(m.RptFlg)
	e.SetKwords(m.Kwords)
	e.SetNote(m.Note)
	e.SetEsessStatus(m.EsessStatus)
	e.SetStatus(m.Status)
	e.SetCreateId(m.CreateId)
	e.SetCreateDt(mtime.MathTime{Time: m.CreateDt})
	e.SetModifyId(m.ModifyId)
	e.SetModifyDt(mtime.MathTime{Time: m.ModifyDt})
	return e
}
