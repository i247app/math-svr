package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"math-ai.com/math-ai/internal/domain/profile"
	mtime "math-ai.com/math-ai/internal/domain/shared/time"
	"math-ai.com/math-ai/internal/infrastructure/database"
	"math-ai.com/math-ai/internal/infrastructure/persistence/mysql/models"
	"math-ai.com/math-ai/internal/shared/enum"
	"math-ai.com/math-ai/internal/shared/pagination"
)

const (
	profileTable = "ma_profiles"

	profileColumns = `p.id, p.profile_id, p.profile_code, p.user_id, p.name, p.role, p.avatar_key, p.dob,
		p.school_id, p.program_id, p.grade_id, p.semester_id, p.is_default,
		p.id_type, p.teacher_id, p.student_id,
		p.note, p.profile_status, p.status,
		p.create_id, p.create_dt, p.modify_id, p.modify_dt`

	profileActiveWhere = `p.status IN (?) AND p.deleted_dt IS NULL`
)

func profileActiveArgs() []any {
	return []any{enum.StatusActive}
}

type ProfileRepository struct {
	db database.Executor
}

func NewProfileRepository(db database.Executor) profile.IRepository {
	return &ProfileRepository{db: db}
}

func scanProfile(s database.RowScanner) (*models.ProfileModel, error) {
	var m models.ProfileModel
	if err := s.Scan(&m.Id, &m.ProfileId, &m.ProfileCode, &m.UserId, &m.Name, &m.Role, &m.AvatarKey, &m.Dob,
		&m.SchoolId, &m.ProgramId, &m.GradeId, &m.SemesterId, &m.IsDefault,
		&m.IdType, &m.TeacherId, &m.StudentId,
		&m.Note, &m.ProfileStatus, &m.Status,
		&m.CreateId, &m.CreateDt, &m.ModifyId, &m.ModifyDt); err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *ProfileRepository) findOneBy(ctx context.Context, where string, args ...any) (*profile.Profile, error) {
	fullArgs := append(profileActiveArgs(), args...)
	query := `SELECT ` + profileColumns + ` FROM ` + profileTable + ` p WHERE ` +
		profileActiveWhere + ` AND (` + where + `)`

	m, err := scanProfile(r.db.QueryRow(ctx, query, fullArgs...))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("profile repo find (%s): %w", where, err)
	}
	return ModelToDomainProfile(m), nil
}

func (r *ProfileRepository) findBareById(ctx context.Context, id int64) (*profile.Profile, error) {
	args := append(profileActiveArgs(), id)
	query := `SELECT ` + profileColumns + ` FROM ` + profileTable + ` p WHERE ` +
		profileActiveWhere + ` AND p.id = ?`

	m, err := scanProfile(r.db.QueryRow(ctx, query, args...))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("profile repo find bare by id: %w", err)
	}
	return ModelToDomainProfile(m), nil
}

func (r *ProfileRepository) FindByProfileId(ctx context.Context, profileId int64) (*profile.Profile, error) {
	return r.findOneBy(ctx, "p.profile_id = ?", profileId)
}

func (r *ProfileRepository) FindDefaultProfileByUserId(ctx context.Context, userId int64) (*profile.Profile, error) {
	return r.findOneBy(ctx, "p.user_id = ? AND p.is_default = true", userId)
}

func (r *ProfileRepository) FindByProfileCode(ctx context.Context, profileCode string) (*profile.Profile, error) {
	return r.findOneBy(ctx, "p.profile_code = ?", profileCode)
}

func (r *ProfileRepository) ListByUserId(ctx context.Context, userId int64) ([]*profile.Profile, error) {
	args := append(profileActiveArgs(), userId)
	query := `SELECT ` + profileColumns + ` FROM ` + profileTable + ` p WHERE ` +
		profileActiveWhere + ` AND p.user_id = ? ORDER BY p.id DESC`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("profile repo list by user id: %w", err)
	}
	defer rows.Close()

	var profiles []*profile.Profile
	for rows.Next() {
		m, err := scanProfile(rows)
		if err != nil {
			return nil, fmt.Errorf("profile repo scan row: %w", err)
		}
		profiles = append(profiles, ModelToDomainProfile(m))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("profile repo rows iteration: %w", err)
	}
	return profiles, nil
}

