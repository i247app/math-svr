package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"slices"
	"strings"

	"math-ai.com/math-ai/internal/domain/exam"
	"math-ai.com/math-ai/internal/domain/shared/mtime"
	"math-ai.com/math-ai/internal/infrastructure/database"
	"math-ai.com/math-ai/internal/infrastructure/persistence/mysql/models"
	"math-ai.com/math-ai/internal/shared/enum"
)

const (
	examPoolTable = "ma_exam_pools"

	examPoolColumns = `a.exam_id, a.req_exam_type, a.req_grade, a.req_level, a.req_num_ques,
		a.req_semester, a.req_program, a.req_extras,
		a.ai_title, a.ai_short_text, a.ai_questions_json,
		a.rpt_flg, a.kwords, a.note, a.exam_status, a.status,
		a.create_id, a.create_dt, a.modify_id, a.modify_dt`

	examPoolActiveWhere = `a.status IN (?) AND a.deleted_dt IS NULL`
)

func examPoolActiveArgs() []any {
	return []any{enum.StatusActive}
}

type ExamPoolRepository struct {
	db database.Executor
}

func NewExamPoolRepository(db database.Executor) exam.IExamPoolRepository {
	return &ExamPoolRepository{db: db}
}

func scanExamPool(s database.RowScanner) (*models.ExamPoolModel, error) {
	var m models.ExamPoolModel
	if err := s.Scan(&m.ExamId, &m.ReqExamType, &m.ReqGrade, &m.ReqLevel, &m.ReqNumQues,
		&m.ReqSemester, &m.ReqProgram, &m.ReqExtras,
		&m.AiTitle, &m.AiShortText, &m.AiQuestionsJson,
		&m.RptFlg, &m.Kwords, &m.Note, &m.ExamStatus, &m.Status,
		&m.CreateId, &m.CreateDt, &m.ModifyId, &m.ModifyDt); err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *ExamPoolRepository) findOneBy(ctx context.Context, where string, args ...any) (*exam.ExamPool, error) {
	fullArgs := slices.Concat(args, examPoolActiveArgs())
	query := `SELECT ` + examPoolColumns + ` FROM ` + examPoolTable + ` a WHERE (` +
		where + `) AND ` + examPoolActiveWhere

	m, err := scanExamPool(r.db.QueryRow(ctx, query, fullArgs...))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("ai exam repo find (%s): %w", where, err)
	}
	return ModelToDomainExamPool(m), nil
}

func (r *ExamPoolRepository) FindByExamId(ctx context.Context, examId int64) (*exam.ExamPool, error) {
	return r.findOneBy(ctx, "a.exam_id = ?", examId)
}

// FindReusableByExtras is the cache read. A miss returns (nil, nil) — the
// caller falls through to a real generation, which is an ordinary outcome
// and not an error.
//
// The NOT EXISTS drops every set the profile already holds an attempt on,
// whatever that attempt's status: an exam handed out and abandoned was
// still seen. When the child has sat every variant under the tag the read
// misses, the caller generates, and the new set joins the pool — so the
// pool grows exactly as fast as its heaviest user needs it to.
//
// ORDER BY RAND() picks among the remaining variants instead of always
// returning the oldest. The cost is acceptable because the index on
// req_extras narrows to one tag's rows before the sort. Revisit if a tag
// ever accumulates thousands of variants.
func (r *ExamPoolRepository) FindReusableByExtras(ctx context.Context, extras string, excludeProfileId int64) (*exam.ExamPool, error) {
	if extras == "" {
		return nil, nil
	}
	args := slices.Concat([]any{extras, excludeProfileId}, examPoolActiveArgs())
	query := `SELECT ` + examPoolColumns + ` FROM ` + examPoolTable + ` a WHERE ` +
		`(a.req_extras = ?` +
		` AND NOT EXISTS (SELECT 1 FROM ` + examLinkTable + ` u WHERE u.exam_id = a.exam_id AND u.profile_id = ?))` +
		` AND ` + examPoolActiveWhere +
		` ORDER BY RAND() LIMIT 1`

	m, err := scanExamPool(r.db.QueryRow(ctx, query, args...))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("ai exam repo find reusable by extras: %w", err)
	}
	return ModelToDomainExamPool(m), nil
}

