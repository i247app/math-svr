package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"math-ai.com/math-ai/internal/domain/chapter"
	mtime "math-ai.com/math-ai/internal/domain/shared/time"
	"math-ai.com/math-ai/internal/infrastructure/database"
	"math-ai.com/math-ai/internal/infrastructure/persistence/mysql/models"
	"math-ai.com/math-ai/internal/shared/enum"
)

// ChapterTranslationRepository owns ma_chapter_translations. As with the
// parent table the column is `description`; the domain surfaces it as
// Description and only the SQL below carries the typo.
const (
	chapterTranslationColumns = `ct.id, ct.chapter_translation_id, ct.chapter_id, ct.language,
		ct.label, ct.description AS description, ct.note,
		ct.ct_status, ct.status,
		ct.create_id, ct.create_dt, ct.modify_id, ct.modify_dt`

	chapterTranslationActiveWhere = `ct.status IN (?) AND ct.deleted_dt IS NULL
		AND (ct.ct_status IS NULL OR ct.ct_status != ?)`
)

func chapterTranslationActiveArgs() []any {
	return []any{enum.StatusActive, enum.StatusInactive}
}

type ChapterTranslationRepository struct {
	db database.Executor
}

func NewChapterTranslationRepository(db database.Executor) chapter.ITranslationRepository {
	return &ChapterTranslationRepository{db: db}
}

func scanChapterTranslation(s database.RowScanner) (*models.ChapterTranslationModel, error) {
	var m models.ChapterTranslationModel
	if err := s.Scan(&m.Id, &m.ChapterTranslationId, &m.ChapterId, &m.Language,
		&m.Label, &m.Description, &m.Note,
		&m.CtStatus, &m.Status,
		&m.CreateId, &m.CreateDt, &m.ModifyId, &m.ModifyDt); err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *ChapterTranslationRepository) ListByChapterId(ctx context.Context, chapterId string) ([]*chapter.ChapterTranslation, error) {
	args := append(chapterTranslationActiveArgs(), chapterId)
	query := `SELECT ` + chapterTranslationColumns + ` FROM ` + chapterTranslationsTable + ` ct
		WHERE ` + chapterTranslationActiveWhere + ` AND ct.chapter_id = ?
		ORDER BY ct.language ASC`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("chapter translation repo list by chapter id: %w", err)
	}
	defer rows.Close()

	var translations []*chapter.ChapterTranslation
	for rows.Next() {
		m, err := scanChapterTranslation(rows)
		if err != nil {
			return nil, fmt.Errorf("chapter translation repo scan row: %w", err)
		}
		translations = append(translations, ModelToDomainChapterTranslation(m))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("chapter translation repo rows iteration: %w", err)
	}
	return translations, nil
}

func (r *ChapterTranslationRepository) FindByChapterIdAndLanguage(ctx context.Context, chapterId string, language string) (*chapter.ChapterTranslation, error) {
	args := append(chapterTranslationActiveArgs(), chapterId, language)
	query := `SELECT ` + chapterTranslationColumns + ` FROM ` + chapterTranslationsTable + ` ct
		WHERE ` + chapterTranslationActiveWhere + ` AND ct.chapter_id = ? AND ct.language = ?
		LIMIT 1`

	m, err := scanChapterTranslation(r.db.QueryRow(ctx, query, args...))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("chapter translation repo find by chapter id and language: %w", err)
	}
	return ModelToDomainChapterTranslation(m), nil
}

