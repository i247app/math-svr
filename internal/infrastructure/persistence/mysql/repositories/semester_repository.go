package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"math-ai.com/math-ai/internal/domain/semester"
	mtime "math-ai.com/math-ai/internal/domain/shared/time"
	"math-ai.com/math-ai/internal/infrastructure/database"
	"math-ai.com/math-ai/internal/infrastructure/persistence/mysql/models"
	"math-ai.com/math-ai/internal/shared/enum"
	"math-ai.com/math-ai/internal/shared/pagination"
)

// Reference-data aggregate. Like program/grade, but the base table uses
// `name` (not `label`) and `description TEXT` is nullable, so the SELECT
// COALESCEs to ” to keep the scan target a plain string.
const (
	semesterTable             = "ma_semesters"
	semesterTranslationsTable = "ma_semester_translations"

	semesterFromJoin = semesterTable + ` s
	LEFT JOIN ` + semesterTranslationsTable + ` t
	  ON t.semester_id = s.semester_id
	  AND t.language = ?
	  AND t.status = ?
	  AND t.deleted_dt IS NULL`

	semesterColumns = `s.id, s.semester_id,
		COALESCE(t.name, s.name) AS name,
		COALESCE(t.description, s.description, '') AS description,
		s.image_key, s.display_order, s.note,
		s.semester_status, s.status,
		s.create_id, s.create_dt, s.modify_id, s.modify_dt`

	semesterActiveWhere = `s.status IN (?) AND s.deleted_dt IS NULL`
)

func semesterJoinArgs(lang enum.LanguageType) []any {
	return []any{lang, enum.StatusActive}
}

func semesterActiveArgs() []any {
	return []any{enum.StatusActive}
}

type SemesterRepository struct {
	db database.Executor
}

func NewSemesterRepository(db database.Executor) semester.IRepository {
	return &SemesterRepository{db: db}
}

func scanSemester(s database.RowScanner) (*models.SemesterModel, error) {
	var m models.SemesterModel
	if err := s.Scan(&m.Id, &m.SemesterId, &m.Name, &m.Description, &m.ImageKey,
		&m.DisplayOrder, &m.Note, &m.SemesterStatus, &m.Status,
		&m.CreateId, &m.CreateDt, &m.ModifyId, &m.ModifyDt); err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *SemesterRepository) ListSemesters(ctx context.Context, params *semester.ListSemestersParams) ([]*semester.Semester, *pagination.Pagination, error) {
	lang := params.Language
	if lang == "" {
		lang = enum.LanguageTypeVietnamese
	}

	var total int64
	countQuery := `SELECT COUNT(*) FROM ` + semesterFromJoin +
		` WHERE ` + semesterActiveWhere
	countArgs := append(semesterJoinArgs(lang), semesterActiveArgs()...)
	if err := r.db.QueryRow(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, nil, fmt.Errorf("semester repo count: %w", err)
	}

	pg := pagination.NewPagination(params.Page, params.Limit, total)

	listArgs := append(semesterJoinArgs(lang), semesterActiveArgs()...)
	listArgs = append(listArgs, pg.Size, pg.Skip)
	query := `SELECT ` + semesterColumns + ` FROM ` + semesterFromJoin +
		` WHERE ` + semesterActiveWhere +
		` ORDER BY s.display_order ASC LIMIT ? OFFSET ?`
	rows, err := r.db.Query(ctx, query, listArgs...)
	if err != nil {
		return nil, nil, fmt.Errorf("semester repo list: %w", err)
	}
	defer rows.Close()

	var semesters []*semester.Semester
	for rows.Next() {
		m, err := scanSemester(rows)
		if err != nil {
			return nil, nil, fmt.Errorf("semester repo scan row: %w", err)
		}
		semesters = append(semesters, ModelToDomainSemester(m))
	}
	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("semester repo rows iteration: %w", err)
	}

	return semesters, pg, nil
}

// ListSemestersByIds — see program_repository's equivalent for rationale.
func (r *SemesterRepository) ListSemestersByIds(ctx context.Context, ids []uuid.UUID, lang enum.LanguageType) ([]*semester.Semester, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	if lang == "" {
		lang = enum.LanguageTypeVietnamese
	}

	placeholders := make([]string, len(ids))
	args := append(semesterJoinArgs(lang), semesterActiveArgs()...)
	for i, id := range ids {
		placeholders[i] = "?"
		args = append(args, id)
	}

	query := `SELECT ` + semesterColumns + ` FROM ` + semesterFromJoin +
		` WHERE ` + semesterActiveWhere +
		` AND s.semester_id IN (` + strings.Join(placeholders, ",") + `)`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("semester repo list by ids: %w", err)
	}
	defer rows.Close()

	var semesters []*semester.Semester
	for rows.Next() {
		m, err := scanSemester(rows)
		if err != nil {
			return nil, fmt.Errorf("semester repo scan row: %w", err)
		}
		semesters = append(semesters, ModelToDomainSemester(m))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("semester repo rows iteration: %w", err)
	}
	return semesters, nil
}

func ModelToDomainSemester(m *models.SemesterModel) *semester.Semester {
	s := semester.NewSemester()
	s.SetId(m.Id)
	s.SetSemesterId(m.SemesterId)
	s.SetName(m.Name)
	s.SetDescription(m.Description)
	s.SetImageKey(m.ImageKey)
	s.SetDisplayOrder(m.DisplayOrder)
	s.SetNote(m.Note)
	s.SetSemesterStatus(m.SemesterStatus)
	s.SetStatus(m.Status)
	s.SetCreateId(m.CreateId)
	s.SetCreateDt(mtime.MathTime{Time: m.CreateDt})
	s.SetModifyId(m.ModifyId)
	s.SetModifyDt(mtime.MathTime{Time: m.ModifyDt})
	return s
}
