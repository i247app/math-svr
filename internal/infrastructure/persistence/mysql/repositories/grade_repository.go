package repositories

import (
	"context"
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

	gradeActiveWhere = `g.status IN (?) AND g.deleted_dt IS NULL`
)

func gradeJoinArgs(lang enum.LanguageType) []any {
	return []any{lang, enum.StatusActive}
}

func gradeActiveArgs() []any {
	return []any{enum.StatusActive}
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

func (r *GradeRepository) ListGrades(ctx context.Context, params *grade.ListGradesParams) ([]*grade.Grade, *pagination.Pagination, error) {
	lang := params.Language
	if lang == "" {
		lang = enum.LanguageTypeVietnamese
	}

	var total int64
	countQuery := `SELECT COUNT(*) FROM ` + gradeFromJoin +
		` WHERE ` + gradeActiveWhere
	countArgs := append(gradeJoinArgs(lang), gradeActiveArgs()...)
	if err := r.db.QueryRow(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, nil, fmt.Errorf("grade repo count: %w", err)
	}

	pg := pagination.NewPagination(params.Page, params.Limit, total)

	listArgs := append(gradeJoinArgs(lang), gradeActiveArgs()...)
	listArgs = append(listArgs, pg.Size, pg.Skip)
	query := `SELECT ` + gradeColumns + ` FROM ` + gradeFromJoin +
		` WHERE ` + gradeActiveWhere +
		` ORDER BY g.display_order ASC LIMIT ? OFFSET ?`
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
func (r *GradeRepository) ListGradesByIds(ctx context.Context, ids []string, lang enum.LanguageType) ([]*grade.Grade, error) {
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
