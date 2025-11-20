package di

import (
	"context"

	dto "math-ai.com/math-ai/internal/applications/dto/user"
	"math-ai.com/math-ai/internal/shared/constant/status"
	"math-ai.com/math-ai/internal/shared/utils/pagination"
)

type IUserService interface {
	ListUsers(ctx context.Context, dto *dto.ListUserRequest) (status.Code, []*dto.UserResponse, *pagination.Pagination, error)
	GetUserByLoginName(ctx context.Context, loginName string) (status.Code, *dto.UserResponse, error)
	GetUserByID(ctx context.Context, id string) (status.Code, *dto.UserResponse, error)
	GetUserByEmail(ctx context.Context, email string) (status.Code, *dto.UserResponse, error)
	CreateUser(ctx context.Context, req *dto.CreateUserRequest) (status.Code, *dto.UserResponse, error)
	UpdateUser(ctx context.Context, req *dto.UpdateUserRequest) (status.Code, *dto.UserResponse, error)
	DeleteUser(ctx context.Context, uid string) (status.Code, error)
	ForceDeleteUser(ctx context.Context, uid string) (status.Code, error)
}
