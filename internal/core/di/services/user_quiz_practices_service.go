package di

import (
	"context"

	"math-ai.com/math-ai/internal/applications/dto"
	"math-ai.com/math-ai/internal/shared/constant/status"
)

type IUserQuizPracticesService interface {
	GenerateQuiz(ctx context.Context, req *dto.GenerateQuizRequest) (status.Code, *dto.ChatBoxResponse[[]dto.Question], error)
	SubmitQuiz(ctx context.Context, req *dto.SubmitQuizRequest) (status.Code, *dto.ChatBoxResponse[dto.QuizAnswer], error)
	GenerateQuizPractice(ctx context.Context, req *dto.GenerateQuizPracticeRequest) (status.Code, *dto.ChatBoxResponse[[]dto.Question], error)

	GetUserQuizPratice(ctx context.Context, req *dto.GetUserQuizPracticesRequest) (status.Code, *dto.UserQuizPracticesResponse, error)
	GetUserQuizPraticeByUID(ctx context.Context, req *dto.GetUserQuizPracticesByUIDRequest) (status.Code, *dto.UserQuizPracticesResponse, error)
	CreateUserQuizPratice(ctx context.Context, req *dto.CreateUserQuizPracticesRequest) (status.Code, *dto.UserQuizPracticesResponse, error)
	UpdateUserQuizPratice(ctx context.Context, req *dto.UpdateUserQuizPracticesRequest) (status.Code, *dto.UserQuizPracticesResponse, error)
	DeleteUserQuizPratice(ctx context.Context, req *dto.DeleteUserQuizPracticesRequest) (status.Code, error)
	ForceDeleteUserQuizPratice(ctx context.Context, req *dto.DeleteUserQuizPracticesRequest) (status.Code, error)
}
