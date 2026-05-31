package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"math-ai.com/math-ai/internal/domain/grade"
	mtime "math-ai.com/math-ai/internal/domain/shared/time"
	"math-ai.com/math-ai/internal/infrastructure/database"
	"math-ai.com/math-ai/internal/infrastructure/persistence/mysql/models"
	"math-ai.com/math-ai/internal/shared/enum"
)

// GradeTranslationRepository owns ma_grade_translations.
const (
	gradeTranslationColumns = `gt.id, gt.grade_translation_id, gt.grade_id, gt.language,
		gt.label, gt.description AS description, gt.note,
		gt.gt_status, gt.status,
		gt.create_id, gt.create_dt, gt.modify_id, gt.modify_dt`

	gradeTranslationActiveWhere = `gt.status IN (?) AND gt.deleted_dt IS NULL
		AND (gt.gt_status IS NULL OR gt.gt_status != ?)`
)

func gradeTranslationActiveArgs() []any {
	return []any{enum.StatusActive, enum.StatusInactive}
}

type GradeTranslationRepository struct {
	db database.Executor
}

func NewGradeTranslationRepository(db database.Executor) grade.ITranslationRepository {
	return &GradeTranslationRepository{db: db}
}

func scanGradeTranslation(s database.RowScanner) (*models.GradeTranslationModel, error) {
	var m models.GradeTranslationModel
	if err := s.Scan(&m.Id, &m.GradeTranslationId, &m.GradeId, &m.Language,
		&m.Label, &m.Description, &m.Note,
		&m.GtStatus, &m.Status,
		&m.CreateId, &m.CreateDt, &m.ModifyId, &m.ModifyDt); err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *GradeTranslationRepository) ListByGradeId(ctx context.Context, gradeId int64) ([]*grade.GradeTranslation, error) {
	args := append(gradeTranslationActiveArgs(), gradeId)
	query := `SELECT ` + gradeTranslationColumns + ` FROM ` + gradeTranslationsTable + ` gt
		WHERE ` + gradeTranslationActiveWhere + ` AND gt.grade_id = ?
		ORDER BY gt.language ASC`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("grade translation repo list by grade id: %w", err)
	}
	defer rows.Close()

	var translations []*grade.GradeTranslation
	for rows.Next() {
		m, err := scanGradeTranslation(rows)
		if err != nil {
			return nil, fmt.Errorf("grade translation repo scan row: %w", err)
		}
		translations = append(translations, ModelToDomainGradeTranslation(m))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("grade translation repo rows iteration: %w", err)
	}
	return translations, nil
}

func (r *GradeTranslationRepository) FindByGradeIdAndLanguage(ctx context.Context, gradeId int64, language string) (*grade.GradeTranslation, error) {
	args := append(gradeTranslationActiveArgs(), gradeId, language)
	query := `SELECT ` + gradeTranslationColumns + ` FROM ` + gradeTranslationsTable + ` gt
		WHERE ` + gradeTranslationActiveWhere + ` AND gt.grade_id = ? AND gt.language = ?
		LIMIT 1`

	m, err := scanGradeTranslation(r.db.QueryRow(ctx, query, args...))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("grade translation repo find by grade id and language: %w", err)
	}
	return ModelToDomainGradeTranslation(m), nil
}

func (r *GradeTranslationRepository) Create(ctx context.Context, t *grade.GradeTranslation) (*grade.GradeTranslation, error) {
	query := `
		INSERT INTO ` + gradeTranslationsTable + `
			(grade_translation_id, grade_id, language, label, description, note, gt_status)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	if _, err := r.db.Exec(ctx, query,
		t.GradeTranslationId(), t.GradeId(), t.Language(),
		t.Label(), t.Description(), t.Note(), t.GtStatus()); err != nil {
		return nil, fmt.Errorf("grade translation repo create: %w", err)
	}
	return r.FindByGradeIdAndLanguage(ctx, t.GradeId(), t.Language())
}

func (r *GradeTranslationRepository) Update(ctx context.Context, t *grade.GradeTranslation) error {
	var label, description any
	if t.Label() != "" {
		label = t.Label()
	}
	if t.Description() != "" {
		description = t.Description()
	}

	query := `
		UPDATE ` + gradeTranslationsTable + `
		SET label       = COALESCE(?, label),
			description = COALESCE(?, description),
			note        = COALESCE(?, note),
			modify_dt   = ?
		WHERE grade_translation_id = ?
	`
	if _, err := r.db.Exec(ctx, query,
		label, description, t.Note(), mtime.Now().Time, t.GradeTranslationId()); err != nil {
		return fmt.Errorf("grade translation repo update: %w", err)
	}
	return nil
}

func (r *GradeTranslationRepository) SoftDeleteByGradeId(ctx context.Context, gradeId int64) error {
	query := `
		UPDATE ` + gradeTranslationsTable + `
		SET gt_status  = ?,
			status     = ?,
			deleted_dt = ?,
			modify_dt  = ?
		WHERE grade_id = ?
	`
	now := mtime.Now().Time
	if _, err := r.db.Exec(ctx, query,
		enum.StatusInactive, enum.StatusInactive, now, now, gradeId); err != nil {
		return fmt.Errorf("grade translation repo soft delete by grade id: %w", err)
	}
	return nil
}

func (r *GradeTranslationRepository) SoftDeleteByTranslationId(ctx context.Context, gradeTranslationId int64) error {
	query := `
		UPDATE ` + gradeTranslationsTable + `
		SET gt_status  = ?,
			status     = ?,
			deleted_dt = ?,
			modify_dt  = ?
		WHERE grade_translation_id = ?
	`
	now := mtime.Now().Time
	if _, err := r.db.Exec(ctx, query,
		enum.StatusInactive, enum.StatusInactive, now, now, gradeTranslationId); err != nil {
		return fmt.Errorf("grade translation repo soft delete by translation id: %w", err)
	}
	return nil
}

func (r *GradeTranslationRepository) ForceDeleteByGradeId(ctx context.Context, gradeId int64) error {
	query := `
		DELETE FROM ` + gradeTranslationsTable + `
		WHERE grade_id = ?
	`
	if _, err := r.db.Exec(ctx, query, gradeId); err != nil {
		return fmt.Errorf("grade translation repo force delete by grade id: %w", err)
	}
	return nil
}

func (r *GradeTranslationRepository) ForceDeleteByTranslationId(ctx context.Context, gradeTranslationId int64) error {
	query := `
		DELETE FROM ` + gradeTranslationsTable + `
		WHERE grade_translation_id = ?
	`
	if _, err := r.db.Exec(ctx, query, gradeTranslationId); err != nil {
		return fmt.Errorf("grade translation repo force delete by translation id: %w", err)
	}
	return nil
}

func ModelToDomainGradeTranslation(m *models.GradeTranslationModel) *grade.GradeTranslation {
	t := grade.NewGradeTranslation()
	t.SetId(m.Id)
	t.SetGradeTranslationId(m.GradeTranslationId)
	t.SetGradeId(m.GradeId)
	t.SetLanguage(m.Language)
	t.SetLabel(m.Label)
	t.SetDescription(m.Description)
	t.SetNote(m.Note)
	t.SetGtStatus(m.GtStatus)
	t.SetStatus(m.Status)
	t.SetCreateId(m.CreateId)
	t.SetCreateDt(mtime.MathTime{Time: m.CreateDt})
	t.SetModifyId(m.ModifyId)
	t.SetModifyDt(mtime.MathTime{Time: m.ModifyDt})
	return t
}
