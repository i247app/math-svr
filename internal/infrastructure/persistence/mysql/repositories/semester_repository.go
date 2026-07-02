package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"math-ai.com/math-ai/internal/domain/semester"
	"math-ai.com/math-ai/internal/domain/shared/mtime"
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

	// LEFT JOIN so a semester without a translation in the requested
	// language still surfaces with its base name / description.
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

	semesterActiveWhere = `s.status = ? AND s.deleted_dt IS NULL
		AND (s.semester_status IS NULL OR s.semester_status != ?)`

	semesterBareActiveWhere = `s.status = ? AND s.deleted_dt IS NULL
		AND (s.semester_status IS NULL OR s.semester_status != ?)`
)

func semesterJoinArgs(lang enum.LanguageType) []any {
	if lang == "" {
		lang = enum.LanguageTypeVietnamese
	}
	return []any{lang, enum.StatusActive}
}

func semesterActiveArgs() []any {
	return []any{enum.StatusActive, enum.StatusInactive}
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

// findOneBy is the single-row read helper. `where` is a package-controlled
// SQL fragment; args supply placeholders. semesterActiveWhere is prepended
// so every read excludes soft-deleted and inactive rows.
func (r *SemesterRepository) findOneBy(ctx context.Context, lang enum.LanguageType, where string, args ...any) (*semester.Semester, error) {
	fullArgs := append(semesterJoinArgs(lang), semesterActiveArgs()...)
	fullArgs = append(fullArgs, args...)
	query := `SELECT ` + semesterColumns + ` FROM ` + semesterFromJoin +
		` WHERE ` + semesterActiveWhere + ` AND (` + where + `)`

	m, err := scanSemester(r.db.QueryRow(ctx, query, fullArgs...))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("semester repo find (%s): %w", where, err)
	}
	return ModelToDomainSemester(m), nil
}

// findBareById hydrates a semester right after INSERT using its surrogate
// id. The translation rows haven't been written yet at that point, so we
// skip the LEFT JOIN and just read the base name/description.
func (r *SemesterRepository) findBareById(ctx context.Context, id int64) (*semester.Semester, error) {
	query := `SELECT s.id, s.semester_id, s.name,
			COALESCE(s.description, '') AS description,
			s.image_key, s.display_order, s.note, s.semester_status, s.status,
			s.create_id, s.create_dt, s.modify_id, s.modify_dt
		FROM ` + semesterTable + ` s
		WHERE ` + semesterBareActiveWhere + ` AND s.id = ?`
	args := append(semesterActiveArgs(), id)

	m, err := scanSemester(r.db.QueryRow(ctx, query, args...))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("semester repo find bare by id: %w", err)
	}
	return ModelToDomainSemester(m), nil
}

func (r *SemesterRepository) FindBySemesterId(ctx context.Context, semesterId int64, language enum.LanguageType) (*semester.Semester, error) {
	return r.findOneBy(ctx, language, "s.semester_id = ?", semesterId)
}

func (r *SemesterRepository) ListSemesters(ctx context.Context, params *semester.ListSemestersParams) ([]*semester.Semester, *pagination.Pagination, error) {
	lang := params.Language
	if lang == "" {
		lang = enum.LanguageTypeVietnamese
	}

	countArgs := append(semesterJoinArgs(lang), semesterActiveArgs()...)
	countQuery := `SELECT COUNT(*) FROM ` + semesterFromJoin +
		` WHERE ` + semesterActiveWhere
	var total int64
	if err := r.db.QueryRow(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, nil, fmt.Errorf("semester repo count: %w", err)
	}

	listArgs := append(semesterJoinArgs(lang), semesterActiveArgs()...)
	query := `SELECT ` + semesterColumns + ` FROM ` + semesterFromJoin +
		` WHERE ` + semesterActiveWhere +
		` ORDER BY s.display_order ASC, s.id ASC`

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
func (r *SemesterRepository) ListSemestersByIds(ctx context.Context, ids []int64, lang enum.LanguageType) ([]*semester.Semester, error) {
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

func (r *SemesterRepository) Create(ctx context.Context, s *semester.Semester) (*semester.Semester, error) {
	query := `
		INSERT INTO ` + semesterTable + `
			(semester_id, name, description, image_key, display_order, note, semester_status, create_dt, modify_dt)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	// description column is nullable TEXT; persist as NULL when empty
	// so reads coalesce cleanly.
	var description any
	if s.Description() != "" {
		description = s.Description()
	}

	result, err := r.db.Exec(ctx, query,
		s.SemesterId(), s.Name(), description, s.ImageKey(),
		s.DisplayOrder(), s.Note(), s.SemesterStatus(), mtime.Now().Time, mtime.Now().Time)
	if err != nil {
		return nil, fmt.Errorf("semester repo create: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("semester repo last insert id: %w", err)
	}
	return r.findBareById(ctx, id)
}

// Update applies a partial update using COALESCE(?, col) for every
// nullable / non-zero column. display_order is int8 — a literal zero is
// a legal value, so it is always written and not coalesced.
func (r *SemesterRepository) Update(ctx context.Context, s *semester.Semester) error {
	var name, description, imageKey any
	if s.Name() != "" {
		name = s.Name()
	}
	if s.Description() != "" {
		description = s.Description()
	}
	if s.ImageKey() != nil && *s.ImageKey() != "" {
		imageKey = *s.ImageKey()
	}

	query := `
		UPDATE ` + semesterTable + `
		SET name          = COALESCE(?, name),
			description   = COALESCE(?, description),
			image_key     = COALESCE(?, image_key),
			display_order = ?,
			note          = COALESCE(?, note),
			modify_dt     = ?
		WHERE semester_id = ?
	`
	if _, err := r.db.Exec(ctx, query,
		name, description, imageKey,
		s.DisplayOrder(), s.Note(), mtime.Now().Time, s.SemesterId()); err != nil {
		return fmt.Errorf("semester repo update: %w", err)
	}
	return nil
}

func (r *SemesterRepository) SoftDeleteBySemesterId(ctx context.Context, semesterId int64) error {
	query := `
		UPDATE ` + semesterTable + `
		SET semester_status = ?,
			status          = ?,
			deleted_dt      = ?,
			modify_dt       = ?
		WHERE semester_id = ?
	`
	now := mtime.Now().Time
	if _, err := r.db.Exec(ctx, query,
		enum.StatusInactive, enum.StatusInactive, now, now, semesterId); err != nil {
		return fmt.Errorf("semester repo soft delete: %w", err)
	}
	return nil
}

func (r *SemesterRepository) ForceDeleteBySemesterId(ctx context.Context, semesterId int64) error {
	query := `
		DELETE FROM ` + semesterTable + `
		WHERE semester_id = ?
	`
	if _, err := r.db.Exec(ctx, query, semesterId); err != nil {
		return fmt.Errorf("semester repo force delete: %w", err)
	}
	return nil
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
