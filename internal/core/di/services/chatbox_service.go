package di

import (
	"context"

	"math-ai.com/math-ai/internal/applications/dto"
	"math-ai.com/math-ai/internal/shared/constant/status"
)

type IChatBoxService interface {
	// SendMessage sends a message to the chatbox and gets a response
	SendMessage(ctx context.Context, req *dto.ChatBoxRequest) (status.Code, *dto.ChatBoxResponse, error)

	// SendMessageStream sends a message and streams the response
	SendMessageStream(ctx context.Context, req *dto.ChatBoxRequest) (status.Code, <-chan dto.ChatBoxStreamChunk, error)
}