func (r *ChapterTranslationRepository) Create(ctx context.Context, t *chapter.ChapterTranslation) (*chapter.ChapterTranslation, error) {
	query := `
		INSERT INTO ` + chapterTranslationsTable + `
			(chapter_translation_id, chapter_id, language, label, description, note, ct_status)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	if _, err := r.db.Exec(ctx, query,
		t.ChapterTranslationId(), t.ChapterId(), t.Language(),
		t.Label(), t.Description(), t.Note(), t.CtStatus()); err != nil {
		return nil, fmt.Errorf("chapter translation repo create: %w", err)
	}
	return r.FindByChapterIdAndLanguage(ctx, t.ChapterId(), t.Language())
}

func (r *ChapterTranslationRepository) Update(ctx context.Context, t *chapter.ChapterTranslation) error {
	var label, description any
	if t.Label() != "" {
		label = t.Label()
	}
	if t.Description() != "" {
		description = t.Description()
	}

	query := `
		UPDATE ` + chapterTranslationsTable + `
		SET label       = COALESCE(?, label),
			description = COALESCE(?, description),
			note        = COALESCE(?, note),
			modify_dt   = ?
		WHERE chapter_translation_id = ?
	`
	if _, err := r.db.Exec(ctx, query,
		label, description, t.Note(), mtime.Now().Time, t.ChapterTranslationId()); err != nil {
		return fmt.Errorf("chapter translation repo update: %w", err)
	}
	return nil
}

func (r *ChapterTranslationRepository) SoftDeleteByChapterId(ctx context.Context, chapterId string) error {
	query := `
		UPDATE ` + chapterTranslationsTable + `
		SET ct_status  = ?,
			status     = ?,
			deleted_dt = ?,
			modify_dt  = ?
		WHERE chapter_id = ?
	`
	now := mtime.Now().Time
	if _, err := r.db.Exec(ctx, query,
		enum.StatusInactive, enum.StatusInactive, now, now, chapterId); err != nil {
		return fmt.Errorf("chapter translation repo soft delete by chapter id: %w", err)
	}
	return nil
}

func (r *ChapterTranslationRepository) SoftDeleteByTranslationId(ctx context.Context, chapterTranslationId string) error {
	query := `
		UPDATE ` + chapterTranslationsTable + `
		SET ct_status  = ?,
			status     = ?,
			deleted_dt = ?,
			modify_dt  = ?
		WHERE chapter_translation_id = ?
	`
	now := mtime.Now().Time
	if _, err := r.db.Exec(ctx, query,
		enum.StatusInactive, enum.StatusInactive, now, now, chapterTranslationId); err != nil {
		return fmt.Errorf("chapter translation repo soft delete by translation id: %w", err)
	}
	return nil
}

func (r *ChapterTranslationRepository) ForceDeleteByChapterId(ctx context.Context, chapterId string) error {
	query := `
		DELETE FROM ` + chapterTranslationsTable + `
		WHERE chapter_id = ?
	`
	if _, err := r.db.Exec(ctx, query, chapterId); err != nil {
		return fmt.Errorf("chapter translation repo force delete by chapter id: %w", err)
	}
	return nil
}

func (r *ChapterTranslationRepository) ForceDeleteByTranslationId(ctx context.Context, chapterTranslationId string) error {
	query := `
		DELETE FROM ` + chapterTranslationsTable + `
		WHERE chapter_translation_id = ?
	`
	if _, err := r.db.Exec(ctx, query, chapterTranslationId); err != nil {
		return fmt.Errorf("chapter translation repo force delete by translation id: %w", err)
	}
	return nil
}

func ModelToDomainChapterTranslation(m *models.ChapterTranslationModel) *chapter.ChapterTranslation {
	t := chapter.NewChapterTranslation()
	t.SetId(m.Id)
	t.SetChapterTranslationId(m.ChapterTranslationId)
	t.SetChapterId(m.ChapterId)
	t.SetLanguage(m.Language)
	t.SetLabel(m.Label)
	t.SetDescription(m.Description)
	t.SetNote(m.Note)
	t.SetCtStatus(m.CtStatus)
	t.SetStatus(m.Status)
	t.SetCreateId(m.CreateId)
	t.SetCreateDt(mtime.MathTime{Time: m.CreateDt})
	t.SetModifyId(m.ModifyId)
	t.SetModifyDt(mtime.MathTime{Time: m.ModifyDt})
	return t
}
