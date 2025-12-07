package user_helper

import (
	"context"
	"database/sql"
	"fmt"

	"math-ai.com/math-ai/internal/core/di/repositories"
)

// UserDeleter handles complex user deletion with transactions
type UserDeleter struct {
	userRepo    repositories.IUserRepository
	loginRepo   repositories.ILoginRepository
	profileRepo repositories.IProfileRepository
}

// NewUserDeleter creates a new UserDeleter instance
func NewUserDeleter(
	userRepo repositories.IUserRepository,
	loginRepo repositories.ILoginRepository,
	profileRepo repositories.IProfileRepository,
) *UserDeleter {
	return &UserDeleter{
		userRepo:    userRepo,
		loginRepo:   loginRepo,
		profileRepo: profileRepo,
	}
}

// DeleteWithTransaction performs soft delete of user and related data
func (u *UserDeleter) DeleteWithTransaction(ctx context.Context, uid string) error {
	handler := func(tx *sql.Tx) error {
		// Delete users
		err := u.userRepo.Delete(ctx, tx, uid)
		if err != nil {
			return fmt.Errorf("failed to delete user in transaction: %v", err)
		}

		// Delete user aliases
		err = u.userRepo.DeleteUserAlias(ctx, tx, uid)
		if err != nil {
			return fmt.Errorf("failed to delete user aliases in transaction: %v", err)
		}

		// Delete user logins
		err = u.loginRepo.DeleteLogin(ctx, tx, uid)
		if err != nil {
			return fmt.Errorf("failed to delete user logins in transaction: %v", err)
		}

		// Delete user profile
		err = u.profileRepo.DeleteByUID(ctx, tx, uid)
		if err != nil {
			return fmt.Errorf("failed to delete user profile in transaction: %v", err)
		}

		return nil
	}

	return u.userRepo.DoTransaction(ctx, handler)
}

// ForceDeleteWithTransaction performs hard delete of user and related data
func (u *UserDeleter) ForceDeleteWithTransaction(ctx context.Context, uid string) error {
	handler := func(tx *sql.Tx) error {
		// Force delete users
		err := u.userRepo.ForceDelete(ctx, tx, uid)
		if err != nil {
			return fmt.Errorf("failed to force delete user in transaction: %v", err)
		}

		// Force delete user aliases
		err = u.userRepo.ForceDeleteUserAlias(ctx, tx, uid)
		if err != nil {
			return fmt.Errorf("failed to force delete user aliases in transaction: %v", err)
		}

		// Force delete user logins
		err = u.loginRepo.ForceDeleteLogin(ctx, tx, uid)
		if err != nil {
			return fmt.Errorf("failed to force delete user logins in transaction: %v", err)
		}

		// Force delete user profile
		err = u.profileRepo.ForceDeleteByUID(ctx, tx, uid)
		if err != nil {
			return fmt.Errorf("failed to force delete user profile in transaction: %v", err)
		}

		return nil
	}

	return u.userRepo.DoTransaction(ctx, handler)
}
