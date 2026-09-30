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
	examLinkTable = "ma_exam_links"

	examLinkColumns = `u.elink_id, u.uid, u.profile_id, u.exam_id, u.esess_id, u.shuffle_map,
		u.req_exam_type, u.req_grade, u.req_level,
		u.res_total_questions, u.res_correct_number, u.res_skipped_number, u.res_score_percentage,
		u.started_dt, u.submitted_dt,
		u.rpt_flg, u.kwords, u.note, u.elink_status, u.status,
		u.create_id, u.create_dt, u.modify_id, u.modify_dt`

	examLinkActiveWhere = `u.status IN (?) AND u.deleted_dt IS NULL`
)

func examLinkActiveArgs() []any {
	return []any{enum.StatusActive}
}

type ExamLinkRepository struct {
	db database.Executor
}

func NewExamLinkRepository(db database.Executor) exam.IExamLinkRepository {
	return &ExamLinkRepository{db: db}
}

func scanExamLink(s database.RowScanner) (*models.ExamLinkModel, error) {
	var m models.ExamLinkModel
	if err := s.Scan(&m.ElinkId, &m.Uid, &m.ProfileId, &m.ExamId, &m.EsessId, &m.ShuffleMap,
		&m.ReqExamType, &m.ReqGrade, &m.ReqLevel,
		&m.ResTotalQuestions, &m.ResCorrectNumber, &m.ResSkippedNumber, &m.ResScorePercentage,
		&m.StartedDt, &m.SubmittedDt,
		&m.RptFlg, &m.Kwords, &m.Note, &m.ElinkStatus, &m.Status,
		&m.CreateId, &m.CreateDt, &m.ModifyId, &m.ModifyDt); err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *ExamLinkRepository) findOneBy(ctx context.Context, where string, args ...any) (*exam.ExamLink, error) {
	fullArgs := slices.Concat(args, examLinkActiveArgs())
	query := `SELECT ` + examLinkColumns + ` FROM ` + examLinkTable + ` u WHERE (` +
		where + `) AND ` + examLinkActiveWhere

	m, err := scanExamLink(r.db.QueryRow(ctx, query, fullArgs...))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("user ai exam repo find (%s): %w", where, err)
	}
	return ModelToDomainExamLink(m), nil
}

func (r *ExamLinkRepository) FindByElinkId(ctx context.Context, elinkId int64) (*exam.ExamLink, error) {
	return r.findOneBy(ctx, "u.elink_id = ?", elinkId)
}

// ListAttempts returns one child's attempt history, newest first. Rows in
// every state are returned, including IN_PROGRESS ones: an unfinished exam
// is exactly what the "you left this open" surface needs, and the caller
// filters by status when it wants only those.
func (r *ExamLinkRepository) ListAttempts(ctx context.Context, filter exam.ListAttemptsFilter, page, limit int64) ([]*exam.ExamLink, *pagination.Pagination, error) {
	if limit <= 0 {
		limit = pagination.DefaultPageSize
	}
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * limit

	filterWhere, filterArgs := buildExamLinkFilterClause(filter)

	countArgs := slices.Concat(filterArgs, examLinkActiveArgs())
	countQuery := `SELECT COUNT(*) FROM ` + examLinkTable + ` u` +
		whereActive(filterWhere, examLinkActiveWhere)

	var total int64
	if err := r.db.QueryRow(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, nil, fmt.Errorf("user ai exam repo count: %w", err)
	}

	args := slices.Concat(filterArgs, examLinkActiveArgs(), []any{limit, offset})
	query := `SELECT ` + examLinkColumns + ` FROM ` + examLinkTable + ` u` +
		whereActive(filterWhere, examLinkActiveWhere) + ` ORDER BY u.elink_id DESC LIMIT ? OFFSET ?`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, nil, fmt.Errorf("user ai exam repo list: %w", err)
	}
	defer rows.Close()

	var attempts []*exam.ExamLink
	for rows.Next() {
		m, err := scanExamLink(rows)
		if err != nil {
			return nil, nil, fmt.Errorf("user ai exam repo scan row: %w", err)
		}
		attempts = append(attempts, ModelToDomainExamLink(m))
	}
	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("user ai exam repo rows iteration: %w", err)
	}
	return attempts, pagination.NewPagination(page, limit, total), nil
}

