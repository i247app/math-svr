package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"slices"
	"strings"

	"math-ai.com/math-ai/internal/domain/school"
	"math-ai.com/math-ai/internal/domain/shared/mtime"
	"math-ai.com/math-ai/internal/infrastructure/database"
	"math-ai.com/math-ai/internal/infrastructure/persistence/mysql/models"
	"math-ai.com/math-ai/internal/shared/enum"
	"math-ai.com/math-ai/internal/shared/pagination"
)

const (
	schoolTable = "ma_schools"

	schoolColumns = `s.id, s.school_id, s.name, s.description, s.image_key,
		s.district, s.province, s.rpt_flg, s.kwords, s.note,
		s.school_status, s.status,
		s.create_id, s.create_dt, s.modify_id, s.modify_dt`

	schoolActiveWhere = `s.status IN (?) AND s.deleted_dt IS NULL`
)

func schoolActiveArgs() []any {
	return []any{enum.StatusActive}
}

type SchoolRepository struct {
	db database.Executor
}

func NewSchoolRepository(db database.Executor) school.IRepository {
	return &SchoolRepository{db: db}
}

func scanSchool(s database.RowScanner) (*models.SchoolModel, error) {
	var m models.SchoolModel
	if err := s.Scan(&m.Id, &m.SchoolId, &m.Name, &m.Description, &m.ImageKey,
		&m.District, &m.Province, &m.RptFlg, &m.Kwords, &m.Note,
		&m.SchoolStatus, &m.Status,
		&m.CreateId, &m.CreateDt, &m.ModifyId, &m.ModifyDt); err != nil {
		return nil, err
	}
	return &m, nil
}

// findOneBy is the single-row read helper. `where` is a package-controlled
// SQL fragment; args supply placeholders. schoolActiveWhere is appended last
// so every read excludes soft-deleted and inactive rows.
func (r *SchoolRepository) findOneBy(ctx context.Context, where string, args ...any) (*school.School, error) {
	fullArgs := slices.Concat(args, schoolActiveArgs())
	query := `SELECT ` + schoolColumns + ` FROM ` + schoolTable + ` s WHERE (` +
		where + `) AND ` + schoolActiveWhere

	m, err := scanSchool(r.db.QueryRow(ctx, query, fullArgs...))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("school repo find (%s): %w", where, err)
	}
	return ModelToDomainSchool(m), nil
}

func (r *SchoolRepository) findBareById(ctx context.Context, id int64) (*school.School, error) {
	args := slices.Concat([]any{id}, schoolActiveArgs())
	query := `SELECT ` + schoolColumns + ` FROM ` + schoolTable + ` s WHERE (s.id = ?) AND ` +
		schoolActiveWhere

	m, err := scanSchool(r.db.QueryRow(ctx, query, args...))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("school repo find bare by id: %w", err)
	}
	return ModelToDomainSchool(m), nil
}

func (r *SchoolRepository) FindBySchoolId(ctx context.Context, schoolId int64) (*school.School, error) {
	return r.findOneBy(ctx, "s.school_id = ?", schoolId)
}

func (r *SchoolRepository) ListSchools(ctx context.Context, params *school.ListSchoolsParams) ([]*school.School, *pagination.Pagination, error) {
	filterWhere, filterArgs := buildSchoolListFilterClause(params)

	countArgs := slices.Concat(filterArgs, schoolActiveArgs())
	countQuery := `SELECT COUNT(*) FROM ` + schoolTable + ` s` +
		whereActive(filterWhere, schoolActiveWhere)

	var total int64
	if err := r.db.QueryRow(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, nil, fmt.Errorf("school repo count: %w", err)
	}

	listArgs := slices.Concat(filterArgs, schoolActiveArgs())
	query := `SELECT ` + schoolColumns + ` FROM ` + schoolTable + ` s` +
		whereActive(filterWhere, schoolActiveWhere) +
		` ORDER BY s.name ASC, s.id ASC`

	var pg *pagination.Pagination
	if params == nil || !params.TakeAll {
		page := int64(1)
		limit := int64(20)
		if params != nil {
			page = params.Page
			limit = params.Limit
		}
		pg = pagination.NewPagination(page, limit, total)
		query += ` LIMIT ? OFFSET ?`
		listArgs = append(listArgs, pg.Size, pg.Skip)
	} else {
		pg = pagination.NewPagination(1, total, total)
	}

	rows, err := r.db.Query(ctx, query, listArgs...)
	if err != nil {
		return nil, nil, fmt.Errorf("school repo list: %w", err)
	}
	defer rows.Close()

	var schools []*school.School
	for rows.Next() {
		m, err := scanSchool(rows)
		if err != nil {
			return nil, nil, fmt.Errorf("school repo scan row: %w", err)
		}
		schools = append(schools, ModelToDomainSchool(m))
	}
	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("school repo rows iteration: %w", err)
	}
	return schools, pg, nil
}

// buildSchoolListFilterClause appends optional narrowing predicates. Search
// is a case-insensitive substring match against name, district, and
// province; District and Province narrow by exact match. Keeping each
// clause guarded preserves placeholder/arg ordering parity.
func buildSchoolListFilterClause(params *school.ListSchoolsParams) (string, []any) {
	if params == nil {
		return "", nil
	}
	var (
		clause string
		args   []any
	)
	if params.Search != nil {
		needle := strings.TrimSpace(*params.Search)
		if needle != "" {
			clause += ` AND (s.name LIKE ? OR s.district LIKE ? OR s.province LIKE ?)`
			like := "%" + needle + "%"
			args = append(args, like, like, like)
		}
	}
	if params.District != nil && strings.TrimSpace(*params.District) != "" {
		clause += ` AND s.district = ?`
		args = append(args, strings.TrimSpace(*params.District))
	}
	if params.Province != nil && strings.TrimSpace(*params.Province) != "" {
		clause += ` AND s.province = ?`
		args = append(args, strings.TrimSpace(*params.Province))
	}
	if len(params.SchoolIds) > 0 {
		placeholders := make([]string, len(params.SchoolIds))
		for i, id := range params.SchoolIds {
			placeholders[i] = "?"
			args = append(args, id)
		}
		clause += ` AND s.school_id IN (` + strings.Join(placeholders, ",") + `)`
	}
	return clause, args
}

