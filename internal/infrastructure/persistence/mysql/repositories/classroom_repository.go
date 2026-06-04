package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"math-ai.com/math-ai/internal/domain/classroom"
	mtime "math-ai.com/math-ai/internal/domain/shared/time"
	"math-ai.com/math-ai/internal/infrastructure/database"
	"math-ai.com/math-ai/internal/infrastructure/persistence/mysql/models"
	"math-ai.com/math-ai/internal/shared/enum"
	"math-ai.com/math-ai/internal/shared/pagination"
)

const (
	classroomTable = "ma_classrooms"

	classroomColumns = `c.id, c.classroom_id, c.owner_profile_id, c.name, c.description,
		c.school_id, c.grade_id,
		c.classroom_code, c.classroom_code_expires_dt,
		c.max_members, c.member_count, c.student_count, c.teacher_count,
		c.cover_key, c.note,
		c.classroom_status, c.status,
		c.create_id, c.create_dt, c.modify_id, c.modify_dt`

	// classroomActiveWhere excludes system-inactive and business-DELETED
	// rows but keeps ARCHIVED rows visible — archived classrooms are
	// read-only history that the UI still wants to render.
	classroomActiveWhere = `c.status = ? AND c.deleted_dt IS NULL
		AND (c.classroom_status IS NULL OR c.classroom_status != ?)`
)

func classroomActiveArgs() []any {
	return []any{enum.StatusActive, enum.ClassroomStatusTypeDeleted}
}

type ClassroomRepository struct {
	db database.Executor
}

func NewClassroomRepository(db database.Executor) classroom.IRepository {
	return &ClassroomRepository{db: db}
}

func scanClassroom(s database.RowScanner) (*models.ClassroomModel, error) {
	var m models.ClassroomModel
	if err := s.Scan(&m.Id, &m.ClassroomId, &m.OwnerProfileId, &m.Name, &m.Description,
		&m.SchoolId, &m.GradeId,
		&m.ClassroomCode, &m.ClassroomCodeExpiresDt,
		&m.MaxMembers, &m.MemberCount, &m.StudentCount, &m.TeacherCount,
		&m.CoverKey, &m.Note,
		&m.ClassroomStatus, &m.Status,
		&m.CreateId, &m.CreateDt, &m.ModifyId, &m.ModifyDt); err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *ClassroomRepository) findOneBy(ctx context.Context, where string, args ...any) (*classroom.Classroom, error) {
	fullArgs := append(classroomActiveArgs(), args...)
	query := `SELECT ` + classroomColumns + ` FROM ` + classroomTable + ` c WHERE ` +
		classroomActiveWhere + ` AND (` + where + `)`

	m, err := scanClassroom(r.db.QueryRow(ctx, query, fullArgs...))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("classroom repo find (%s): %w", where, err)
	}
	return ModelToDomainClassroom(m), nil
}

func (r *ClassroomRepository) findBareById(ctx context.Context, id int64) (*classroom.Classroom, error) {
	args := append(classroomActiveArgs(), id)
	query := `SELECT ` + classroomColumns + ` FROM ` + classroomTable + ` c WHERE ` +
		classroomActiveWhere + ` AND c.id = ?`

	m, err := scanClassroom(r.db.QueryRow(ctx, query, args...))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("classroom repo find bare by id: %w", err)
	}
	return ModelToDomainClassroom(m), nil
}

func (r *ClassroomRepository) FindByClassroomId(ctx context.Context, classroomId int64) (*classroom.Classroom, error) {
	return r.findOneBy(ctx, "c.classroom_id = ?", classroomId)
}

func (r *ClassroomRepository) FindByClassroomCode(ctx context.Context, code string) (*classroom.Classroom, error) {
	return r.findOneBy(ctx, "c.classroom_code = ?", code)
}

