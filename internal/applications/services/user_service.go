package services

import (
	"context"
	"database/sql"
	"fmt"

	"math-ai.com/math-ai/internal/applications/dto"
	"math-ai.com/math-ai/internal/core/di/repositories"
	di "math-ai.com/math-ai/internal/core/di/services"
	"math-ai.com/math-ai/internal/shared/constant/status"
	err_svc "math-ai.com/math-ai/internal/shared/error"
	"math-ai.com/math-ai/internal/shared/utils/pagination"
)

type UserService struct {
	repo      repositories.IUserRepository
	loginRepo repositories.ILoginRepository
}

func NewUserService(
	repo repositories.IUserRepository,
	loginRepo repositories.ILoginRepository,
) di.IUserService {
	return &UserService{
		repo:      repo,
		loginRepo: loginRepo,
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
	user, err := s.repo.CreateUserWithAssociations(ctx, handler, createUserDomain.ID())
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

func (s *UserService) DeleteUser(ctx context.Context, uid string) (status.Code, error) {
	return 0, nil
}

func (s *UserService) ForceDeleteUser(ctx context.Context, uid string) (status.Code, error) {
	return 0, nil
}
