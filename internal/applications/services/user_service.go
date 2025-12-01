package services

import (
	"context"
	"database/sql"
	"fmt"

	"math-ai.com/math-ai/internal/applications/dto"
	"math-ai.com/math-ai/internal/applications/validators"
	"math-ai.com/math-ai/internal/core/di/repositories"
	di "math-ai.com/math-ai/internal/core/di/services"
	"math-ai.com/math-ai/internal/shared/constant/status"
	err_svc "math-ai.com/math-ai/internal/shared/error"
	"math-ai.com/math-ai/internal/shared/utils/pagination"
)

type UserService struct {
	validator   validators.IUserValidator
	repo        repositories.IUserRepository
	loginRepo   repositories.ILoginRepository
	profileRepo repositories.IProfileRepository
}

func NewUserService(
	validator validators.IUserValidator,
	repo repositories.IUserRepository,
	loginRepo repositories.ILoginRepository,
	profileRepo repositories.IProfileRepository,
) di.IUserService {
	return &UserService{
		validator:   validator,
		repo:        repo,
		loginRepo:   loginRepo,
		profileRepo: profileRepo,
	}
}

func (s *UserService) ListUsers(ctx context.Context, req *dto.ListUserRequest) (status.Code, []*dto.UserResponse, *pagination.Pagination, error) {
	params := repositories.ListUsersParams{
		Search:    req.Search,
		Page:      req.Page,
		Limit:     req.Limit,
		OrderBy:   req.OrderBy,
		OrderDesc: req.OrderDesc,
		TakeAll:   req.TakeAll,
	}

	users, pagination, err := s.repo.List(ctx, params)
	if err != nil {
		return status.INTERNAL, nil, nil, err
	}

	if len(users) == 0 {
		return status.SUCCESS, []*dto.UserResponse{}, pagination, nil
	}

	res := make([]*dto.UserResponse, len(users))

	for i, user := range users {
		userRes := dto.UserResponseFromDomain(user)
		res[i] = &userRes
	}

	return status.SUCCESS, res, pagination, nil
}

func (s *UserService) GetUserByLoginName(ctx context.Context, loginName string) (status.Code, *dto.UserResponse, error) {
	user, err := s.repo.GetUserByLoginName(ctx, loginName)
	if err != nil {
		return status.INTERNAL, nil, err
	}

	res := dto.UserResponseFromDomain(user)

	return status.SUCCESS, &res, nil
}

func (s *UserService) GetUserByID(ctx context.Context, uid string) (status.Code, *dto.UserResponse, error) {
	user, err := s.repo.FindByID(ctx, uid)
	if err != nil {
		return status.INTERNAL, nil, err
	}
	if user == nil {
		return status.USER_NOT_FOUND, nil, err_svc.ErrUserNotFound
	}

	res := dto.UserResponseFromDomain(user)

	return status.SUCCESS, &res, nil
}

func (s *UserService) GetUserByEmail(ctx context.Context, email string) (status.Code, *dto.UserResponse, error) {
	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return status.INTERNAL, nil, err
	}
	if user == nil {
		return status.USER_NOT_FOUND, nil, err_svc.ErrUserNotFound
	}

	res := dto.UserResponseFromDomain(user)

	return status.SUCCESS, &res, nil
}

func (s *UserService) CreateUser(ctx context.Context, req *dto.CreateUserRequest) (status.Code, *dto.UserResponse, error) {
	for _, aka := range []string{req.Email, req.Phone} {
		if aka == "" {
			continue // Skip empty aliases
		}
		existingUser, err := s.repo.GetUserByLoginName(ctx, aka)
		if err != nil {
			return status.INTERNAL, nil, err
		}

		if existingUser != nil {
			if existingUser.Email() == req.Email {
				return status.USER_EMAIL_ALREADY_EXISTS, nil, err_svc.ErrEmailAlreadyExists
			} else if existingUser.Phone() == req.Phone {
				return status.USER_PHONE_ALREADY_EXISTS, nil, err_svc.ErrPhoneAlreadyExists
			}
		}
	}

	createUserDomain := dto.BuildUserDomainForCreate(req)
	handler := func(tx *sql.Tx) error {
		// Create the user
		_, err := s.repo.Create(ctx, tx, createUserDomain)
		if err != nil {
			return fmt.Errorf("failed to create user in transaction: %v", err)
		}

		// Store aliases
		for _, aka := range []string{createUserDomain.Email(), createUserDomain.Phone()} {
			if aka == "" {
				continue // Skip empty aliases
			}

			createAliasDomain := dto.BuildAliasDomain(createUserDomain.ID(), aka)
			if err := s.repo.StoreUserAlias(ctx, tx, createAliasDomain); err != nil {
				return fmt.Errorf("failed to store user alias in transaction: %v", err)
			}
		}

		// Store login
		createLoginDomain := dto.BuildLoginDomain(createUserDomain.ID(), createUserDomain.Password())
		if err := s.loginRepo.StoreLogin(ctx, tx, createLoginDomain); err != nil {
			return fmt.Errorf("failed to store user login in transaction: %v", err)
		}

		return nil
	}

	// Store login
	err := s.repo.DoTransaction(ctx, handler)
	if err != nil {
		return status.INTERNAL, nil, err
	}

	user, err := s.repo.FindByID(ctx, createUserDomain.ID())
	if err != nil {
		return status.INTERNAL, nil, err
	}

	res := dto.UserResponseFromDomain(user)

	return status.SUCCESS, &res, nil
}

