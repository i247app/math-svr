package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"math-ai.com/math-ai/internal/domain/grade"
	mtime "math-ai.com/math-ai/internal/domain/shared/time"
	"math-ai.com/math-ai/internal/infrastructure/database"
	"math-ai.com/math-ai/internal/infrastructure/persistence/mysql/models"
	"math-ai.com/math-ai/internal/shared/enum"
	"math-ai.com/math-ai/internal/shared/pagination"
)

const (
	gradeTable             = "ma_grades"
	gradeTranslationsTable = "ma_grade_translations"

	// LEFT JOIN so a grade without a translation in the requested
	// language still surfaces with its base label / description.
	gradeFromJoin = gradeTable + ` g
	LEFT JOIN ` + gradeTranslationsTable + ` t
	  ON t.grade_id = g.grade_id
	  AND t.language = ?
	  AND t.status = ?
	  AND t.deleted_dt IS NULL`

	gradeColumns = `g.id, g.grade_id,
		COALESCE(t.label, g.label) AS label,
		COALESCE(t.description, g.description) AS description,
		g.image_key, g.display_order, g.note,
		g.grade_status, g.status,
		g.create_id, g.create_dt, g.modify_id, g.modify_dt`

	gradeActiveWhere = `g.status = ? AND g.deleted_dt IS NULL
		AND (g.grade_status IS NULL OR g.grade_status != ?)`

	gradeBareActiveWhere = `g.status = ? AND g.deleted_dt IS NULL
		AND (g.grade_status IS NULL OR g.grade_status != ?)`
)

func gradeJoinArgs(lang enum.LanguageType) []any {
	if lang == "" {
		lang = enum.LanguageTypeVietnamese
	}
	return []any{lang, enum.StatusActive}
}

func gradeActiveArgs() []any {
	return []any{enum.StatusActive, enum.StatusInactive}
}

type GradeRepository struct {
	db database.Executor
}

func NewGradeRepository(db database.Executor) grade.IRepository {
	return &GradeRepository{db: db}
}

func scanGrade(s database.RowScanner) (*models.GradeModel, error) {
	var m models.GradeModel
	if err := s.Scan(&m.Id, &m.GradeId, &m.Label, &m.Description, &m.ImageKey,
		&m.DisplayOrder, &m.Note, &m.GradeStatus, &m.Status,
		&m.CreateId, &m.CreateDt, &m.ModifyId, &m.ModifyDt); err != nil {
		return nil, err
	}
	return &m, nil
}

// findOneBy is the single-row read helper. `where` is a package-controlled
// SQL fragment; args supply placeholders. gradeActiveWhere is prepended
// so every read excludes soft-deleted and inactive rows.
func (r *GradeRepository) findOneBy(ctx context.Context, lang enum.LanguageType, where string, args ...any) (*grade.Grade, error) {
	fullArgs := append(gradeJoinArgs(lang), gradeActiveArgs()...)
	fullArgs = append(fullArgs, args...)
	query := `SELECT ` + gradeColumns + ` FROM ` + gradeFromJoin +
		` WHERE ` + gradeActiveWhere + ` AND (` + where + `)`

	m, err := scanGrade(r.db.QueryRow(ctx, query, fullArgs...))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("grade repo find (%s): %w", where, err)
	}
	return ModelToDomainGrade(m), nil
}

// findBareById hydrates a grade right after INSERT using its surrogate
// id. The translation rows haven't been written yet at that point, so we
// skip the LEFT JOIN and just read the base label/description.
func (r *GradeRepository) findBareById(ctx context.Context, id int64) (*grade.Grade, error) {
	query := `SELECT g.id, g.grade_id, g.label, g.description AS description,
			g.image_key, g.display_order, g.note, g.grade_status, g.status,
			g.create_id, g.create_dt, g.modify_id, g.modify_dt
		FROM ` + gradeTable + ` g
		WHERE ` + gradeBareActiveWhere + ` AND g.id = ?`
	args := append(gradeActiveArgs(), id)

	m, err := scanGrade(r.db.QueryRow(ctx, query, args...))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("grade repo find bare by id: %w", err)
	}
	return ModelToDomainGrade(m), nil
}

func (r *GradeRepository) FindByGradeId(ctx context.Context, gradeId int64, language enum.LanguageType) (*grade.Grade, error) {
	return r.findOneBy(ctx, language, "g.grade_id = ?", gradeId)
}

func (r *GradeRepository) ListGrades(ctx context.Context, params *grade.ListGradesParams) ([]*grade.Grade, *pagination.Pagination, error) {
	lang := params.Language
	if lang == "" {
		lang = enum.LanguageTypeVietnamese
	}

	countArgs := append(gradeJoinArgs(lang), gradeActiveArgs()...)
	countQuery := `SELECT COUNT(*) FROM ` + gradeFromJoin +
		` WHERE ` + gradeActiveWhere
	var total int64
	if err := r.db.QueryRow(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, nil, fmt.Errorf("grade repo count: %w", err)
	}

	listArgs := append(gradeJoinArgs(lang), gradeActiveArgs()...)
	query := `SELECT ` + gradeColumns + ` FROM ` + gradeFromJoin +
		` WHERE ` + gradeActiveWhere +
		` ORDER BY g.display_order ASC, g.id ASC`

	var pg *pagination.Pagination
	if !params.TakeAll {
		pg = pagination.NewPagination(params.Page, params.Limit, total)
		query += ` LIMIT ? OFFSET ?`
		listArgs = append(listArgs, pg.Size, pg.Skip)
	} else {
		pg = pagination.NewPagination(1, total, total)
	}

	rows, err := r.db.Query(ctx, query, listArgs...)
	if err != nil {
		return nil, nil, fmt.Errorf("grade repo list: %w", err)
	}
	defer rows.Close()

	var grades []*grade.Grade
	for rows.Next() {
		m, err := scanGrade(rows)
		if err != nil {
			return nil, nil, fmt.Errorf("grade repo scan row: %w", err)
		}
		grades = append(grades, ModelToDomainGrade(m))
	}
	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("grade repo rows iteration: %w", err)
	}

	return grades, pg, nil
}

