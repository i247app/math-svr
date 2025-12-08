package di

import (
	"context"

	"math-ai.com/math-ai/internal/applications/dto"
	domain "math-ai.com/math-ai/internal/core/domain/chatbox"
	"math-ai.com/math-ai/internal/shared/constant/status"
)

type IChatBoxService interface {
	Generate(ctx context.Context, conv *domain.Conversation) (status.Code, *dto.ChatBoxResponse[[]dto.Question], error)
	Submit(ctx context.Context, conv *domain.Conversation) (status.Code, *dto.ChatBoxResponse[dto.QuizAnswer], error)
	GeneratePractice(ctx context.Context, conv *domain.Conversation) (status.Code, *dto.ChatBoxResponse[[]dto.Question], error)
}
