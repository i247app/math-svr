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
}
