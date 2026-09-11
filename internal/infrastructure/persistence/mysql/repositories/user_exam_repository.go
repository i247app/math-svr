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
		e.res_review, e.res_grade, e.res_level, e.last_submitted_dt,
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
		&m.ResReview, &m.ResGrade, &m.ResLevel, &m.LastSubmittedDt,
		&m.Note, &m.UserExamStatus, &m.Status,
		&m.CreateId, &m.CreateDt, &m.ModifyId, &m.ModifyDt); err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *UserExamRepository) FindByUserProfileType(ctx context.Context, userId, profileId int64, examType string) (*exam.UserExam, error) {
	args := append(userExamActiveArgs(), userId, profileId, examType)
	query := `SELECT ` + userExamColumns + ` FROM ` + userExamTable + ` e WHERE ` +
		userExamActiveWhere + ` AND e.user_id = ? AND e.profile_id = ? AND e.req_exam_type = ?`

	m, err := scanUserExam(r.db.QueryRow(ctx, query, args...))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("user exam repo find by user/profile/type: %w", err)
	}
	return ModelToDomainUserExam(m), nil
}

// ListByUserProfile returns every exam type's lifetime row for one child.
// The result is at most three rows, so it is deliberately unpaginated.
func (r *UserExamRepository) ListByUserProfile(ctx context.Context, userId, profileId int64) ([]*exam.UserExam, error) {
	args := append(userExamActiveArgs(), userId, profileId)
	query := `SELECT ` + userExamColumns + ` FROM ` + userExamTable + ` e WHERE ` +
		userExamActiveWhere + ` AND e.user_id = ? AND e.profile_id = ? ORDER BY e.req_exam_type ASC`

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

// Upsert creates the lifetime row for a (user, profile, exam type) triple
// or accumulates onto the existing one. It leans on the table's
// uk_user_profile_type key, so two submits racing for the same child
// serialise on that unique index instead of both inserting.
//
// Counters ADD, placement fields OVERWRITE. The percentage is the subtle
// one: it must be recomputed from the NEW totals, not accumulated, so the
// statement reassigns it from the two columns it has just updated. MySQL
// evaluates ON DUPLICATE KEY UPDATE assignments left to right, which is
// what makes reading res_correct_number and res_total_questions on the
// fourth line see the values written on lines one and two. Reorder those
// lines and the percentage silently goes stale by one submission.
//
// VALUES() is deprecated in MySQL 8.0.20+ in favour of a row alias, but it
// is kept here because it works on every 8.x the project may meet.
func (r *UserExamRepository) Upsert(ctx context.Context, e *exam.UserExam, delta exam.StatsDelta) error {
	query := `
		INSERT INTO ` + userExamTable + `
			(user_exam_id, user_id, profile_id, req_exam_type,
			 res_total_questions, res_correct_number, res_skipped_number, res_score_percentage,
			 res_review, res_grade, res_level, last_submitted_dt,
			 user_exam_status, create_id, create_dt, modify_dt)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			res_total_questions  = res_total_questions + VALUES(res_total_questions),
			res_correct_number   = res_correct_number  + VALUES(res_correct_number),
			res_skipped_number   = res_skipped_number  + VALUES(res_skipped_number),
			res_score_percentage = IF(res_total_questions > 0,
				ROUND(res_correct_number * 100 / res_total_questions), NULL),
			res_review           = VALUES(res_review),
			res_grade            = VALUES(res_grade),
			res_level            = VALUES(res_level),
			last_submitted_dt    = VALUES(last_submitted_dt),
			modify_dt            = VALUES(modify_dt)
	`

	// The percentage supplied on the INSERT path only applies when this is
	// the first submission for the triple; the UPDATE path recomputes it.
	var firstPercentage *int
	if delta.TotalQuestions > 0 {
		p := int(float64(delta.CorrectNumber)/float64(delta.TotalQuestions)*100 + 0.5)
		firstPercentage = &p
	}

	status := e.UserExamStatus()
	if status == nil {
		s := string(enum.UserExamStatusActive)
		status = &s
	}
	now := mtime.Now().Time
	lastSubmitted := mtime.MathTimePtrToTime(e.LastSubmittedDt().Ptr())

	if _, err := r.db.Exec(ctx, query,
		e.UserExamId(), e.UserId(), e.ProfileId(), e.ReqExamType(),
		delta.TotalQuestions, delta.CorrectNumber, delta.SkippedNumber, firstPercentage,
		e.ResReview(), e.ResGrade(), e.ResLevel(), lastSubmitted,
		status, e.CreateId(), now, now); err != nil {
		return fmt.Errorf("user exam repo upsert: %w", err)
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
	e.SetNote(m.Note)
	e.SetUserExamStatus(m.UserExamStatus)
	e.SetStatus(m.Status)
	e.SetCreateId(m.CreateId)
	e.SetCreateDt(mtime.MathTime{Time: m.CreateDt})
	e.SetModifyId(m.ModifyId)
	e.SetModifyDt(mtime.MathTime{Time: m.ModifyDt})
	return e
}
