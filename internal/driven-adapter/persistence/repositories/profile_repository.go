package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	di "math-ai.com/math-ai/internal/core/di/repositories"
	domain "math-ai.com/math-ai/internal/core/domain/profile"
	"math-ai.com/math-ai/internal/driven-adapter/persistence/models"
	"math-ai.com/math-ai/internal/shared/constant/enum"
	"math-ai.com/math-ai/internal/shared/db"
)

type profileRepository struct {
	db db.IDatabase
}

func NewProfileRepository(db db.IDatabase) di.IProfileRepository {
	return &profileRepository{
		db: db,
	}
}

// FindByID retrieves a profile by ID with user information.
func (r *profileRepository) FindByID(ctx context.Context, id string) (*domain.Profile, error) {
	query := `
		SELECT p.id, p.uid, u.name, u.email, u.phone, u.avatar_key, u.dob, g.label, s.name, p.status,
		p.create_id, p.create_dt, p.modify_id, p.modify_dt
		FROM profiles p
		INNER JOIN users u ON p.uid = u.id
		INNER JOIN terms s ON p.term_id = s.id
		INNER JOIN grades g ON p.grade_id = g.id
		WHERE p.id = ? AND p.deleted_dt IS NULL AND u.deleted_dt IS NULL
	`

	result := r.db.QueryRow(ctx, nil, query, id)

	var p models.ProfileModel
	err := result.Scan(
		&p.ID, &p.UID, &p.Name, &p.Email, &p.Phone, &p.AvatarKey, &p.Dob, &p.Grade, &p.Term, &p.Status,
		&p.CreateID, &p.CreateDT, &p.ModifyID, &p.ModifyDT,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("scan error: %v", err)
	}

	profile := domain.BuildProfileDomainFromModel(&p)

	return profile, nil
}

// FindByUID retrieves a profile by user ID with user information.
func (r *profileRepository) FindByUID(ctx context.Context, uid string) (*domain.Profile, error) {
	query := `
		SELECT p.id, p.uid, u.name, u.email, u.phone, u.avatar_key, u.dob, g.label, s.name, p.status,
		p.create_id, p.create_dt, p.modify_id, p.modify_dt
		FROM profiles p
		INNER JOIN users u ON p.uid = u.id
		INNER JOIN terms s ON p.term_id = s.id
		INNER JOIN grades g ON p.grade_id = g.id
		WHERE p.uid = ? AND p.deleted_dt IS NULL AND u.deleted_dt IS NULL
	`

	result := r.db.QueryRow(ctx, nil, query, uid)

	var p models.ProfileModel
	err := result.Scan(
		&p.ID, &p.UID, &p.Name, &p.Email, &p.Phone, &p.AvatarKey, &p.Dob, &p.Grade, &p.Term, &p.Status,
		&p.CreateID, &p.CreateDT, &p.ModifyID, &p.ModifyDT,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("scan error: %v", err)
	}

	profile := domain.BuildProfileDomainFromModel(&p)

	return profile, nil
}

// Create inserts a new profile into the database.
func (r *profileRepository) Create(ctx context.Context, tx *sql.Tx, profile *domain.Profile) (int64, error) {
	query := `
		INSERT INTO profiles (id, uid, grade_id, term_id ,status)
		VALUES (?, ?, ?, ?, ?)
	`
	result, err := r.db.Exec(ctx, tx, query,
		profile.ID(),
		profile.UID(),
		profile.GradeID(),
		profile.TermID(),
		enum.StatusActive,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to create profile: %v", err)
	}

	return result.RowsAffected()
}

// Update modifies an existing profile in the database.
func (r *profileRepository) Update(ctx context.Context, profile *domain.Profile) (int64, error) {
	var queryBuilder strings.Builder
	args := []interface{}{}

	queryBuilder.WriteString("UPDATE profiles SET ")
	updates := []string{}

	if profile.GradeID() != "" {
		updates = append(updates, "grade_id = ?")
		args = append(args, profile.GradeID())
	}

	if profile.TermID() != "" {
		updates = append(updates, "term_id = ?")
		args = append(args, profile.TermID())
	}

	if profile.Status() != "" {
		updates = append(updates, "status = ?")
		args = append(args, profile.Status())
	}

	updates = append(updates, "modify_dt = ?")
	args = append(args, time.Now().UTC())

	if len(updates) == 0 {
		return 0, fmt.Errorf("no fields to update")
	}

	queryBuilder.WriteString(strings.Join(updates, ", "))
	queryBuilder.WriteString(" WHERE uid = ? AND deleted_dt IS NULL")
	args = append(args, profile.UID())

	result, err := r.db.Exec(ctx, nil, queryBuilder.String(), args...)
	if err != nil {
		return 0, fmt.Errorf("failed to update profile: %v", err)
	}

	return result.RowsAffected()
}

func (r *profileRepository) DeleteByUID(ctx context.Context, tx *sql.Tx, uid string) error {
	query := `
		UPDATE profiles
		SET deleted_dt = ?,
			modify_dt = ?
		WHERE uid = ? AND deleted_dt IS NULL
	`
	_, err := r.db.Exec(ctx, tx, query, time.Now().UTC(), time.Now().UTC(), uid)
	if err != nil {
		return fmt.Errorf("failed to delete profile: %v", err)
	}

	return nil
}

func (r *profileRepository) ForceDeleteByUID(ctx context.Context, tx *sql.Tx, uid string) error {
	query := `
		DELETE FROM profiles
		WHERE uid = ?
	`
	_, err := r.db.Exec(ctx, tx, query, uid)
	if err != nil {
		return fmt.Errorf("failed to force delete profile: %v", err)
	}

	return nil
}