// ListGradesByIds — see program_repository's equivalent for rationale.
func (r *GradeRepository) ListGradesByIds(ctx context.Context, ids []int64, lang enum.LanguageType) ([]*grade.Grade, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	if lang == "" {
		lang = enum.LanguageTypeVietnamese
	}

	placeholders := make([]string, len(ids))
	args := append(gradeJoinArgs(lang), gradeActiveArgs()...)
	for i, id := range ids {
		placeholders[i] = "?"
		args = append(args, id)
	}

	query := `SELECT ` + gradeColumns + ` FROM ` + gradeFromJoin +
		` WHERE ` + gradeActiveWhere +
		` AND g.grade_id IN (` + strings.Join(placeholders, ",") + `)`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("grade repo list by ids: %w", err)
	}
	defer rows.Close()

	var grades []*grade.Grade
	for rows.Next() {
		m, err := scanGrade(rows)
		if err != nil {
			return nil, fmt.Errorf("grade repo scan row: %w", err)
		}
		grades = append(grades, ModelToDomainGrade(m))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("grade repo rows iteration: %w", err)
	}
	return grades, nil
}

func (r *GradeRepository) Create(ctx context.Context, g *grade.Grade) (*grade.Grade, error) {
	query := `
		INSERT INTO ` + gradeTable + `
			(grade_id, label, description, image_key, display_order, note, grade_status)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	result, err := r.db.Exec(ctx, query,
		g.GradeId(), g.Label(), g.Description(), g.ImageKey(),
		g.DisplayOrder(), g.Note(), g.GradeStatus())
	if err != nil {
		return nil, fmt.Errorf("grade repo create: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("grade repo last insert id: %w", err)
	}
	return r.findBareById(ctx, id)
}

// Update applies a partial update using COALESCE(?, col) for every
// nullable / non-zero column. display_order is int8 — a literal zero is
// a legal value, so it is always written and not coalesced.
func (r *GradeRepository) Update(ctx context.Context, g *grade.Grade) error {
	var label, description, imageKey any
	if g.Label() != "" {
		label = g.Label()
	}
	if g.Description() != "" {
		description = g.Description()
	}
	if g.ImageKey() != nil && *g.ImageKey() != "" {
		imageKey = *g.ImageKey()
	}

	query := `
		UPDATE ` + gradeTable + `
		SET label         = COALESCE(?, label),
			description   = COALESCE(?, description),
			image_key     = COALESCE(?, image_key),
			display_order = ?,
			note          = COALESCE(?, note),
			modify_dt     = ?
		WHERE grade_id = ?
	`
	if _, err := r.db.Exec(ctx, query,
		label, description, imageKey,
		g.DisplayOrder(), g.Note(), mtime.Now().Time, g.GradeId()); err != nil {
		return fmt.Errorf("grade repo update: %w", err)
	}
	return nil
}

func (r *GradeRepository) SoftDeleteByGradeId(ctx context.Context, gradeId int64) error {
	query := `
		UPDATE ` + gradeTable + `
		SET grade_status = ?,
			status       = ?,
			deleted_dt   = ?,
			modify_dt    = ?
		WHERE grade_id = ?
	`
	now := mtime.Now().Time
	if _, err := r.db.Exec(ctx, query,
		enum.StatusInactive, enum.StatusInactive, now, now, gradeId); err != nil {
		return fmt.Errorf("grade repo soft delete: %w", err)
	}
	return nil
}

func (r *GradeRepository) ForceDeleteByGradeId(ctx context.Context, gradeId int64) error {
	query := `
		DELETE FROM ` + gradeTable + `
		WHERE grade_id = ?
	`
	if _, err := r.db.Exec(ctx, query, gradeId); err != nil {
		return fmt.Errorf("grade repo force delete: %w", err)
	}
	return nil
}

func ModelToDomainGrade(m *models.GradeModel) *grade.Grade {
	g := grade.NewGrade()
	g.SetId(m.Id)
	g.SetGradeId(m.GradeId)
	g.SetLabel(m.Label)
	g.SetDescription(m.Description)
	g.SetImageKey(m.ImageKey)
	g.SetDisplayOrder(m.DisplayOrder)
	g.SetNote(m.Note)
	g.SetGradeStatus(m.GradeStatus)
	g.SetStatus(m.Status)
	g.SetCreateId(m.CreateId)
	g.SetCreateDt(mtime.MathTime{Time: m.CreateDt})
	g.SetModifyId(m.ModifyId)
	g.SetModifyDt(mtime.MathTime{Time: m.ModifyDt})
	return g
}
