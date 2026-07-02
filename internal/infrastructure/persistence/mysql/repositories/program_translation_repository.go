package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"math-ai.com/math-ai/internal/domain/program"
	"math-ai.com/math-ai/internal/domain/shared/mtime"
	"math-ai.com/math-ai/internal/infrastructure/database"
	"math-ai.com/math-ai/internal/infrastructure/persistence/mysql/models"
	"math-ai.com/math-ai/internal/shared/enum"
)

// ProgramTranslationRepository owns ma_program_translations. The
// per-row status column is named `gt_status` in the migration — a
// copy-paste typo from grade — so the SQL below references that column
// literally while the domain field stays `PtStatus`.
const (
	programTranslationColumns = `pt.id, pt.program_translation_id, pt.program_id, pt.language,
		pt.label, pt.description AS description, pt.note,
		pt.gt_status, pt.status,
		pt.create_id, pt.create_dt, pt.modify_id, pt.modify_dt`

	programTranslationActiveWhere = `pt.status IN (?) AND pt.deleted_dt IS NULL
		AND (pt.gt_status IS NULL OR pt.gt_status != ?)`
)

func programTranslationActiveArgs() []any {
	return []any{enum.StatusActive, enum.StatusInactive}
}

type ProgramTranslationRepository struct {
	db database.Executor
}

func NewProgramTranslationRepository(db database.Executor) program.ITranslationRepository {
	return &ProgramTranslationRepository{db: db}
}

func scanProgramTranslation(s database.RowScanner) (*models.ProgramTranslationModel, error) {
	var m models.ProgramTranslationModel
	if err := s.Scan(&m.Id, &m.ProgramTranslationId, &m.ProgramId, &m.Language,
		&m.Label, &m.Description, &m.Note,
		&m.PtStatus, &m.Status,
		&m.CreateId, &m.CreateDt, &m.ModifyId, &m.ModifyDt); err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *ProgramTranslationRepository) ListByProgramId(ctx context.Context, programId int64) ([]*program.ProgramTranslation, error) {
	args := append(programTranslationActiveArgs(), programId)
	query := `SELECT ` + programTranslationColumns + ` FROM ` + programTranslationsTable + ` pt
		WHERE ` + programTranslationActiveWhere + ` AND pt.program_id = ?
		ORDER BY pt.language ASC`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("program translation repo list by program id: %w", err)
	}
	defer rows.Close()

	var translations []*program.ProgramTranslation
	for rows.Next() {
		m, err := scanProgramTranslation(rows)
		if err != nil {
			return nil, fmt.Errorf("program translation repo scan row: %w", err)
		}
		translations = append(translations, ModelToDomainProgramTranslation(m))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("program translation repo rows iteration: %w", err)
	}
	return translations, nil
}

func (r *ProgramTranslationRepository) FindByProgramIdAndLanguage(ctx context.Context, programId int64, language string) (*program.ProgramTranslation, error) {
	args := append(programTranslationActiveArgs(), programId, language)
	query := `SELECT ` + programTranslationColumns + ` FROM ` + programTranslationsTable + ` pt
		WHERE ` + programTranslationActiveWhere + ` AND pt.program_id = ? AND pt.language = ?
		LIMIT 1`

	m, err := scanProgramTranslation(r.db.QueryRow(ctx, query, args...))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("program translation repo find by program id and language: %w", err)
	}
	return ModelToDomainProgramTranslation(m), nil
}

