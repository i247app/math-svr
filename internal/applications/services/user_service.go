package services

import (
	"context"

	dto "math-ai.com/math-ai/internal/applications/dto/user"
	"math-ai.com/math-ai/internal/core/di/repositories"
	di "math-ai.com/math-ai/internal/core/di/services"
	"math-ai.com/math-ai/internal/shared/constant/status"
	"math-ai.com/math-ai/internal/shared/utils/pagination"
)

type UserService struct {
	repo repositories.IUserRepository
}

func NewUserService(repo repositories.IUserRepository) di.IUserService {
	return &UserService{
		repo: repo,
	}
}

func (s *UserService) ListUsers(ctx context.Context, dto *dto.ListUserRequest) (status.Code, []*dto.UserResponse, *pagination.Pagination, error) {
	return 0, nil, nil, nil
}

func (s *UserService) GetUserByLoginName(ctx context.Context, loginName string) (status.Code, *dto.UserResponse, error) {
	return 0, nil, nil
}

func (s *UserService) GetUserByID(ctx context.Context, id string) (status.Code, *dto.UserResponse, error) {
	return 0, nil, nil
}
func (s *UserService) GetUserByEmail(ctx context.Context, email string) (status.Code, *dto.UserResponse, error) {
	return 0, nil, nil
}

func (s *UserService) CreateUser(ctx context.Context, req *dto.CreateUserRequest) (status.Code, *dto.UserResponse, error) {
	return 0, nil, nil
}

func (s *UserService) UpdateUser(ctx context.Context, req *dto.UpdateUserRequest) (status.Code, *dto.UserResponse, error) {
	return 0, nil, nil
}

func (s *UserService) DeleteUser(ctx context.Context, uid string) (status.Code, error) {
	return 0, nil
}

func (s *UserService) ForceDeleteUser(ctx context.Context, uid string) (status.Code, error) {
	return 0, nil
}
