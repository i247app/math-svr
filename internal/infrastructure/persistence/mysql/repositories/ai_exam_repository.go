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
	aiExamTable = "ma_ai_exams"

	aiExamColumns = `a.id, a.ai_exam_id, a.req_exam_type, a.req_grade, a.req_level, a.req_num_ques,
		a.req_semester, a.req_program, a.req_extras,
		a.ai_title, a.ai_short_text, a.ai_questions_json,
		a.rpt_flg, a.kwords, a.note, a.ai_exam_status, a.status,
		a.create_id, a.create_dt, a.modify_id, a.modify_dt`

	aiExamActiveWhere = `a.status IN (?) AND a.deleted_dt IS NULL`
)

func aiExamActiveArgs() []any {
	return []any{enum.StatusActive}
}

type AiExamRepository struct {
	db database.Executor
}

func NewAiExamRepository(db database.Executor) exam.IAiExamRepository {
	return &AiExamRepository{db: db}
}

func scanAiExam(s database.RowScanner) (*models.AiExamModel, error) {
	var m models.AiExamModel
	if err := s.Scan(&m.Id, &m.AiExamId, &m.ReqExamType, &m.ReqGrade, &m.ReqLevel, &m.ReqNumQues,
		&m.ReqSemester, &m.ReqProgram, &m.ReqExtras,
		&m.AiTitle, &m.AiShortText, &m.AiQuestionsJson,
		&m.RptFlg, &m.Kwords, &m.Note, &m.AiExamStatus, &m.Status,
		&m.CreateId, &m.CreateDt, &m.ModifyId, &m.ModifyDt); err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *AiExamRepository) findOneBy(ctx context.Context, where string, args ...any) (*exam.AiExam, error) {
	fullArgs := slices.Concat(args, aiExamActiveArgs())
	query := `SELECT ` + aiExamColumns + ` FROM ` + aiExamTable + ` a WHERE (` +
		where + `) AND ` + aiExamActiveWhere

	m, err := scanAiExam(r.db.QueryRow(ctx, query, fullArgs...))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("ai exam repo find (%s): %w", where, err)
	}
	return ModelToDomainAiExam(m), nil
}

func (r *AiExamRepository) FindByAiExamId(ctx context.Context, aiExamId int64) (*exam.AiExam, error) {
	return r.findOneBy(ctx, "a.ai_exam_id = ?", aiExamId)
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
func (r *AiExamRepository) FindReusableByExtras(ctx context.Context, extras string, excludeProfileId int64) (*exam.AiExam, error) {
	if extras == "" {
		return nil, nil
	}
	args := slices.Concat([]any{extras, excludeProfileId}, aiExamActiveArgs())
	query := `SELECT ` + aiExamColumns + ` FROM ` + aiExamTable + ` a WHERE ` +
		`(a.req_extras = ?` +
		` AND NOT EXISTS (SELECT 1 FROM ` + userAiExamTable + ` u WHERE u.ai_exam_id = a.ai_exam_id AND u.profile_id = ?))` +
		` AND ` + aiExamActiveWhere +
		` ORDER BY RAND() LIMIT 1`

	m, err := scanAiExam(r.db.QueryRow(ctx, query, args...))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("ai exam repo find reusable by extras: %w", err)
	}
	return ModelToDomainAiExam(m), nil
}

// CountByExtras counts the reusable question sets under one tag. It is a
// separate round trip rather than a column on some other row because the
// number changes every time a generation lands, and a cached counter that
// drifts would silently stop the pool from ever growing.
func (r *AiExamRepository) CountByExtras(ctx context.Context, extras string) (int64, error) {
	if extras == "" {
		return 0, nil
	}
	args := slices.Concat([]any{extras}, aiExamActiveArgs())
	query := `SELECT COUNT(*) FROM ` + aiExamTable + ` a WHERE (a.req_extras = ?) AND ` +
		aiExamActiveWhere

	var total int64
	if err := r.db.QueryRow(ctx, query, args...).Scan(&total); err != nil {
		return 0, fmt.Errorf("ai exam repo count by extras: %w", err)
	}
	return total, nil
}

// ListByAiExamIds hydrates a batch of question sets. It builds the IN
// list from the id count rather than interpolating values, so the query
// stays parameterised however many ids arrive.
func (r *AiExamRepository) ListByAiExamIds(ctx context.Context, aiExamIds []int64) ([]*exam.AiExam, error) {
	if len(aiExamIds) == 0 {
		return nil, nil
	}

	placeholders := make([]string, 0, len(aiExamIds))
	args := make([]any, 0, len(aiExamIds)+1)
	for _, id := range aiExamIds {
		placeholders = append(placeholders, "?")
		args = append(args, id)
	}
	args = append(args, aiExamActiveArgs()...)

	query := `SELECT ` + aiExamColumns + ` FROM ` + aiExamTable + ` a WHERE (a.ai_exam_id IN (` +
		strings.Join(placeholders, ", ") + `)) AND ` + aiExamActiveWhere

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("ai exam repo list by ids: %w", err)
	}
	defer rows.Close()

	var out []*exam.AiExam
	for rows.Next() {
		m, err := scanAiExam(rows)
		if err != nil {
			return nil, fmt.Errorf("ai exam repo scan row: %w", err)
		}
		out = append(out, ModelToDomainAiExam(m))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ai exam repo rows iteration: %w", err)
	}
	return out, nil
}

func (r *AiExamRepository) findBareById(ctx context.Context, id int64) (*exam.AiExam, error) {
	args := slices.Concat([]any{id}, aiExamActiveArgs())
	query := `SELECT ` + aiExamColumns + ` FROM ` + aiExamTable + ` a WHERE (a.id = ?) AND ` +
		aiExamActiveWhere

	m, err := scanAiExam(r.db.QueryRow(ctx, query, args...))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("ai exam repo find bare by id: %w", err)
	}
	return ModelToDomainAiExam(m), nil
}

func (r *AiExamRepository) Create(ctx context.Context, e *exam.AiExam) (*exam.AiExam, error) {
	query := `
		INSERT INTO ` + aiExamTable + `
			(ai_exam_id, req_exam_type, req_grade, req_level, req_num_ques,
			 req_semester, req_program, req_extras,
			 ai_title, ai_short_text, ai_questions_json,
			 rpt_flg, kwords, note, ai_exam_status, create_id, create_dt, modify_dt)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	now := mtime.Now().Time
	result, err := r.db.Exec(ctx, query,
		e.AiExamId(), e.ReqExamType(), e.ReqGrade(), e.ReqLevel(), e.ReqNumQues(),
		e.ReqSemester(), e.ReqProgram(), e.ReqExtras(),
		e.AiTitle(), e.AiShortText(), e.AiQuestionsJson(),
		e.RptFlg(), e.Kwords(), e.Note(), e.AiExamStatus(), e.CreateId(), now, now)
	if err != nil {
		return nil, fmt.Errorf("ai exam repo create: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("ai exam repo last insert id: %w", err)
	}
	return r.findBareById(ctx, id)
}

func ModelToDomainAiExam(m *models.AiExamModel) *exam.AiExam {
	e := exam.NewAiExam()
	e.SetId(m.Id)
	e.SetAiExamId(m.AiExamId)
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
	e.SetAiExamStatus(m.AiExamStatus)
	e.SetStatus(m.Status)
	e.SetCreateId(m.CreateId)
	e.SetCreateDt(mtime.MathTime{Time: m.CreateDt})
	e.SetModifyId(m.ModifyId)
	e.SetModifyDt(mtime.MathTime{Time: m.ModifyDt})
	return e
}
