package user_helper

import (
	"context"
	"database/sql"
	"fmt"

	"math-ai.com/math-ai/internal/applications/dto"
	di "math-ai.com/math-ai/internal/core/di/repositories"
	domain "math-ai.com/math-ai/internal/core/domain/user"
	"math-ai.com/math-ai/internal/shared/constant/status"
	err_svc "math-ai.com/math-ai/internal/shared/error"
)

// UserCreator handles complex user creation with transactions
type UserCreator struct {
	userRepo    di.IUserRepository
	loginRepo   di.ILoginRepository
	profileRepo di.IProfileRepository
}

// NewUserCreator creates a new UserCreator instance
func NewUserCreator(
	userRepo di.IUserRepository,
	loginRepo di.ILoginRepository,
	profileRepo di.IProfileRepository,
) *UserCreator {
	return &UserCreator{
		userRepo:    userRepo,
		loginRepo:   loginRepo,
		profileRepo: profileRepo,
	}
}

// CheckDuplicateUser checks if a user with the same email or phone already exists
func (u *UserCreator) CheckDuplicateUser(ctx context.Context, email, phone string) (status.Code, error) {
	for _, aka := range []string{email, phone} {
		if aka == "" {
			continue // Skip empty aliases
		}

		existingUser, err := u.userRepo.GetUserByLoginName(ctx, aka)
		if err != nil {
			return status.FAIL, err
		}

		if existingUser != nil {
			if existingUser.Email() == email {
				return status.USER_EMAIL_ALREADY_EXISTS, err_svc.ErrEmailAlreadyExists
			} else if existingUser.Phone() == phone {
				return status.USER_PHONE_ALREADY_EXISTS, err_svc.ErrPhoneAlreadyExists
			}
		}
	}

	return status.SUCCESS, nil
}

// CreateWithTransaction creates a user with aliases and login in a transaction
func (u *UserCreator) CreateWithTransaction(ctx context.Context, userDomain *domain.User) error {
	handler := func(tx *sql.Tx) error {
		// Create the user
		_, err := u.userRepo.Create(ctx, tx, userDomain)
		if err != nil {
			return fmt.Errorf("failed to create user in transaction: %v", err)
		}

		// Store aliases
		for _, aka := range []string{userDomain.Email(), userDomain.Phone()} {
			if aka == "" {
				continue // Skip empty aliases
			}

			createAliasDomain := dto.BuildAliasDomain(userDomain.ID(), aka)
			if err := u.userRepo.StoreUserAlias(ctx, tx, createAliasDomain); err != nil {
				return fmt.Errorf("failed to store user alias in transaction: %v", err)
			}
		}

		// Store login
		createLoginDomain := dto.BuildLoginDomain(userDomain.ID(), userDomain.Password())
		if err := u.loginRepo.StoreLogin(ctx, tx, createLoginDomain); err != nil {
			return fmt.Errorf("failed to store user login in transaction: %v", err)
		}

		// Store profile
		createProfileDomain := dto.BuildProfileDomainForCreate(&dto.CreateProfileRequest{
			UID:        userDomain.ID(),
			GradeID:    userDomain.GradeID(),
			SemesterID: userDomain.SemesterID(),
		})

		if _, err := u.profileRepo.Create(ctx, tx, createProfileDomain); err != nil {
			return fmt.Errorf("failed to store user profile in transaction: %v", err)
		}

		return nil
	}

	return u.userRepo.DoTransaction(ctx, handler)
}