func (r *ProfileRepository) ListByProfileIds(ctx context.Context, profileIds []int64) ([]*profile.Profile, error) {
	if len(profileIds) == 0 {
		return nil, nil
	}
	// Build IN clause with placeholders: ? , ? , ?
	placeholders := make([]string, len(profileIds))
	// args := make([]any, len(profileIds))
	args := profileActiveArgs()
	// args = append(args, enum.StatusActive)

	for i, id := range profileIds {
		placeholders[i] = "?"
		args = append(args, id)
	}

	query := `SELECT ` + profileColumns + ` FROM ` + profileTable + ` p WHERE ` +
		profileActiveWhere + ` AND p.profile_id IN (` + strings.Join(placeholders, ", ") + `) ORDER BY p.id DESC`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("profile repo list by profile ids: %w", err)
	}
	defer rows.Close()

	var profiles []*profile.Profile
	for rows.Next() {
		m, err := scanProfile(rows)
		if err != nil {
			return nil, fmt.Errorf("profile repo scan row: %w", err)
		}
		profiles = append(profiles, ModelToDomainProfile(m))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("profile repo rows iteration: %w", err)
	}
	return profiles, nil
}

// ListProfiles returns a paginated slice of active profiles matching the
// provided filters. The filter shape mirrors classroom/user list — every
// predicate is optional and contributes to the WHERE clause only when set.
// Search runs against name with LIKE %?%. Pagination is computed against
// COUNT(DISTINCT p.id) so the totals stay correct when future filters
// introduce joins. TakeAll bypasses LIMIT/OFFSET for admin exports.
func (r *ProfileRepository) ListProfiles(ctx context.Context, params *profile.ListProfilesParams) ([]*profile.Profile, *pagination.Pagination, error) {
	filterWhere, filterArgs := buildProfileListFilter(params)

	countArgs := append(profileActiveArgs(), filterArgs...)
	countQuery := `SELECT COUNT(DISTINCT p.id) FROM ` + profileTable + ` p WHERE ` +
		profileActiveWhere + filterWhere

	var total int64
	if err := r.db.QueryRow(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, nil, fmt.Errorf("profile repo count: %w", err)
	}

	listArgs := append(profileActiveArgs(), filterArgs...)
	query := `SELECT ` + profileColumns + ` FROM ` + profileTable + ` p WHERE ` +
		profileActiveWhere + filterWhere +
		` ORDER BY p.id DESC`

	var pg *pagination.Pagination
	if params == nil || !params.TakeAll {
		page := int64(1)
		limit := int64(0)
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
		return nil, nil, fmt.Errorf("profile repo list: %w", err)
	}
	defer rows.Close()

	var profiles []*profile.Profile
	for rows.Next() {
		m, err := scanProfile(rows)
		if err != nil {
			return nil, nil, fmt.Errorf("profile repo scan row: %w", err)
		}
		profiles = append(profiles, ModelToDomainProfile(m))
	}
	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("profile repo rows iteration: %w", err)
	}
	return profiles, pg, nil
}

func buildProfileListFilter(params *profile.ListProfilesParams) (string, []any) {
	if params == nil {
		return "", nil
	}
	var (
		clause strings.Builder
		args   []any
	)
	if params.UserId != nil && *params.UserId != 0 {
		clause.WriteString(` AND p.user_id = ?`)
		args = append(args, *params.UserId)
	}
	if params.Role != nil && *params.Role != "" {
		clause.WriteString(` AND p.role = ?`)
		args = append(args, *params.Role)
	}
	if params.ProfileStatus != nil && *params.ProfileStatus != "" {
		clause.WriteString(` AND p.profile_status = ?`)
		args = append(args, *params.ProfileStatus)
	}
	if params.SchoolId != nil && *params.SchoolId != 0 {
		clause.WriteString(` AND p.school_id = ?`)
		args = append(args, *params.SchoolId)
	}
	if params.ProgramId != nil && *params.ProgramId != 0 {
		clause.WriteString(` AND p.program_id = ?`)
		args = append(args, *params.ProgramId)
	}
	if params.GradeId != nil && *params.GradeId != 0 {
		clause.WriteString(` AND p.grade_id = ?`)
		args = append(args, *params.GradeId)
	}
	if params.SemesterId != nil && *params.SemesterId != 0 {
		clause.WriteString(` AND p.semester_id = ?`)
		args = append(args, *params.SemesterId)
	}
	if params.IsDefault != nil {
		clause.WriteString(` AND p.is_default = ?`)
		args = append(args, *params.IsDefault)
	}
	if params.Search != nil {
		needle := strings.TrimSpace(*params.Search)
		if needle != "" {
			// OR'd LIKE on name + profile_code so a single needle
			// covers partial names, partial codes, and full codes
			// uniformly. utf8mb4_0900_ai_ci makes both predicates
			// case-insensitive without an UPPER() wrap. Args are
			// bound positionally — same wildcarded value twice keeps
			// the planner's choice opaque to the caller.
			clause.WriteString(` AND (p.name LIKE ? OR p.profile_code LIKE ?)`)
			wildcard := "%" + needle + "%"
			args = append(args, wildcard, wildcard)
		}
	}
	return clause.String(), args
}