// buildExamLinkFilterClause keeps placeholder order in lockstep with the
// returned args slice.
func buildExamLinkFilterClause(filter exam.ListAttemptsFilter) (string, []any) {
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
		clause += ` AND u.elink_status = ?`
		args = append(args, *filter.Status)
	}
	if filter.EsessID != nil && *filter.EsessID != 0 {
		clause += ` AND u.esess_id = ?`
		args = append(args, *filter.EsessID)
	}
	return clause, args
}

// ListInProgressByProfile reads ix_profile_status_started: one child, one
// status, in the order the sittings were handed out.
// CountHandedOutSince counts what this child was handed since `since`.
// started_dt is the hand-out moment, so an exam opened and walked away
// from counts the same as one that was finished — the model call was
// made either way, which is what the ceiling is protecting.
// ReassignOwnerByProfile re-points every sitting of one child at another
// account. Addressed by profile_id because the child is what moves; the
// rows themselves are unchanged apart from who owns them.
func (r *ExamLinkRepository) ReassignOwnerByProfile(ctx context.Context, profileId int64, newUid int64) error {
	query := `UPDATE ` + examLinkTable + ` SET uid = ?, modify_dt = ? WHERE profile_id = ?`
	if _, err := r.db.Exec(ctx, query, newUid, mtime.Now().Time, profileId); err != nil {
		return fmt.Errorf("user ai exam repo reassign owner: %w", err)
	}
	return nil
}

func (r *ExamLinkRepository) CountHandedOutSince(ctx context.Context, profileId int64, since mtime.MathTime) (int64, error) {
	args := slices.Concat([]any{profileId, since.Time}, examLinkActiveArgs())
	query := `SELECT COUNT(*) FROM ` + examLinkTable + ` u WHERE ` +
		`(u.profile_id = ? AND u.started_dt >= ?) AND ` + examLinkActiveWhere

	var total int64
	if err := r.db.QueryRow(ctx, query, args...).Scan(&total); err != nil {
		return 0, fmt.Errorf("user ai exam repo count handed out: %w", err)
	}
	return total, nil
}

func (r *ExamLinkRepository) ListInProgressByProfile(ctx context.Context, profileId int64) ([]*exam.ExamLink, error) {
	args := slices.Concat([]any{profileId, string(enum.ElinkStatusInProgress)}, examLinkActiveArgs())
	query := `SELECT ` + examLinkColumns + ` FROM ` + examLinkTable + ` u WHERE ` +
		`(u.profile_id = ? AND u.elink_status = ?) AND ` + examLinkActiveWhere +
		` ORDER BY u.started_dt ASC, u.elink_id ASC`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("user ai exam repo list in progress: %w", err)
	}
	defer rows.Close()

	var out []*exam.ExamLink
	for rows.Next() {
		m, err := scanExamLink(rows)
		if err != nil {
			return nil, fmt.Errorf("user ai exam repo scan row: %w", err)
		}
		out = append(out, ModelToDomainExamLink(m))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("user ai exam repo rows iteration: %w", err)
	}
	return out, nil
}

// FindLatestSubmittedByEsessId walks ix_esess_submitted backwards:
// the newest submitted_dt of the journey, id as the tie-break so two
// sittings submitted in the same microsecond still order deterministically.
func (r *ExamLinkRepository) FindLatestSubmittedByEsessId(ctx context.Context, esessId int64) (*exam.ExamLink, error) {
	args := slices.Concat([]any{esessId, string(enum.ElinkStatusSubmitted)}, examLinkActiveArgs())
	query := `SELECT ` + examLinkColumns + ` FROM ` + examLinkTable + ` u WHERE ` +
		`(u.esess_id = ? AND u.elink_status = ?) AND ` + examLinkActiveWhere +
		` ORDER BY u.submitted_dt DESC, u.elink_id DESC LIMIT 1`

	m, err := scanExamLink(r.db.QueryRow(ctx, query, args...))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("user ai exam repo find latest submitted: %w", err)
	}
	return ModelToDomainExamLink(m), nil
}