// ListSchoolsByIds resolves a batch of schools in a single round-trip.
// Caller (typically the profile service composing a response) is
// responsible for keying the result by SchoolId() — order is not preserved
// because the IN-clause makes no ordering guarantee.
func (r *SchoolRepository) ListSchoolsByIds(ctx context.Context, ids []int64) ([]*school.School, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	placeholders := make([]string, len(ids))
	args := make([]any, 0, len(ids)+1)
	for i, id := range ids {
		placeholders[i] = "?"
		args = append(args, id)
	}

	args = append(args, schoolActiveArgs()...)

	query := `SELECT ` + schoolColumns + ` FROM ` + schoolTable + ` s WHERE (s.school_id IN (` +
		strings.Join(placeholders, ",") + `)) AND ` + schoolActiveWhere

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("school repo list by ids: %w", err)
	}
	defer rows.Close()

	var schools []*school.School
	for rows.Next() {
		m, err := scanSchool(rows)
		if err != nil {
			return nil, fmt.Errorf("school repo scan row: %w", err)
		}
		schools = append(schools, ModelToDomainSchool(m))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("school repo rows iteration: %w", err)
	}
	return schools, nil
}

func (r *SchoolRepository) Create(ctx context.Context, s *school.School) (*school.School, error) {
	query := `
		INSERT INTO ` + schoolTable + `
			(school_id, name, description, image_key, district, province, rpt_flg, kwords, note, school_status, create_id, create_dt, modify_dt)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	result, err := r.db.Exec(ctx, query,
		s.SchoolId(), s.Name(), s.Description(), s.ImageKey(),
		s.District(), s.Province(), s.RptFlg(), s.Kwords(), s.Note(), s.SchoolStatus(), s.CreateId(), mtime.Now().Time, mtime.Now().Time)
	if err != nil {
		return nil, fmt.Errorf("school repo create: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("school repo last insert id: %w", err)
	}
	return r.findBareById(ctx, id)
}

// Update applies a partial update using COALESCE(?, col) for nullable
// columns and a non-zero check for non-nullable ones. Pointer fields are
// passed through directly; the *string nil case becomes SQL NULL inside
// COALESCE and leaves the existing column unchanged.
func (r *SchoolRepository) Update(ctx context.Context, s *school.School) error {
	var nameArg any
	if s.Name() != "" {
		nameArg = s.Name()
	}

	query := `
		UPDATE ` + schoolTable + `
		SET name        = COALESCE(?, name),
			description = COALESCE(?, description),
			image_key   = COALESCE(?, image_key),
			district    = COALESCE(?, district),
			province    = COALESCE(?, province),
			rpt_flg     = COALESCE(?, rpt_flg),
			kwords      = COALESCE(?, kwords),
			note        = COALESCE(?, note),
			modify_id   = COALESCE(?, modify_id),
			modify_dt   = ?
		WHERE school_id = ?
	`
	if _, err := r.db.Exec(ctx, query,
		nameArg, s.Description(), s.ImageKey(),
		s.District(), s.Province(), s.RptFlg(), s.Kwords(), s.Note(), s.ModifyId(),
		mtime.Now().Time, s.SchoolId()); err != nil {
		return fmt.Errorf("school repo update: %w", err)
	}
	return nil
}

func (r *SchoolRepository) SoftDeleteBySchoolId(ctx context.Context, schoolId int64) error {
	query := `
		UPDATE ` + schoolTable + `
		SET school_status = ?,
			status        = ?,
			deleted_dt    = ?,
			modify_dt     = ?
		WHERE school_id = ?
	`
	now := mtime.Now().Time
	if _, err := r.db.Exec(ctx, query,
		enum.SchoolStatusTypeDeleted, enum.StatusInactive, now, now, schoolId); err != nil {
		return fmt.Errorf("school repo soft delete: %w", err)
	}
	return nil
}

func (r *SchoolRepository) ForceDeleteBySchoolId(ctx context.Context, schoolId int64) error {
	query := `
		DELETE FROM ` + schoolTable + `
		WHERE school_id = ?
	`
	if _, err := r.db.Exec(ctx, query, schoolId); err != nil {
		return fmt.Errorf("school repo force delete: %w", err)
	}
	return nil
}

func ModelToDomainSchool(m *models.SchoolModel) *school.School {
	s := school.NewSchool()
	s.SetId(m.Id)
	s.SetSchoolId(m.SchoolId)
	s.SetName(m.Name)
	s.SetDescription(m.Description)
	s.SetImageKey(m.ImageKey)
	s.SetDistrict(m.District)
	s.SetProvince(m.Province)
	s.SetRptFlg(m.RptFlg)
	s.SetKwords(m.Kwords)
	s.SetNote(m.Note)
	s.SetSchoolStatus(m.SchoolStatus)
	s.SetStatus(m.Status)
	s.SetCreateId(m.CreateId)
	s.SetCreateDt(mtime.MathTime{Time: m.CreateDt})
	s.SetModifyId(m.ModifyId)
	s.SetModifyDt(mtime.MathTime{Time: m.ModifyDt})
	return s
}
