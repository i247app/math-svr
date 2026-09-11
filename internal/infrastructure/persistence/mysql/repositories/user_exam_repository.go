package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

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
		e.res_review, e.res_grade, e.res_level, e.last_submitted_dt, e.ended_dt,
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
		&m.ResReview, &m.ResGrade, &m.ResLevel, &m.LastSubmittedDt, &m.EndedDt,
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

func (r *UserExamRepository) FindByUserExamId(ctx context.Context, userExamId int64) (*exam.UserExam, error) {
	return r.findOneBy(ctx, "e.user_exam_id = ?", userExamId)
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
		` ORDER BY e.req_exam_type ASC, e.create_dt DESC, e.id DESC`

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
// expected state: two marks racing on the same row both read ACTIVE, and
// this is what makes the loser update zero rows and get
// exam.ErrJourneyNotActive rather than silently "ending" it twice with a
// different status.
//
// Flipping user_exam_status also flips the generated active_key to NULL,
// which is what frees the (user, profile, type) slot for the next journey.
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

// Create opens a journey. It is a plain INSERT on purpose: the previous
// INSERT ... ON DUPLICATE KEY UPDATE quietly redirected the write into
// whatever row collided on ANY unique key — and when the schema drifted
// and the triple key came back without its active_key column, that row
// was an already-ended journey. Now a collision is reported, not absorbed.
//
// uk_active_journey is what makes "at most one open journey per triple"
// hold under concurrency: two first-ever submits both try to INSERT, one
// wins, the other gets ErrJourneyConflict and folds into the winner.
func (r *UserExamRepository) Create(ctx context.Context, e *exam.UserExam, delta exam.StatsDelta) error {
	query := `
		INSERT INTO ` + userExamTable + `
			(user_exam_id, user_id, profile_id, req_exam_type,
			 res_total_questions, res_correct_number, res_skipped_number, res_score_percentage,
			 res_review, res_grade, res_level, last_submitted_dt,
			 user_exam_status, create_id, create_dt, modify_dt)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	var percentage *int
	if delta.TotalQuestions > 0 {
		p := int(float64(delta.CorrectNumber)/float64(delta.TotalQuestions)*100 + 0.5)
		percentage = &p
	}

	now := mtime.Now().Time
	lastSubmitted := mtime.MathTimePtrToTime(e.LastSubmittedDt().Ptr())

	if _, err := r.db.Exec(ctx, query,
		e.UserExamId(), e.UserId(), e.ProfileId(), e.ReqExamType(),
		delta.TotalQuestions, delta.CorrectNumber, delta.SkippedNumber, percentage,
		e.ResReview(), e.ResGrade(), e.ResLevel(), lastSubmitted,
		string(enum.UserExamStatusActive), e.CreateId(), now, now); err != nil {
		if isDuplicateEntry(err) {
			return exam.ErrJourneyConflict
		}
		return fmt.Errorf("user exam repo create: %w", err)
	}
	return nil
}

// Accumulate folds a sitting into an open journey. Counters ADD;
// placement fields OVERWRITE.
//
// The percentage is the subtle assignment: it must come from the NEW
// totals, so it is computed from the two columns updated just above it.
// MySQL evaluates SET assignments left to right, which is what lets line
// four read what lines one and two wrote — reorder them and the
// percentage silently lags one submission.
//
// The WHERE clause carries ACTIVE. A journey that ended between the
// caller's read and this write matches zero rows and gets
// ErrJourneyNotActive, rather than having a sitting folded into history.
func (r *UserExamRepository) Accumulate(ctx context.Context, userExamId int64, e *exam.UserExam, delta exam.StatsDelta) error {
	query := `
		UPDATE ` + userExamTable + `
		SET res_total_questions  = res_total_questions + ?,
			res_correct_number   = res_correct_number  + ?,
			res_skipped_number   = res_skipped_number  + ?,
			res_score_percentage = IF(res_total_questions > 0,
				ROUND(res_correct_number * 100 / res_total_questions), NULL),
			res_review           = ?,
			res_grade            = ?,
			res_level            = ?,
			last_submitted_dt    = ?,
			modify_dt            = ?
		WHERE user_exam_id = ? AND user_exam_status = ?
	`
	lastSubmitted := mtime.MathTimePtrToTime(e.LastSubmittedDt().Ptr())

	result, err := r.db.Exec(ctx, query,
		delta.TotalQuestions, delta.CorrectNumber, delta.SkippedNumber,
		e.ResReview(), e.ResGrade(), e.ResLevel(), lastSubmitted, mtime.Now().Time,
		userExamId, string(enum.UserExamStatusActive))
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
	e.SetResGrade(m.ResGrade)
	e.SetResLevel(m.ResLevel)
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
