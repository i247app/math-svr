package di

import (
	"context"

	"math-ai.com/math-ai/internal/applications/dto"
	"math-ai.com/math-ai/internal/shared/constant/status"
)

type IChatBoxService interface {
	GenerateQuiz(ctx context.Context, req *dto.GenerateQuizRequest) (status.Code, *dto.ChatBoxResponse[[]dto.Question], error)
	SubmitQuiz(ctx context.Context, req *dto.SubmitQuizRequest) (status.Code, *dto.ChatBoxResponse[dto.QuizAnswer], error)
	GenerateQuizPractice(ctx context.Context, req *dto.GenerateQuizPracticeRequest) (status.Code, *dto.ChatBoxResponse[[]dto.Question], error)

	SendMessageStream(ctx context.Context, req *dto.GenerateQuizRequest) (status.Code, <-chan dto.ChatBoxStreamChunk, error)
}
