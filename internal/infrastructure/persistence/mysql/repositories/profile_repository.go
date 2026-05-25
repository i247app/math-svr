package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"math-ai.com/math-ai/internal/domain/profile"
	mtime "math-ai.com/math-ai/internal/domain/shared/time"
	"math-ai.com/math-ai/internal/infrastructure/database"
	"math-ai.com/math-ai/internal/infrastructure/persistence/mysql/models"
	"math-ai.com/math-ai/internal/shared/enum"
)

const (
	profileTable = "ma_profiles"

	profileColumns = `p.id, p.profile_id, p.user_id, p.name, p.avatar_key, p.dob,
		p.program_id, p.grade_id, p.semester_id, p.is_default, p.note, p.profile_status, p.status,
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
	if err := s.Scan(&m.Id, &m.ProfileId, &m.UserId, &m.Name, &m.AvatarKey, &m.Dob,
		&m.ProgramId, &m.GradeId, &m.SemesterId, &m.IsDefault, &m.Note, &m.ProfileStatus, &m.Status,
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

func (r *ProfileRepository) FindByProfileId(ctx context.Context, profileId string) (*profile.Profile, error) {
	return r.findOneBy(ctx, "p.profile_id = ?", profileId)
}

func (r *ProfileRepository) ListByUserId(ctx context.Context, userId string) ([]*profile.Profile, error) {
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

// ListAvatarKeysByUserId returns every non-null avatar_key for the user's
// profiles, regardless of status. Used by force-delete to drive S3 cleanup
// after the DB cascade — soft-deleted profiles still own S3 objects that need
// to go.
func (r *ProfileRepository) ListAvatarKeysByUserId(ctx context.Context, userId string) ([]string, error) {
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
			(profile_id, user_id, name, avatar_key, dob, program_id, grade_id, semester_id, is_default, note, profile_status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := r.db.Exec(ctx, query,
		p.ProfileId(), p.UserId(), p.Name(), p.AvatarKey(), p.Dob(),
		p.ProgramId(), p.GradeId(), p.SemesterId(), p.IsDefault(), p.Note(), p.ProfileStatus())
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
		SET name        = COALESCE(?, name),
			dob         = COALESCE(?, dob),
			program_id  = COALESCE(?, program_id),
			grade_id    = COALESCE(?, grade_id),
			semester_id = COALESCE(?, semester_id),
			is_default  = COALESCE(?, is_default),
			note        = COALESCE(?, note),
			avatar_key  = COALESCE(?, avatar_key)
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

	if _, err := r.db.Exec(ctx, query,
		nameArg, dobArg, p.ProgramId(), p.GradeId(), p.SemesterId(), p.IsDefault(), p.Note(), p.AvatarKey(), p.ProfileId()); err != nil {
		return fmt.Errorf("profile repo update: %w", err)
	}
	return nil
}

func (r *ProfileRepository) UpdateAvatarKey(ctx context.Context, profileId string, avatarKey string) error {
	query := `UPDATE ` + profileTable + ` SET avatar_key = ? WHERE profile_id = ?`
	if _, err := r.db.Exec(ctx, query, avatarKey, profileId); err != nil {
		return fmt.Errorf("profile repo update avatar key: %w", err)
	}
	return nil
}

func (r *ProfileRepository) MarkStatusByProfileId(ctx context.Context, profileId string, profileStatus string) error {
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

func (r *ProfileRepository) SoftDelete(ctx context.Context, profileId string) error {
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

func (r *ProfileRepository) ForceDelete(ctx context.Context, profileId string) error {
	query := `
		DELETE FROM ` + profileTable + `
		WHERE profile_id = ?
	`
	if _, err := r.db.Exec(ctx, query, profileId); err != nil {
		return fmt.Errorf("profile repo force delete: %w", err)
	}
	return nil
}

func (r *ProfileRepository) SoftDeleteByUserId(ctx context.Context, userId string) error {
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

func (r *ProfileRepository) ForceDeleteByUserId(ctx context.Context, userId string) error {
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
	p.SetUserId(m.UserId)
	p.SetName(m.Name)
	p.SetAvatarKey(m.AvatarKey)
	if m.Dob != nil {
		p.SetDob(mtime.MathTime{Time: *m.Dob})
	}
	p.SetProgramId(m.ProgramId)
	p.SetGradeId(m.GradeId)
	p.SetSemesterId(m.SemesterId)
	p.SetIsDefault(m.IsDefault)
	p.SetNote(m.Note)
	p.SetProfileStatus(m.ProfileStatus)
	p.SetStatus(m.Status)
	p.SetCreateId(m.CreateId)
	p.SetCreateDt(mtime.MathTime{Time: m.CreateDt})
	p.SetModifyId(m.ModifyId)
	p.SetModifyDt(mtime.MathTime{Time: m.ModifyDt})
	return p
}
