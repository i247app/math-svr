package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"math-ai.com/math-ai/internal/domain/semester"
	mtime "math-ai.com/math-ai/internal/domain/shared/time"
	"math-ai.com/math-ai/internal/infrastructure/database"
	"math-ai.com/math-ai/internal/infrastructure/persistence/mysql/models"
	"math-ai.com/math-ai/internal/shared/enum"
)

// SemesterTranslationRepository owns ma_semester_translations.
// description is TEXT and nullable, coalesced to ” on read so the
// domain field is always a plain string.
const (
	semesterTranslationColumns = `st.id, st.semester_translation_id, st.semester_id, st.language,
		st.name, COALESCE(st.description, '') AS description, st.note,
		st.st_status, st.status,
		st.create_id, st.create_dt, st.modify_id, st.modify_dt`

	semesterTranslationActiveWhere = `st.status IN (?) AND st.deleted_dt IS NULL
		AND (st.st_status IS NULL OR st.st_status != ?)`
)

func semesterTranslationActiveArgs() []any {
	return []any{enum.StatusActive, enum.StatusInactive}
}

type SemesterTranslationRepository struct {
	db database.Executor
}

func NewSemesterTranslationRepository(db database.Executor) semester.ITranslationRepository {
	return &SemesterTranslationRepository{db: db}
}

func scanSemesterTranslation(s database.RowScanner) (*models.SemesterTranslationModel, error) {
	var m models.SemesterTranslationModel
	if err := s.Scan(&m.Id, &m.SemesterTranslationId, &m.SemesterId, &m.Language,
		&m.Name, &m.Description, &m.Note,
		&m.StStatus, &m.Status,
		&m.CreateId, &m.CreateDt, &m.ModifyId, &m.ModifyDt); err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *SemesterTranslationRepository) ListBySemesterId(ctx context.Context, semesterId int64) ([]*semester.SemesterTranslation, error) {
	args := append(semesterTranslationActiveArgs(), semesterId)
	query := `SELECT ` + semesterTranslationColumns + ` FROM ` + semesterTranslationsTable + ` st
		WHERE ` + semesterTranslationActiveWhere + ` AND st.semester_id = ?
		ORDER BY st.language ASC`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("semester translation repo list by semester id: %w", err)
	}
	defer rows.Close()

	var translations []*semester.SemesterTranslation
	for rows.Next() {
		m, err := scanSemesterTranslation(rows)
		if err != nil {
			return nil, fmt.Errorf("semester translation repo scan row: %w", err)
		}
		translations = append(translations, ModelToDomainSemesterTranslation(m))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("semester translation repo rows iteration: %w", err)
	}
	return translations, nil
}

func (r *SemesterTranslationRepository) FindBySemesterIdAndLanguage(ctx context.Context, semesterId int64, language enum.LanguageType) (*semester.SemesterTranslation, error) {
	args := append(semesterTranslationActiveArgs(), semesterId, language)
	query := `SELECT ` + semesterTranslationColumns + ` FROM ` + semesterTranslationsTable + ` st
		WHERE ` + semesterTranslationActiveWhere + ` AND st.semester_id = ? AND st.language = ?
		LIMIT 1`

	m, err := scanSemesterTranslation(r.db.QueryRow(ctx, query, args...))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("semester translation repo find by semester id and language: %w", err)
	}
	return ModelToDomainSemesterTranslation(m), nil
}