func (r *ClassroomRepository) ListClassrooms(ctx context.Context, params *classroom.ListClassroomsParams) ([]*classroom.Classroom, *pagination.Pagination, error) {
	filterWhere, filterArgs, joinClause := buildClassroomListFilter(params)

	baseFrom := classroomTable + ` c` + joinClause

	countArgs := append(classroomActiveArgs(), filterArgs...)
	countQuery := `SELECT COUNT(DISTINCT c.id) FROM ` + baseFrom + ` WHERE ` +
		classroomActiveWhere + filterWhere

	var total int64
	if err := r.db.QueryRow(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, nil, fmt.Errorf("classroom repo count: %w", err)
	}

	listArgs := append(classroomActiveArgs(), filterArgs...)
	query := `SELECT DISTINCT ` + classroomColumns + ` FROM ` + baseFrom + ` WHERE ` +
		classroomActiveWhere + filterWhere +
		` ORDER BY c.modify_dt DESC, c.id DESC`

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
		return nil, nil, fmt.Errorf("classroom repo list: %w", err)
	}
	defer rows.Close()

	var out []*classroom.Classroom
	for rows.Next() {
		m, err := scanClassroom(rows)
		if err != nil {
			return nil, nil, fmt.Errorf("classroom repo scan row: %w", err)
		}
		out = append(out, ModelToDomainClassroom(m))
	}
	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("classroom repo rows iteration: %w", err)
	}
	return out, pg, nil
}