// FindInProgressInJourney walks ix_profile_status_started (profile_id,
// elink_status) — a child has only a handful of open sittings — and
// filters on the journey and type.
func (r *ExamLinkRepository) FindInProgressInJourney(ctx context.Context, profileId, esessId int64, examType string) (*exam.ExamLink, error) {
	args := slices.Concat(
		[]any{profileId, string(enum.ElinkStatusInProgress), esessId, examType},
		examLinkActiveArgs())
	query := `SELECT ` + examLinkColumns + ` FROM ` + examLinkTable + ` u WHERE ` +
		`(u.profile_id = ? AND u.elink_status = ? AND u.esess_id = ? AND u.req_exam_type = ?) AND ` +
		examLinkActiveWhere + ` ORDER BY u.started_dt DESC, u.elink_id DESC LIMIT 1`

	m, err := scanExamLink(r.db.QueryRow(ctx, query, args...))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("user ai exam repo find in progress in journey: %w", err)
	}
	return ModelToDomainExamLink(m), nil
}

// ListRecentByProfileGrade reads a child's newest sittings at one grade.
// It walks ix_profile_status_started's profile prefix and filters on
// req_grade; the candidate set per child is small enough that a
// dedicated index is not worth its write cost.
func (r *ExamLinkRepository) ListRecentByProfileGrade(ctx context.Context, profileId int64, grade int, limit int) ([]*exam.ExamLink, error) {
	if limit <= 0 {
		return nil, nil
	}
	args := slices.Concat([]any{profileId, grade}, examLinkActiveArgs(), []any{limit})
	query := `SELECT ` + examLinkColumns + ` FROM ` + examLinkTable + ` u WHERE ` +
		`(u.profile_id = ? AND u.req_grade = ?) AND ` + examLinkActiveWhere +
		` ORDER BY u.started_dt DESC, u.elink_id DESC LIMIT ?`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("user ai exam repo list recent by grade: %w", err)
	}
	defer rows.Close()

	var out []*exam.ExamLink
	for rows.Next() {
		m, err := scanExamLink(rows)
		if err != nil {
			return nil, fmt.Errorf("user ai exam repo scan row: %w", err)
		}
		out = append(out, ModelToDomainExamLink(m))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("user ai exam repo rows iteration: %w", err)
	}
	return out, nil
}

// ListByElinkIds hydrates a batch of attempts, oldest first. The IN
// list is built from the id count so the query stays parameterised.
func (r *ExamLinkRepository) ListByElinkIds(ctx context.Context, elinkIds []int64) ([]*exam.ExamLink, error) {
	if len(elinkIds) == 0 {
		return nil, nil
	}

	placeholders := make([]string, 0, len(elinkIds))
	args := make([]any, 0, len(elinkIds)+1)
	for _, id := range elinkIds {
		placeholders = append(placeholders, "?")
		args = append(args, id)
	}
	args = append(args, examLinkActiveArgs()...)

	query := `SELECT ` + examLinkColumns + ` FROM ` + examLinkTable + ` u WHERE (u.elink_id IN (` +
		strings.Join(placeholders, ", ") + `)) AND ` + examLinkActiveWhere +
		` ORDER BY u.elink_id ASC`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("user ai exam repo list by ids: %w", err)
	}
	defer rows.Close()

	var out []*exam.ExamLink
	for rows.Next() {
		m, err := scanExamLink(rows)
		if err != nil {
			return nil, fmt.Errorf("user ai exam repo scan row: %w", err)
		}
		out = append(out, ModelToDomainExamLink(m))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("user ai exam repo rows iteration: %w", err)
	}
	return out, nil
}

