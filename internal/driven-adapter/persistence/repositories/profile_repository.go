package repositories

import (
	"context"
	"database/sql"
	"fmt"

	di "math-ai.com/math-ai/internal/core/di/repositories"
	domain "math-ai.com/math-ai/internal/core/domain/profile"
	"math-ai.com/math-ai/internal/driven-adapter/persistence/models"
	"math-ai.com/math-ai/internal/driven-adapter/persistence/queries"
	"math-ai.com/math-ai/internal/shared/constant/enum"
	"math-ai.com/math-ai/internal/shared/db"
	mathtime "math-ai.com/math-ai/internal/shared/utils/time"
)

const (
	profileTableName = "ma_profiles"
)

type profileRepository struct {
	BaseRepository // Embed BaseRepository for common operations
}

func NewProfileRepository(database db.IDatabase) di.IProfileRepository {
	return &profileRepository{
		BaseRepository: NewBaseRepository(database),
	}
}

// scanProfile is a reusable helper method to scan profile data from a row
func (r *profileRepository) scanProfile(scanner Scanner) (*domain.Profile, error) {
	var p models.ProfileModel
	err := scanner.Scan(
		&p.ID, &p.UID, &p.Name, &p.Email, &p.Phone, &p.AvatarKey, &p.Dob, &p.Grade, &p.Semester, &p.Status,
		&p.CreateID, &p.CreateDT, &p.ModifyID, &p.ModifyDT,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("scan error: %v", err)
	}

	return domain.BuildProfileDomainFromModel(&p), nil
}

// FindByID retrieves a profile by ID with user information.
// Now uses query constants from queries package
func (r *profileRepository) FindByID(ctx context.Context, id string) (*domain.Profile, error) {
	row := r.db.QueryRow(ctx, nil, queries.ProfileFindByID, id)
	return r.scanProfile(row)
}

// FindByUID retrieves a profile by user ID with user information.
// Now uses query constants from queries package
func (r *profileRepository) FindByUID(ctx context.Context, uid string) (*domain.Profile, error) {
	row := r.db.QueryRow(ctx, nil, queries.ProfileFindByUID, uid)
	return r.scanProfile(row)
}

// Create inserts a new profile into the database.
// Now uses query constants from queries package
func (r *profileRepository) Create(ctx context.Context, tx *sql.Tx, profile *domain.Profile) (int64, error) {
	result, err := r.db.Exec(ctx, tx, queries.ProfileInsert,
		profile.ID(),
		profile.UID(),
		profile.GradeID(),
		profile.SemesterID(),
		enum.StatusActive,
		mathtime.Now(),
		mathtime.Now(),
	)
	if err != nil {
		return 0, fmt.Errorf("failed to create profile: %v", err)
	}

	return result.RowsAffected()
}

// Update modifies an existing profile in the database.
// Now uses query constants from queries package
func (r *profileRepository) Update(ctx context.Context, tx *sql.Tx, profile *domain.Profile) (int64, error) {
	result, err := r.db.Exec(ctx, tx, queries.ProfileUpdate,
		PrepareForUpdate(profile.GradeID()),
		PrepareForUpdate(profile.SemesterID()),
		PrepareForUpdate(profile.Status()),
		mathtime.Now(),
		profile.UID(),
	)
	if err != nil {
		return 0, fmt.Errorf("failed to update profile: %v", err)
	}

	return result.RowsAffected()
}

// DeleteByUID performs a soft delete of a profile by user ID.
// Now uses query constants from queries package
func (r *profileRepository) DeleteByUID(ctx context.Context, tx *sql.Tx, uid string) error {
	_, err := r.db.Exec(ctx, tx, queries.ProfileSoftDeleteByUID, mathtime.Now(), mathtime.Now(), uid)
	if err != nil {
		return fmt.Errorf("failed to delete profile: %v", err)
	}

	return nil
}

// ForceDeleteByUID permanently deletes a profile by user ID.
// Now uses query constants from queries package
func (r *profileRepository) ForceDeleteByUID(ctx context.Context, tx *sql.Tx, uid string) error {
	_, err := r.db.Exec(ctx, tx, queries.ProfileForceDeleteByUID, uid)
	if err != nil {
		return fmt.Errorf("failed to force delete profile: %v", err)
	}

	return nil
}
