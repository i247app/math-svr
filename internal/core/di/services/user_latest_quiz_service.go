package di

import (
	"context"

	"math-ai.com/math-ai/internal/applications/dto"
	"math-ai.com/math-ai/internal/shared/constant/status"
	"math-ai.com/math-ai/internal/shared/utils/pagination"
)

type IUserLatestQuizService interface {
	ListQuizzes(ctx context.Context, req *dto.ListUserLatestQuizzesRequest) (status.Code, []*dto.UserLatestQuizResponse, *pagination.Pagination, error)
	GetQuiz(ctx context.Context, req *dto.GetUserLatestQuizRequest) (status.Code, *dto.UserLatestQuizResponse, error)
	GetQuizByUID(ctx context.Context, req *dto.GetUserLatestQuizByUIDRequest) (status.Code, *dto.UserLatestQuizResponse, error)
	CreateQuiz(ctx context.Context, req *dto.CreateUserLatestQuizRequest) (status.Code, *dto.UserLatestQuizResponse, error)
	UpdateQuiz(ctx context.Context, req *dto.UpdateUserLatestQuizRequest) (status.Code, *dto.UserLatestQuizResponse, error)
	DeleteQuiz(ctx context.Context, req *dto.DeleteUserLatestQuizRequest) (status.Code, error)
	ForceDeleteQuiz(ctx context.Context, req *dto.ForceDeleteUserLatestQuizRequest) (status.Code, error)
}