func (r *ExamLinkRepository) Create(ctx context.Context, a *exam.ExamLink) (*exam.ExamLink, error) {
	query := `
		INSERT INTO ` + examLinkTable + `
			(elink_id, uid, profile_id, exam_id, esess_id, shuffle_map,
			 req_exam_type, req_grade, req_level,
			 started_dt, rpt_flg, kwords, note, elink_status, create_id, create_dt, modify_dt)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	now := mtime.Now().Time
	startedDt := mtime.MathTimePtrToTime(a.StartedDt().Ptr())

	_, err := r.db.Exec(ctx, query,
		a.ElinkId(), a.Uid(), a.ProfileId(), a.ExamId(), a.EsessId(), a.ShuffleMap(),
		a.ReqExamType(), a.ReqGrade(), a.ReqLevel(),
		startedDt, a.RptFlg(), a.Kwords(), a.Note(), a.ElinkStatus(), a.CreateId(), now, now)
	if err != nil {
		return nil, fmt.Errorf("user ai exam repo create: %w", err)
	}

	return r.FindByElinkId(ctx, a.ElinkId())
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
// esess_id is COALESCEd: a nil result leaves what the hand-out wrote,
// a value pins the sitting to the journey it actually folded into.
func (r *ExamLinkRepository) MarkSubmitted(ctx context.Context, elinkId int64, res exam.AttemptResult) error {
	query := `
		UPDATE ` + examLinkTable + `
		SET res_total_questions  = ?,
			res_correct_number   = ?,
			res_skipped_number   = ?,
			res_score_percentage = ?,
			submitted_dt         = ?,
			esess_id         = COALESCE(?, esess_id),
			elink_status  = ?,
			modify_dt            = ?
		WHERE elink_id = ? AND elink_status = ?
	`
	submittedDt := res.SubmittedDt.Time
	if res.SubmittedDt.IsZero() {
		submittedDt = mtime.Now().Time
	}

	result, err := r.db.Exec(ctx, query,
		res.TotalQuestions, res.CorrectNumber, res.SkippedNumber, res.ScorePercentage,
		submittedDt, res.EsessId, string(enum.ElinkStatusSubmitted), mtime.Now().Time,
		elinkId, string(enum.ElinkStatusInProgress))
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
const progressPointColumnsExam = `u.elink_id, u.exam_id, u.req_exam_type, u.req_grade,
	u.res_score_percentage, u.res_correct_number, u.res_total_questions, u.submitted_dt`

func scanExamProgressPoint(s database.RowScanner) (*exam.ProgressPoint, error) {
	var (
		p           exam.ProgressPoint
		correct     *int64
		total       *int64
		submittedDt *time.Time
	)
	if err := s.Scan(&p.ElinkId, &p.ExamId, &p.ExamType, &p.Grade,
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
func (r *ExamLinkRepository) ListProgressPoints(ctx context.Context, params exam.ProgressPointsParams) ([]*exam.ProgressPoint, error) {
	where := `u.elink_status = ? AND u.profile_id = ? AND u.res_score_percentage IS NOT NULL`
	args := []any{string(enum.ElinkStatusSubmitted), params.ProfileID}

	if params.ExamType != nil && *params.ExamType != "" {
		where += ` AND u.req_exam_type = ?`
		args = append(args, *params.ExamType)
	}
	if params.EsessID != nil && *params.EsessID != 0 {
		where += ` AND u.esess_id = ?`
		args = append(args, *params.EsessID)
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
	args = append(args, examLinkActiveArgs()...)
	args = append(args, limit)

	query := `SELECT ` + progressPointColumnsExam + ` FROM ` + examLinkTable + ` u WHERE (` +
		where + `) AND ` + examLinkActiveWhere + ` ORDER BY u.submitted_dt DESC, u.elink_id DESC LIMIT ?`

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

func ModelToDomainExamLink(m *models.ExamLinkModel) *exam.ExamLink {
	a := exam.NewExamLink()
	a.SetElinkId(m.ElinkId)
	a.SetUid(m.Uid)
	a.SetProfileId(m.ProfileId)
	a.SetExamId(m.ExamId)
	a.SetEsessId(m.EsessId)
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
	a.SetRptFlg(m.RptFlg)
	a.SetKwords(m.Kwords)
	a.SetNote(m.Note)
	a.SetElinkStatus(m.ElinkStatus)
	a.SetStatus(m.Status)
	a.SetCreateId(m.CreateId)
	a.SetCreateDt(mtime.MathTime{Time: m.CreateDt})
	a.SetModifyId(m.ModifyId)
	a.SetModifyDt(mtime.MathTime{Time: m.ModifyDt})
	return a
}
