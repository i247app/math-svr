package chatbox

import (
	"context"

	domain "math-ai.com/math-ai/internal/core/domain/chatbox"
)

// IChatBoxClient defines the interface for chatbox clients (real or mock)
type IChatBoxClient interface {
	SendMessage(ctx context.Context, conv *domain.Conversation) (*ChatCompletionResponse, error)
	StreamMessage(ctx context.Context, conv *domain.Conversation) (<-chan StreamChunk, error)
}