// CountByExtras counts the reusable question sets under one tag. It is a
// separate round trip rather than a column on some other row because the
// number changes every time a generation lands, and a cached counter that
// drifts would silently stop the pool from ever growing.
func (r *ExamPoolRepository) CountByExtras(ctx context.Context, extras string) (int64, error) {
	if extras == "" {
		return 0, nil
	}
	args := slices.Concat([]any{extras}, examPoolActiveArgs())
	query := `SELECT COUNT(*) FROM ` + examPoolTable + ` a WHERE (a.req_extras = ?) AND ` +
		examPoolActiveWhere

	var total int64
	if err := r.db.QueryRow(ctx, query, args...).Scan(&total); err != nil {
		return 0, fmt.Errorf("ai exam repo count by extras: %w", err)
	}
	return total, nil
}

// ListByExamIds hydrates a batch of question sets. It builds the IN
// list from the id count rather than interpolating values, so the query
// stays parameterised however many ids arrive.
func (r *ExamPoolRepository) ListByExamIds(ctx context.Context, examIds []int64) ([]*exam.ExamPool, error) {
	if len(examIds) == 0 {
		return nil, nil
	}

	placeholders := make([]string, 0, len(examIds))
	args := make([]any, 0, len(examIds)+1)
	for _, id := range examIds {
		placeholders = append(placeholders, "?")
		args = append(args, id)
	}
	args = append(args, examPoolActiveArgs()...)

	query := `SELECT ` + examPoolColumns + ` FROM ` + examPoolTable + ` a WHERE (a.exam_id IN (` +
		strings.Join(placeholders, ", ") + `)) AND ` + examPoolActiveWhere

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("ai exam repo list by ids: %w", err)
	}
	defer rows.Close()

	var out []*exam.ExamPool
	for rows.Next() {
		m, err := scanExamPool(rows)
		if err != nil {
			return nil, fmt.Errorf("ai exam repo scan row: %w", err)
		}
		out = append(out, ModelToDomainExamPool(m))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ai exam repo rows iteration: %w", err)
	}
	return out, nil
}

func (r *ExamPoolRepository) Create(ctx context.Context, e *exam.ExamPool) (*exam.ExamPool, error) {
	query := `
		INSERT INTO ` + examPoolTable + `
			(exam_id, req_exam_type, req_grade, req_level, req_num_ques,
			 req_semester, req_program, req_extras,
			 ai_title, ai_short_text, ai_questions_json,
			 rpt_flg, kwords, note, exam_status, create_id, create_dt, modify_dt)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	now := mtime.Now().Time
	_, err := r.db.Exec(ctx, query,
		e.ExamId(), e.ReqExamType(), e.ReqGrade(), e.ReqLevel(), e.ReqNumQues(),
		e.ReqSemester(), e.ReqProgram(), e.ReqExtras(),
		e.AiTitle(), e.AiShortText(), e.AiQuestionsJson(),
		e.RptFlg(), e.Kwords(), e.Note(), e.ExamStatus(), e.CreateId(), now, now)
	if err != nil {
		return nil, fmt.Errorf("ai exam repo create: %w", err)
	}

	return r.FindByExamId(ctx, e.ExamId())
}

func ModelToDomainExamPool(m *models.ExamPoolModel) *exam.ExamPool {
	e := exam.NewExamPool()
	e.SetExamId(m.ExamId)
	e.SetReqExamType(m.ReqExamType)
	e.SetReqGrade(m.ReqGrade)
	e.SetReqLevel(m.ReqLevel)
	e.SetReqNumQues(m.ReqNumQues)
	e.SetReqSemester(m.ReqSemester)
	e.SetReqProgram(m.ReqProgram)
	e.SetReqExtras(m.ReqExtras)
	e.SetAiTitle(m.AiTitle)
	e.SetAiShortText(m.AiShortText)
	e.SetAiQuestionsJson(m.AiQuestionsJson)
	e.SetRptFlg(m.RptFlg)
	e.SetKwords(m.Kwords)
	e.SetNote(m.Note)
	e.SetExamStatus(m.ExamStatus)
	e.SetStatus(m.Status)
	e.SetCreateId(m.CreateId)
	e.SetCreateDt(mtime.MathTime{Time: m.CreateDt})
	e.SetModifyId(m.ModifyId)
	e.SetModifyDt(mtime.MathTime{Time: m.ModifyDt})
	return e
}