// mergeProgramIDFilters returns the OR-union of the singular ProgramId
// pointer and the ProgramIds slice, with blanks dropped and duplicates
// removed. Used by the list-filter builder so callers can supply either
// or both without the SQL needing to know which path they took.
func mergeProgramIDFilters(single *int64, slice []int64) []int64 {
	seen := make(map[int64]struct{}, len(slice)+1)
	out := make([]int64, 0, len(slice)+1)
	add := func(id int64) {
		if id == 0 {
			return
		}
		if _, ok := seen[id]; ok {
			return
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	if single != nil {
		add(*single)
	}
	for _, id := range slice {
		add(id)
	}
	return out
}

// buildClassroomListFilter composes the optional WHERE fragments and the
// LEFT JOIN we need when filtering by ProfileId (membership). Keeping the
// join LEFT lets us still apply the classroom-level active filter; the
// JOIN side's active filter is mixed into the where clause.
func buildClassroomListFilter(params *classroom.ListClassroomsParams) (string, []any, string) {
	if params == nil {
		return "", nil, ""
	}
	var (
		clause string
		args   []any
		join   string
	)
	if params.ProfileId != nil && *params.ProfileId != 0 {
		join += ` JOIN ma_classroom_members mem ON mem.classroom_id = c.classroom_id`
		clause += ` AND mem.profile_id = ? AND mem.status = ?
			AND mem.deleted_dt IS NULL
			AND (mem.member_status IS NULL OR mem.member_status = ?)`
		args = append(args,
			*params.ProfileId,
			enum.StatusActive,
			enum.ClassroomMemberStatusTypeActive,
		)
	}
	if params.OwnerProfileId != nil && *params.OwnerProfileId != 0 {
		clause += ` AND c.owner_profile_id = ?`
		args = append(args, *params.OwnerProfileId)
	}
	if params.SchoolId != nil && *params.SchoolId != 0 {
		clause += ` AND c.school_id = ?`
		args = append(args, *params.SchoolId)
	}
	// Program filter — union of the singular ProgramId and the
	// ProgramIds slice. Matches via EXISTS against the junction table
	// so a classroom is included if it carries any of the requested
	// programs (OR semantics, per the redesign agreement).
	programIDs := mergeProgramIDFilters(params.ProgramId, params.ProgramIds)
	if len(programIDs) > 0 {
		placeholders := make([]string, len(programIDs))
		for i := range programIDs {
			placeholders[i] = "?"
		}
		clause += ` AND EXISTS (
			SELECT 1 FROM ma_classroom_programs cp
			WHERE cp.classroom_id = c.classroom_id
				AND cp.status = ? AND cp.deleted_dt IS NULL
				AND cp.program_id IN (` + strings.Join(placeholders, ",") + `)
		)`
		args = append(args, enum.StatusActive)
		for _, id := range programIDs {
			args = append(args, id)
		}
	}
	if params.GradeId != nil && *params.GradeId != 0 {
		clause += ` AND c.grade_id = ?`
		args = append(args, *params.GradeId)
	}
	if params.Search != nil {
		needle := strings.TrimSpace(*params.Search)
		if needle != "" {
			clause += ` AND (c.name LIKE ? OR c.classroom_code LIKE ?) `
			args = append(args, "%"+needle+"%", "%"+needle+"%")
		}
	}
	if !params.IncludeArchived {
		clause += ` AND (c.classroom_status IS NULL OR c.classroom_status != ?)`
		args = append(args, enum.ClassroomStatusTypeArchived)
	}
	return clause, args, join
}

func (r *ClassroomRepository) ListClassroomsByIds(ctx context.Context, ids []int64) ([]*classroom.Classroom, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	placeholders := make([]string, len(ids))
	args := classroomActiveArgs()
	for i, id := range ids {
		placeholders[i] = "?"
		args = append(args, id)
	}

	query := `SELECT ` + classroomColumns + ` FROM ` + classroomTable + ` c WHERE ` +
		classroomActiveWhere +
		` AND c.classroom_id IN (` + strings.Join(placeholders, ",") + `)`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("classroom repo list by ids: %w", err)
	}
	defer rows.Close()

	var out []*classroom.Classroom
	for rows.Next() {
		m, err := scanClassroom(rows)
		if err != nil {
			return nil, fmt.Errorf("classroom repo scan row: %w", err)
		}
		out = append(out, ModelToDomainClassroom(m))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("classroom repo rows iteration: %w", err)
	}
	return out, nil
}

func (r *ClassroomRepository) Create(ctx context.Context, c *classroom.Classroom) (*classroom.Classroom, error) {
	var expiresArg any
	if c.ClassroomCodeExpiresDt().IsValid() {
		expiresArg = c.ClassroomCodeExpiresDt().Time
	}

	query := `
		INSERT INTO ` + classroomTable + `
			(classroom_id, owner_profile_id, name, description,
			 school_id, grade_id,
			 classroom_code, classroom_code_expires_dt,
			 max_members, member_count, student_count, teacher_count,
			 cover_key, note,
			 classroom_status, create_id, create_dt, modify_dt)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	result, err := r.db.Exec(ctx, query,
		c.ClassroomId(), c.OwnerProfileId(), c.Name(), c.Description(),
		c.SchoolId(), c.GradeId(),
		c.ClassroomCode(), expiresArg,
		c.MaxMembers(), c.MemberCount(), c.StudentCount(), c.TeacherCount(),
		c.CoverKey(), c.Note(),
		c.ClassroomStatus(), c.CreateId(), mtime.Now().Time, mtime.Now().Time)
	if err != nil {
		return nil, fmt.Errorf("classroom repo create: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("classroom repo last insert id: %w", err)
	}
	return r.findBareById(ctx, id)
}

// Update applies a partial patch. Non-nullable scalar fields use a
// non-zero check; nullable pointer fields go through COALESCE(?, col)
// so passing nil leaves the column unchanged.
func (r *ClassroomRepository) Update(ctx context.Context, c *classroom.Classroom) error {
	var nameArg any
	if c.Name() != "" {
		nameArg = c.Name()
	}

	query := `
		UPDATE ` + classroomTable + `
		SET name        = COALESCE(?, name),
			description = COALESCE(?, description),
			school_id   = COALESCE(?, school_id),
			grade_id    = COALESCE(?, grade_id),
			max_members = COALESCE(?, max_members),
			cover_key   = COALESCE(?, cover_key),
			note        = COALESCE(?, note),
			modify_id   = COALESCE(?, modify_id),
			modify_dt   = ?
		WHERE classroom_id = ?
	`
	if _, err := r.db.Exec(ctx, query,
		nameArg, c.Description(), c.SchoolId(), c.GradeId(),
		c.MaxMembers(), c.CoverKey(), c.Note(), c.ModifyId(),
		mtime.Now().Time, c.ClassroomId()); err != nil {
		return fmt.Errorf("classroom repo update: %w", err)
	}
	return nil
}

// IncCounts applies signed deltas to the three counters in a single
// UPDATE. GREATEST(0, …) clamps each at zero so drift can never produce
// a negative counter on disk.
func (r *ClassroomRepository) IncCounts(ctx context.Context, classroomId int64, memberDelta, studentDelta, teacherDelta int64) error {
	query := `
		UPDATE ` + classroomTable + `
		SET member_count  = GREATEST(0, CAST(member_count  AS SIGNED) + ?),
			student_count = GREATEST(0, CAST(student_count AS SIGNED) + ?),
			teacher_count = GREATEST(0, CAST(teacher_count AS SIGNED) + ?),
			modify_dt     = ?
		WHERE classroom_id = ?
	`
	if _, err := r.db.Exec(ctx, query,
		memberDelta, studentDelta, teacherDelta,
		mtime.Now().Time, classroomId); err != nil {
		return fmt.Errorf("classroom repo inc counts: %w", err)
	}
	return nil
}

func (r *ClassroomRepository) UpdateClassroomCode(ctx context.Context, classroomId int64, code *string, expiresDt mtime.MathTime) error {
	var expiresArg any
	if expiresDt.IsValid() {
		expiresArg = expiresDt.Time
	}
	query := `
		UPDATE ` + classroomTable + `
		SET classroom_code             = ?,
			classroom_code_expires_dt  = ?,
			modify_dt               = ?
		WHERE classroom_id = ?
	`
	if _, err := r.db.Exec(ctx, query, code, expiresArg, mtime.Now().Time, classroomId); err != nil {
		return fmt.Errorf("classroom repo update invite code: %w", err)
	}
	return nil
}

func (r *ClassroomRepository) SetOwnerProfileId(ctx context.Context, classroomId, newOwnerProfileId int64) error {
	query := `
		UPDATE ` + classroomTable + `
		SET owner_profile_id = ?,
			modify_dt        = ?
		WHERE classroom_id = ?
	`
	if _, err := r.db.Exec(ctx, query, newOwnerProfileId, mtime.Now().Time, classroomId); err != nil {
		return fmt.Errorf("classroom repo set owner: %w", err)
	}
	return nil
}

func (r *ClassroomRepository) ArchiveByClassroomId(ctx context.Context, classroomId int64) error {
	query := `
		UPDATE ` + classroomTable + `
		SET classroom_status = ?,
			modify_dt        = ?
		WHERE classroom_id = ?
	`
	now := mtime.Now().Time
	if _, err := r.db.Exec(ctx, query, enum.ClassroomStatusTypeArchived, now, classroomId); err != nil {
		return fmt.Errorf("classroom repo archive: %w", err)
	}
	return nil
}

func (r *ClassroomRepository) RestoreByClassroomId(ctx context.Context, classroomId int64) error {
	query := `
		UPDATE ` + classroomTable + `
		SET classroom_status = ?,
			modify_dt        = ?
		WHERE classroom_id = ?
	`
	now := mtime.Now().Time
	if _, err := r.db.Exec(ctx, query, enum.ClassroomStatusTypeActive, now, classroomId); err != nil {
		return fmt.Errorf("classroom repo restore: %w", err)
	}
	return nil
}

func (r *ClassroomRepository) SoftDeleteByClassroomId(ctx context.Context, classroomId int64) error {
	query := `
		UPDATE ` + classroomTable + `
		SET classroom_status = ?,
			status           = ?,
			deleted_dt       = ?,
			modify_dt        = ?
		WHERE classroom_id = ?
	`
	now := mtime.Now().Time
	if _, err := r.db.Exec(ctx, query,
		enum.ClassroomStatusTypeDeleted, enum.StatusInactive, now, now, classroomId); err != nil {
		return fmt.Errorf("classroom repo soft delete: %w", err)
	}
	return nil
}

func (r *ClassroomRepository) ForceDeleteByClassroomId(ctx context.Context, classroomId int64) error {
	query := `DELETE FROM ` + classroomTable + ` WHERE classroom_id = ?`
	if _, err := r.db.Exec(ctx, query, classroomId); err != nil {
		return fmt.Errorf("classroom repo force delete: %w", err)
	}
	return nil
}

func ModelToDomainClassroom(m *models.ClassroomModel) *classroom.Classroom {
	c := classroom.NewClassroom()
	c.SetId(m.Id)
	c.SetClassroomId(m.ClassroomId)
	c.SetOwnerProfileId(m.OwnerProfileId)
	c.SetName(m.Name)
	c.SetDescription(m.Description)
	c.SetSchoolId(m.SchoolId)
	c.SetGradeId(m.GradeId)
	c.SetClassroomCode(m.ClassroomCode)
	if m.ClassroomCodeExpiresDt != nil {
		c.SetClassroomCodeExpiresDt(mtime.MathTime{Time: *m.ClassroomCodeExpiresDt})
	}
	c.SetMaxMembers(m.MaxMembers)
	c.SetMemberCount(m.MemberCount)
	c.SetStudentCount(m.StudentCount)
	c.SetTeacherCount(m.TeacherCount)
	c.SetCoverKey(m.CoverKey)
	c.SetNote(m.Note)
	c.SetClassroomStatus(m.ClassroomStatus)
	c.SetStatus(m.Status)
	c.SetCreateId(m.CreateId)
	c.SetCreateDt(mtime.MathTime{Time: m.CreateDt})
	c.SetModifyId(m.ModifyId)
	c.SetModifyDt(mtime.MathTime{Time: m.ModifyDt})
	return c
}
