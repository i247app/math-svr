package di

import (
	"context"

	"math-ai.com/math-ai/internal/applications/dto"
	"math-ai.com/math-ai/internal/shared/constant/status"
)

type IChatBoxService interface {
	// GenerateQuiz handles quiz generation requests
	GenerateQuiz(ctx context.Context, req *dto.GenerateQuizRequest) (status.Code, *dto.ChatBoxResponse[[]dto.Question], error)

	// SubmitQuiz handles quiz submission requests
	SubmitQuiz(ctx context.Context, req *dto.SubmitQuizRequest) (status.Code, *dto.ChatBoxResponse[dto.QuizAnswer], error)

	SendMessageStream(ctx context.Context, req *dto.GenerateQuizRequest) (status.Code, <-chan dto.ChatBoxStreamChunk, error)
}
