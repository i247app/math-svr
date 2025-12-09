package di

import (
	"context"

	"math-ai.com/math-ai/internal/applications/dto"
	"math-ai.com/math-ai/internal/shared/constant/status"
)

type IUserQuizAssessmentService interface {
	GenerateQuizAssessment(ctx context.Context, req *dto.GenerateQuizAssessmentRequest) (status.Code, *dto.ChatBoxResponse[[]dto.Question], error)
	SubmitQuizAssessment(ctx context.Context, req *dto.SubmitQuizAssessmentRequest) (status.Code, *dto.ChatBoxResponse[dto.QuizAssessmentAnswer], error)
	ReinforceQuizAssessment(ctx context.Context, req *dto.ReinforceQuizAssessmentRequest) (status.Code, *dto.ChatBoxResponse[[]dto.Question], error)
	SubmitReinforceQuizAssessment(ctx context.Context, req *dto.SubmitReinforceQuizAssessmentRequest) (status.Code, *dto.ChatBoxResponse[dto.QuizAssessmentAnswer], error)
	GetUserQuizAssessmentsHistory(ctx context.Context, req *dto.GetUserQuizAssessmentsHistoryRequest) (status.Code, *dto.UserQuizAssessmentsHistoryResponse, error)

	GetUserQuizAssessmentByID(ctx context.Context, id string) (status.Code, *dto.UserQuizAssessmentResponse, error)
	CreateUserQuizAssessment(ctx context.Context, req *dto.CreateUserQuizAssessmentRequest) (status.Code, *dto.UserQuizAssessmentResponse, error)
	UpdateUserQuizAssessment(ctx context.Context, req *dto.UpdateUserQuizAssessmentRequest) (status.Code, *dto.UserQuizAssessmentResponse, error)
	DeleteUserQuizAssessment(ctx context.Context, req *dto.DeleteUserQuizAssessmentRequest) (status.Code, error)
	ForceDeleteUserQuizAssessment(ctx context.Context, req *dto.DeleteUserQuizAssessmentRequest) (status.Code, error)
}
