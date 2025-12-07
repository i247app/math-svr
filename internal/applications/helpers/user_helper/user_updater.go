package user_helper

import (
	"context"
	"database/sql"

	di "math-ai.com/math-ai/internal/core/di/repositories"
	profileDomain "math-ai.com/math-ai/internal/core/domain/profile"
	domain "math-ai.com/math-ai/internal/core/domain/user"
)

// UserUpdater handles complex user update operations with transactions
type UserUpdater struct {
	userRepo    di.IUserRepository
	profileRepo di.IProfileRepository
}

// NewUserUpdater creates a new UserUpdater instance
func NewUserUpdater(
	userRepo di.IUserRepository,
	profileRepo di.IProfileRepository,
) *UserUpdater {
	return &UserUpdater{
		userRepo:    userRepo,
		profileRepo: profileRepo,
	}
}

// UpdateWithTransaction updates user and optionally profile in a transaction
// avatarKey parameter is used if the avatar was changed (can be nil if no change)
func (u *UserUpdater) UpdateWithTransaction(
	ctx context.Context,
	userDomain *domain.User,
	profileDomain *profileDomain.Profile,
) error {
	handler := func(tx *sql.Tx) error {

		// Update user
		_, updateErr := u.userRepo.Update(ctx, userDomain)
		if updateErr != nil {
			return updateErr
		}

		// Update profile if provided
		if profileDomain != nil {
			_, profileUpdateErr := u.profileRepo.Update(ctx, profileDomain)
			if profileUpdateErr != nil {
				return profileUpdateErr
			}
		}

		return nil
	}

	return u.userRepo.DoTransaction(ctx, handler)
}