func (r *SemesterTranslationRepository) Create(ctx context.Context, t *semester.SemesterTranslation) (*semester.SemesterTranslation, error) {
	query := `
		INSERT INTO ` + semesterTranslationsTable + `
			(semester_translation_id, semester_id, language, name, description, note, st_status, create_dt, modify_dt)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	var description any
	if t.Description() != "" {
		description = t.Description()
	}
	if _, err := r.db.Exec(ctx, query,
		t.SemesterTranslationId(), t.SemesterId(), t.Language(),
		t.Name(), description, t.Note(), t.StStatus(), mtime.Now().Time, mtime.Now().Time); err != nil {
		return nil, fmt.Errorf("semester translation repo create: %w", err)
	}
	return r.FindBySemesterIdAndLanguage(ctx, t.SemesterId(), enum.LanguageType(t.Language()))
}

func (r *SemesterTranslationRepository) Update(ctx context.Context, t *semester.SemesterTranslation) error {
	var name, description any
	if t.Name() != "" {
		name = t.Name()
	}
	if t.Description() != "" {
		description = t.Description()
	}

	query := `
		UPDATE ` + semesterTranslationsTable + `
		SET name        = COALESCE(?, name),
			description = COALESCE(?, description),
			note        = COALESCE(?, note),
			modify_dt   = ?
		WHERE semester_translation_id = ?
	`
	if _, err := r.db.Exec(ctx, query,
		name, description, t.Note(), mtime.Now().Time, t.SemesterTranslationId()); err != nil {
		return fmt.Errorf("semester translation repo update: %w", err)
	}
	return nil
}

func (r *SemesterTranslationRepository) SoftDeleteBySemesterId(ctx context.Context, semesterId int64) error {
	query := `
		UPDATE ` + semesterTranslationsTable + `
		SET st_status  = ?,
			status     = ?,
			deleted_dt = ?,
			modify_dt  = ?
		WHERE semester_id = ?
	`
	now := mtime.Now().Time
	if _, err := r.db.Exec(ctx, query,
		enum.StatusInactive, enum.StatusInactive, now, now, semesterId); err != nil {
		return fmt.Errorf("semester translation repo soft delete by semester id: %w", err)
	}
	return nil
}

func (r *SemesterTranslationRepository) SoftDeleteByTranslationId(ctx context.Context, semesterTranslationId int64) error {
	query := `
		UPDATE ` + semesterTranslationsTable + `
		SET st_status  = ?,
			status     = ?,
			deleted_dt = ?,
			modify_dt  = ?
		WHERE semester_translation_id = ?
	`
	now := mtime.Now().Time
	if _, err := r.db.Exec(ctx, query,
		enum.StatusInactive, enum.StatusInactive, now, now, semesterTranslationId); err != nil {
		return fmt.Errorf("semester translation repo soft delete by translation id: %w", err)
	}
	return nil
}

func (r *SemesterTranslationRepository) ForceDeleteBySemesterId(ctx context.Context, semesterId int64) error {
	query := `
		DELETE FROM ` + semesterTranslationsTable + `
		WHERE semester_id = ?
	`
	if _, err := r.db.Exec(ctx, query, semesterId); err != nil {
		return fmt.Errorf("semester translation repo force delete by semester id: %w", err)
	}
	return nil
}

func (r *SemesterTranslationRepository) ForceDeleteByTranslationId(ctx context.Context, semesterTranslationId int64) error {
	query := `
		DELETE FROM ` + semesterTranslationsTable + `
		WHERE semester_translation_id = ?
	`
	if _, err := r.db.Exec(ctx, query, semesterTranslationId); err != nil {
		return fmt.Errorf("semester translation repo force delete by translation id: %w", err)
	}
	return nil
}

func ModelToDomainSemesterTranslation(m *models.SemesterTranslationModel) *semester.SemesterTranslation {
	t := semester.NewSemesterTranslation()
	t.SetId(m.Id)
	t.SetSemesterTranslationId(m.SemesterTranslationId)
	t.SetSemesterId(m.SemesterId)
	t.SetLanguage(m.Language)
	t.SetName(m.Name)
	t.SetDescription(m.Description)
	t.SetNote(m.Note)
	t.SetStStatus(m.StStatus)
	t.SetStatus(m.Status)
	t.SetCreateId(m.CreateId)
	t.SetCreateDt(mtime.MathTime{Time: m.CreateDt})
	t.SetModifyId(m.ModifyId)
	t.SetModifyDt(mtime.MathTime{Time: m.ModifyDt})
	return t
}
