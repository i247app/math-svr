package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"slices"
	"strings"

	"math-ai.com/math-ai/internal/domain/prompt"
	"math-ai.com/math-ai/internal/domain/shared/mtime"
	"math-ai.com/math-ai/internal/infrastructure/database"
	"math-ai.com/math-ai/internal/infrastructure/persistence/mysql/models"
	"math-ai.com/math-ai/internal/shared/enum"
)

const (
	promptTable = "ma_prompts"

	promptColumns = `p.prompt_id, p.prompt_code, p.grade, p.language, p.system_prompt, p.prompt_version, p.description,
		p.prompt_status, p.rpt_flg, p.kwords, p.note, p.status,
		p.create_id, p.create_dt, p.modify_id, p.modify_dt`

	promptActiveWhere = `p.status IN (?) AND p.deleted_dt IS NULL`
)

func promptActiveArgs() []any {
	return []any{enum.StatusActive}
}

type PromptRepository struct {
	db database.Executor
}

func NewPromptRepository(db database.Executor) prompt.IRepository {
	return &PromptRepository{db: db}
}

func scanPrompt(s database.RowScanner) (*models.PromptModel, error) {
	var m models.PromptModel
	if err := s.Scan(&m.PromptId, &m.PromptCode, &m.Grade, &m.Language, &m.SystemPrompt, &m.PromptVersion, &m.Description,
		&m.PromptStatus, &m.RptFlg, &m.Kwords, &m.Note, &m.Status,
		&m.CreateId, &m.CreateDt, &m.ModifyId, &m.ModifyDt); err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *PromptRepository) findOneBy(ctx context.Context, where string, args ...any) (*prompt.Prompt, error) {
	query := `SELECT ` + promptColumns + ` FROM ` + promptTable + ` p WHERE (` + where + `) AND ` + promptActiveWhere

	m, err := scanPrompt(r.db.QueryRow(ctx, query, slices.Concat(args, promptActiveArgs())...))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("prompt repo find (%s): %w", where, err)
	}
	return ModelToDomainPrompt(m), nil
}

// FindByScope matches NULL to NULL with <=>, so a code that does not vary
// by grade finds its one row through Grade == nil.
func (r *PromptRepository) FindByScope(ctx context.Context, scope prompt.Scope) (*prompt.Prompt, error) {
	return r.findOneBy(ctx, `p.prompt_code = ? AND p.grade <=> ? AND p.language <=> ?`,
		string(scope.Code), scope.Grade, scope.Language)
}

func (r *PromptRepository) FindByPromptId(ctx context.Context, promptId int64) (*prompt.Prompt, error) {
	return r.findOneBy(ctx, `p.prompt_id = ?`, promptId)
}

func (r *PromptRepository) List(ctx context.Context, filter prompt.ListFilter) ([]*prompt.Prompt, error) {
	clause, args := buildPromptFilterClause(filter)
	query := `SELECT ` + promptColumns + ` FROM ` + promptTable + ` p ` + whereActive(clause, promptActiveWhere) +
		` ORDER BY p.prompt_code, p.grade, p.language`

	rows, err := r.db.Query(ctx, query, slices.Concat(args, promptActiveArgs())...)
	if err != nil {
		return nil, fmt.Errorf("prompt repo list: %w", err)
	}
	defer rows.Close()

	var out []*prompt.Prompt
	for rows.Next() {
		m, err := scanPrompt(rows)
		if err != nil {
			return nil, fmt.Errorf("prompt repo list scan: %w", err)
		}
		out = append(out, ModelToDomainPrompt(m))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("prompt repo list rows: %w", err)
	}
	return out, nil
}

// buildPromptFilterClause returns an optional " AND …" fragment for
// whereActive, plus its args in the same order.
func buildPromptFilterClause(f prompt.ListFilter) (string, []any) {
	var parts []string
	var args []any
	if f.Code != nil {
		parts = append(parts, "p.prompt_code = ?")
		args = append(args, string(*f.Code))
	}
	if f.Grade != nil {
		parts = append(parts, "p.grade = ?")
		args = append(args, *f.Grade)
	}
	if f.Language != nil {
		parts = append(parts, "p.language = ?")
		args = append(args, *f.Language)
	}
	if len(parts) == 0 {
		return "", nil
	}
	return " AND " + strings.Join(parts, " AND "), args
}

// UpdatePrompt relies on prompt_version changing on every match. MySQL
// reports CHANGED rows by default (FOUND rows with clientFoundRows), and a
// statement that bumps a counter always changes the row it matched, so
// either way RowsAffected is 1 exactly when the row exists — even when the
// new text equals the old one.
func (r *PromptRepository) UpdatePrompt(ctx context.Context, promptId int64, systemPrompt string, modifyId *int64) (bool, error) {
	query := `UPDATE ` + promptTable + ` p SET p.system_prompt = ?, p.prompt_version = p.prompt_version + 1, p.modify_id = ?` +
		` WHERE (p.prompt_id = ?) AND ` + promptActiveWhere

	res, err := r.db.Exec(ctx, query, slices.Concat([]any{systemPrompt, modifyId, promptId}, promptActiveArgs())...)
	if err != nil {
		return false, fmt.Errorf("prompt repo update %d: %w", promptId, err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("prompt repo update %d rows affected: %w", promptId, err)
	}
	return n > 0, nil
}

func ModelToDomainPrompt(m *models.PromptModel) *prompt.Prompt {
	p := prompt.NewPrompt()
	p.SetPromptId(m.PromptId)
	p.SetPromptCode(prompt.Code(m.PromptCode))
	p.SetGrade(m.Grade)
	p.SetLanguage(m.Language)
	p.SetSystemPrompt(m.SystemPrompt)
	p.SetPromptVersion(m.PromptVersion)
	p.SetDescription(m.Description)
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