func (r *ProgramTranslationRepository) Create(ctx context.Context, t *program.ProgramTranslation) (*program.ProgramTranslation, error) {
	query := `
		INSERT INTO ` + programTranslationsTable + `
			(program_translation_id, program_id, language, label, description, note, gt_status, create_dt, modify_dt)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	if _, err := r.db.Exec(ctx, query,
		t.ProgramTranslationId(), t.ProgramId(), t.Language(),
		t.Label(), t.Description(), t.Note(), t.PtStatus(), mtime.Now().Time, mtime.Now().Time); err != nil {
		return nil, fmt.Errorf("program translation repo create: %w", err)
	}
	return r.FindByProgramIdAndLanguage(ctx, t.ProgramId(), t.Language())
}

func (r *ProgramTranslationRepository) Update(ctx context.Context, t *program.ProgramTranslation) error {
	var label, description any
	if t.Label() != "" {
		label = t.Label()
	}
	if t.Description() != "" {
		description = t.Description()
	}

	query := `
		UPDATE ` + programTranslationsTable + `
		SET label       = COALESCE(?, label),
			description = COALESCE(?, description),
			note        = COALESCE(?, note),
			modify_dt   = ?
		WHERE program_translation_id = ?
	`
	if _, err := r.db.Exec(ctx, query,
		label, description, t.Note(), mtime.Now().Time, t.ProgramTranslationId()); err != nil {
		return fmt.Errorf("program translation repo update: %w", err)
	}
	return nil
}

func (r *ProgramTranslationRepository) SoftDeleteByProgramId(ctx context.Context, programId int64) error {
	query := `
		UPDATE ` + programTranslationsTable + `
		SET gt_status  = ?,
			status     = ?,
			deleted_dt = ?,
			modify_dt  = ?
		WHERE program_id = ?
	`
	now := mtime.Now().Time
	if _, err := r.db.Exec(ctx, query,
		enum.StatusInactive, enum.StatusInactive, now, now, programId); err != nil {
		return fmt.Errorf("program translation repo soft delete by program id: %w", err)
	}
	return nil
}

func (r *ProgramTranslationRepository) SoftDeleteByTranslationId(ctx context.Context, programTranslationId int64) error {
	query := `
		UPDATE ` + programTranslationsTable + `
		SET gt_status  = ?,
			status     = ?,
			deleted_dt = ?,
			modify_dt  = ?
		WHERE program_translation_id = ?
	`
	now := mtime.Now().Time
	if _, err := r.db.Exec(ctx, query,
		enum.StatusInactive, enum.StatusInactive, now, now, programTranslationId); err != nil {
		return fmt.Errorf("program translation repo soft delete by translation id: %w", err)
	}
	return nil
}

func (r *ProgramTranslationRepository) ForceDeleteByProgramId(ctx context.Context, programId int64) error {
	query := `
		DELETE FROM ` + programTranslationsTable + `
		WHERE program_id = ?
	`
	if _, err := r.db.Exec(ctx, query, programId); err != nil {
		return fmt.Errorf("program translation repo force delete by program id: %w", err)
	}
	return nil
}

func (r *ProgramTranslationRepository) ForceDeleteByTranslationId(ctx context.Context, programTranslationId int64) error {
	query := `
		DELETE FROM ` + programTranslationsTable + `
		WHERE program_translation_id = ?
	`
	if _, err := r.db.Exec(ctx, query, programTranslationId); err != nil {
		return fmt.Errorf("program translation repo force delete by translation id: %w", err)
	}
	return nil
}

func ModelToDomainProgramTranslation(m *models.ProgramTranslationModel) *program.ProgramTranslation {
	t := program.NewProgramTranslation()
	t.SetId(m.Id)
	t.SetProgramTranslationId(m.ProgramTranslationId)
	t.SetProgramId(m.ProgramId)
	t.SetLanguage(m.Language)
	t.SetLabel(m.Label)
	t.SetDescription(m.Description)
	t.SetNote(m.Note)
	t.SetPtStatus(m.PtStatus)
	t.SetStatus(m.Status)
	t.SetCreateId(m.CreateId)
	t.SetCreateDt(mtime.MathTime{Time: m.CreateDt})
	t.SetModifyId(m.ModifyId)
	t.SetModifyDt(mtime.MathTime{Time: m.ModifyDt})
	return t
}