func (s *UserService) UpdateUser(ctx context.Context, req *dto.UpdateUserRequest) (status.Code, *dto.UserResponse, error) {
	// statusCode, err := ValidateUpdateUserRequest(req)
	// if err != nil {
	// 	return statusCode, nil, err
	// }

	userDomain := dto.BuildUserDomainForUpdate(req)
	_, err := s.repo.Update(ctx, userDomain)
	if err != nil {
		return status.INTERNAL, nil, err
	}

	user, err := s.repo.FindByID(ctx, userDomain.ID())
	if err != nil {
		return status.INTERNAL, nil, err
	}

	res := dto.UserResponseFromDomain(user)

	return status.SUCCESS, &res, nil
}

func (s *UserService) DeleteUser(ctx context.Context, req *dto.DeleteUserRequest) (status.Code, error) {
	if statusCode, err := s.validator.ValidateDeleteUserRequest(req); err != nil {
		return statusCode, err
	}

	handler := func(tx *sql.Tx) error {
		// Delete users
		err := s.repo.Delete(ctx, tx, req.UID)
		if err != nil {
			return fmt.Errorf("failed to create user in transaction: %v", err)
		}

		// Delete user aliases
		err = s.repo.DeleteUserAlias(ctx, tx, req.UID)
		if err != nil {
			return fmt.Errorf("failed to delete user aliases in transaction: %v", err)
		}

		// Delete user logins
		err = s.loginRepo.DeleteLogin(ctx, tx, req.UID)
		if err != nil {
			return fmt.Errorf("failed to delete user logins in transaction: %v", err)
		}

		// Delete user profile
		err = s.profileRepo.DeleteByUID(ctx, tx, req.UID)
		if err != nil {
			return fmt.Errorf("failed to delete user profile in transaction: %v", err)
		}

		return nil
	}

	err := s.repo.DoTransaction(ctx, handler)
	if err != nil {
		return status.INTERNAL, err
	}
	return status.SUCCESS, nil
}

func (s *UserService) ForceDeleteUser(ctx context.Context, req *dto.DeleteUserRequest) (status.Code, error) {
	if statusCode, err := s.validator.ValidateDeleteUserRequest(req); err != nil {
		return statusCode, err
	}

	handler := func(tx *sql.Tx) error {
		// Delete users
		err := s.repo.ForceDelete(ctx, tx, req.UID)
		if err != nil {
			return fmt.Errorf("failed to create user in transaction: %v", err)
		}

		// Delete user aliases
		err = s.repo.ForceDeleteUserAlias(ctx, tx, req.UID)
		if err != nil {
			return fmt.Errorf("failed to delete user aliases in transaction: %v", err)
		}

		// Delete user logins
		err = s.loginRepo.ForceDeleteLogin(ctx, tx, req.UID)
		if err != nil {
			return fmt.Errorf("failed to delete user logins in transaction: %v", err)
		}

		// Delete user profile
		err = s.profileRepo.ForceDeleteByUID(ctx, tx, req.UID)
		if err != nil {
			return fmt.Errorf("failed to delete user profile in transaction: %v", err)
		}

		return nil
	}

	err := s.repo.DoTransaction(ctx, handler)
	if err != nil {
		return status.INTERNAL, err
	}
	return status.SUCCESS, nil
}