// ListAvatarKeysByUserId returns every non-null avatar_key for the user's
// profiles, regardless of status. Used by force-delete to drive S3 cleanup
// after the DB cascade — soft-deleted profiles still own S3 objects that need
// to go.
func (r *ProfileRepository) ListAvatarKeysByUserId(ctx context.Context, userId int64) ([]string, error) {
	query := `SELECT avatar_key FROM ` + profileTable + ` WHERE user_id = ? AND avatar_key IS NOT NULL`

	rows, err := r.db.Query(ctx, query, userId)
	if err != nil {
		return nil, fmt.Errorf("profile repo list avatar keys by user id: %w", err)
	}
	defer rows.Close()

	var keys []string
	for rows.Next() {
		var key string
		if err := rows.Scan(&key); err != nil {
			return nil, fmt.Errorf("profile repo scan avatar key: %w", err)
		}
		if key != "" {
			keys = append(keys, key)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("profile repo avatar keys iteration: %w", err)
	}
	return keys, nil
}

func (r *ProfileRepository) Create(ctx context.Context, p *profile.Profile) (*profile.Profile, error) {
	query := `
		INSERT INTO ` + profileTable + `
			(profile_id, profile_code, user_id, name, role, avatar_key, dob, school_id, program_id, grade_id, semester_id, is_default,
			 id_type, teacher_id, student_id, note, profile_status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := r.db.Exec(ctx, query,
		p.ProfileId(), p.ProfileCode(), p.UserId(), p.Name(), p.Role(), p.AvatarKey(), p.Dob(),
		p.SchoolId(), p.ProgramId(), p.GradeId(), p.SemesterId(), p.IsDefault(),
		p.IdType(), p.TeacherId(), p.StudentId(), p.Note(), p.ProfileStatus())
	if err != nil {
		return nil, fmt.Errorf("profile repo create: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("profile repo last insert id: %w", err)
	}
	return r.findBareById(ctx, id)
}

func (r *ProfileRepository) Update(ctx context.Context, p *profile.Profile) error {
	query := `
		UPDATE ` + profileTable + `
		SET name           = COALESCE(?, name),
			role           = COALESCE(?, role),
			is_default     = COALESCE(?, is_default),
			dob            = COALESCE(?, dob),
			school_id      = COALESCE(?, school_id),
			program_id     = COALESCE(?, program_id),
			grade_id       = COALESCE(?, grade_id),
			semester_id    = COALESCE(?, semester_id),
			id_type        = COALESCE(?, id_type),
			teacher_id     = COALESCE(?, teacher_id),
			student_id     = COALESCE(?, student_id),
			profile_status = COALESCE(?, profile_status),
			note           = COALESCE(?, note),
			avatar_key     = COALESCE(?, avatar_key)
		WHERE profile_id = ?
	`

	var dobArg any
	if p.Dob().IsValid() {
		dobArg = p.Dob()
	}

	var nameArg any
	if p.Name() != "" {
		nameArg = p.Name()
	}

	var roleArg any
	if p.Role() != "" {
		roleArg = p.Role()
	}

	var isDefaultArg any
	if p.IsDefault() {
		isDefaultArg = p.IsDefault()
	}

	if _, err := r.db.Exec(ctx, query,
		nameArg, roleArg, isDefaultArg, dobArg, p.SchoolId(), p.ProgramId(), p.GradeId(),
		p.SemesterId(), p.IdType(), p.TeacherId(), p.StudentId(), p.ProfileStatus(),
		p.Note(), p.AvatarKey(), p.ProfileId()); err != nil {
		return fmt.Errorf("profile repo update: %w", err)
	}
	return nil
}

// SetSchoolId is a dedicated write path for assign/remove flows. Unlike
// Update (which uses COALESCE to leave existing values alone on nil),
// this always writes the column — a nil pointer becomes SQL NULL,
// effectively removing the school link. Keeping it separate avoids
// re-using Update with sentinel values to express "clear this column".
func (r *ProfileRepository) SetSchoolId(ctx context.Context, profileId int64, schoolId *int64) error {
	query := `
		UPDATE ` + profileTable + `
		SET school_id = ?,
			modify_dt = ?
		WHERE profile_id = ?
	`
	if _, err := r.db.Exec(ctx, query, schoolId, mtime.Now().Time, profileId); err != nil {
		return fmt.Errorf("profile repo set school id: %w", err)
	}
	return nil
}

func (r *ProfileRepository) UpdateAvatarKey(ctx context.Context, profileId int64, avatarKey string) error {
	query := `UPDATE ` + profileTable + ` SET avatar_key = ? WHERE profile_id = ?`
	if _, err := r.db.Exec(ctx, query, avatarKey, profileId); err != nil {
		return fmt.Errorf("profile repo update avatar key: %w", err)
	}
	return nil
}

func (r *ProfileRepository) MarkStatusByProfileId(ctx context.Context, profileId int64, profileStatus string) error {
	query := `
		UPDATE ` + profileTable + `
		SET profile_status = ?,
			modify_dt      = ?
		WHERE profile_id = ?
	`
	if _, err := r.db.Exec(ctx, query, profileStatus, mtime.Now().Time, profileId); err != nil {
		return fmt.Errorf("profile repo mark status: %w", err)
	}
	return nil
}

func (r *ProfileRepository) MarkDefaultByProfileId(ctx context.Context, userId int64, profileId int64) error {
	query := `
		UPDATE ` + profileTable + `
		SET is_default = CASE WHEN profile_id = ? THEN TRUE ELSE FALSE END,
			modify_dt    = ?
		WHERE user_id = ?
	`
	if _, err := r.db.Exec(ctx, query, profileId, mtime.Now().Time, userId); err != nil {
		return fmt.Errorf("profile repo mark is default: %w", err)
	}
	return nil
}

func (r *ProfileRepository) SoftDelete(ctx context.Context, profileId int64) error {
	query := `
		UPDATE ` + profileTable + `
		SET profile_status = ?,
			status = ?,
			deleted_dt = ?
		WHERE profile_id = ?
	`
	if _, err := r.db.Exec(ctx, query, enum.ProfileStatusTypeDeleted, enum.StatusInactive, mtime.Now().Time, profileId); err != nil {
		return fmt.Errorf("profile repo soft delete: %w", err)
	}
	return nil
}

func (r *ProfileRepository) ForceDelete(ctx context.Context, profileId int64) error {
	query := `
		DELETE FROM ` + profileTable + `
		WHERE profile_id = ?
	`
	if _, err := r.db.Exec(ctx, query, profileId); err != nil {
		return fmt.Errorf("profile repo force delete: %w", err)
	}
	return nil
}

func (r *ProfileRepository) SoftDeleteByUserId(ctx context.Context, userId int64) error {
	query := `
		UPDATE ` + profileTable + `
		SET profile_status = ?,
			status = ?,
			deleted_dt = ?
		WHERE user_id = ?
	`
	if _, err := r.db.Exec(ctx, query, enum.ProfileStatusTypeDeleted, enum.StatusInactive, mtime.Now().Time, userId); err != nil {
		return fmt.Errorf("profile repo soft delete by user id: %w", err)
	}
	return nil
}

func (r *ProfileRepository) ForceDeleteByUserId(ctx context.Context, userId int64) error {
	query := `
		DELETE FROM ` + profileTable + `
		WHERE user_id = ?
	`
	if _, err := r.db.Exec(ctx, query, userId); err != nil {
		return fmt.Errorf("profile repo force delete by user id: %w", err)
	}
	return nil
}

func ModelToDomainProfile(m *models.ProfileModel) *profile.Profile {
	p := profile.NewProfile()
	p.SetId(m.Id)
	p.SetProfileId(m.ProfileId)
	p.SetProfileCode(m.ProfileCode)
	p.SetUserId(m.UserId)
	p.SetName(m.Name)
	p.SetRole(m.Role)
	p.SetAvatarKey(m.AvatarKey)
	if m.Dob != nil {
		p.SetDob(mtime.MathTime{Time: *m.Dob})
	}
	p.SetSchoolId(m.SchoolId)
	p.SetProgramId(m.ProgramId)
	p.SetGradeId(m.GradeId)
	p.SetSemesterId(m.SemesterId)
	p.SetIsDefault(m.IsDefault)
	p.SetIdType(m.IdType)
	p.SetTeacherId(m.TeacherId)
	p.SetStudentId(m.StudentId)
	p.SetNote(m.Note)
	p.SetProfileStatus(m.ProfileStatus)
	p.SetStatus(m.Status)
	p.SetCreateId(m.CreateId)
	p.SetCreateDt(mtime.MathTime{Time: m.CreateDt})
	p.SetModifyId(m.ModifyId)
	p.SetModifyDt(mtime.MathTime{Time: m.ModifyDt})
	return p
}
