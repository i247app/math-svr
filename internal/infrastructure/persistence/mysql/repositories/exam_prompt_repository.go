package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"slices"

	"math-ai.com/math-ai/internal/domain/exam"
	"math-ai.com/math-ai/internal/domain/shared/mtime"
	"math-ai.com/math-ai/internal/infrastructure/database"
	"math-ai.com/math-ai/internal/infrastructure/persistence/mysql/models"
	"math-ai.com/math-ai/internal/shared/enum"
)

const (
	examPromptTable = "ma_exam_prompts"

	examPromptColumns = `ep.prompt_id, ep.grade, ep.system_prompt, ep.prompt_version,
		ep.prompt_status, ep.rpt_flg, ep.kwords, ep.note, ep.status,
		ep.create_id, ep.create_dt, ep.modify_id, ep.modify_dt`

	examPromptActiveWhere = `ep.status IN (?) AND ep.deleted_dt IS NULL`
)

func examPromptActiveArgs() []any {
	return []any{enum.StatusActive}
}

type ExamPromptRepository struct {
	db database.Executor
}

func NewExamPromptRepository(db database.Executor) exam.IExamPromptRepository {
	return &ExamPromptRepository{db: db}
}

func scanExamPrompt(s database.RowScanner) (*models.ExamPromptModel, error) {
	var m models.ExamPromptModel
	if err := s.Scan(&m.PromptId, &m.Grade, &m.SystemPrompt, &m.PromptVersion,
		&m.PromptStatus, &m.RptFlg, &m.Kwords, &m.Note, &m.Status,
		&m.CreateId, &m.CreateDt, &m.ModifyId, &m.ModifyDt); err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *ExamPromptRepository) FindByGrade(ctx context.Context, grade int) (*exam.ExamPrompt, error) {
	query := `SELECT ` + examPromptColumns + ` FROM ` + examPromptTable + ` ep` +
		` WHERE (ep.grade = ?) AND ` + examPromptActiveWhere

	m, err := scanExamPrompt(r.db.QueryRow(ctx, query, slices.Concat([]any{grade}, examPromptActiveArgs())...))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("exam prompt repo find by grade %d: %w", grade, err)
	}
	return ModelToDomainExamPrompt(m), nil
}

func (r *ExamPromptRepository) List(ctx context.Context) ([]*exam.ExamPrompt, error) {
	query := `SELECT ` + examPromptColumns + ` FROM ` + examPromptTable + ` ep` +
		` WHERE ` + examPromptActiveWhere + ` ORDER BY ep.grade`

	rows, err := r.db.Query(ctx, query, examPromptActiveArgs()...)
	if err != nil {
		return nil, fmt.Errorf("exam prompt repo list: %w", err)
	}
	defer rows.Close()

	var out []*exam.ExamPrompt
	for rows.Next() {
		m, err := scanExamPrompt(rows)
		if err != nil {
			return nil, fmt.Errorf("exam prompt repo list scan: %w", err)
		}
		out = append(out, ModelToDomainExamPrompt(m))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("exam prompt repo list rows: %w", err)
	}
	return out, nil
}

// UpdatePrompt relies on prompt_version changing on every match. MySQL
// reports CHANGED rows by default (FOUND rows with clientFoundRows), and a
// statement that bumps a counter always changes the row it matched, so
// either way RowsAffected is 1 exactly when the grade exists — even when
// the new text equals the old one.
func (r *ExamPromptRepository) UpdatePrompt(ctx context.Context, grade int, systemPrompt string, modifyId *int64) (bool, error) {
	query := `UPDATE ` + examPromptTable + ` ep SET ep.system_prompt = ?, ep.prompt_version = ep.prompt_version + 1, ep.modify_id = ?` +
		` WHERE (ep.grade = ?) AND ` + examPromptActiveWhere

	res, err := r.db.Exec(ctx, query, slices.Concat([]any{systemPrompt, modifyId, grade}, examPromptActiveArgs())...)
	if err != nil {
		return false, fmt.Errorf("exam prompt repo update grade %d: %w", grade, err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("exam prompt repo update grade %d rows affected: %w", grade, err)
	}
	return n > 0, nil
}

func ModelToDomainExamPrompt(m *models.ExamPromptModel) *exam.ExamPrompt {
	p := exam.NewExamPrompt()
	p.SetPromptId(m.PromptId)
	p.SetGrade(m.Grade)
	p.SetSystemPrompt(m.SystemPrompt)
	p.SetPromptVersion(m.PromptVersion)
	p.SetPromptStatus(m.PromptStatus)
	p.SetRptFlg(m.RptFlg)
	p.SetKwords(m.Kwords)
	p.SetNote(m.Note)
	p.SetStatus(m.Status)
	p.SetCreateId(m.CreateId)
	p.SetCreateDt(mtime.MathTime{Time: m.CreateDt})
	p.SetModifyId(m.ModifyId)
	p.SetModifyDt(mtime.MathTime{Time: m.ModifyDt})
	return p
}
