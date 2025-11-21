package chatbox

import (
	"context"
	"fmt"
	"strings"
	"time"

	domain "math-ai.com/math-ai/internal/core/domain/chatbox"
	"math-ai.com/math-ai/internal/shared/logger"
)

// MockOpenAIClient is a mock implementation for testing without OpenAI API
type MockOpenAIClient struct {
	enabled bool
}

// NewMockOpenAIClient creates a new mock OpenAI client for testing
func NewMockOpenAIClient() *MockOpenAIClient {
	logger.Info("Initializing Mock OpenAI client (TEST MODE - No API calls will be made)")
	return &MockOpenAIClient{
		enabled: true,
	}
}

// SendMessage sends a mock message and returns a simulated response
func (m *MockOpenAIClient) SendMessage(ctx context.Context, conv *domain.Conversation) (*ChatCompletionResponse, error) {
	if !m.enabled {
		return nil, fmt.Errorf("mock client is disabled")
	}

	logger.Info("[MOCK MODE] Processing message (no actual API call)")

	// Get the last user message
	messages := conv.Messages()
	var userMessage string
	if len(messages) > 0 {
		userMessage = messages[len(messages)-1].Content()
	}

	// Generate mock response based on user input
	mockResponse := m.generateMockResponse(userMessage)

	// Simulate API delay
	time.Sleep(500 * time.Millisecond)

	return &ChatCompletionResponse{
		Message:          mockResponse,
		Role:             "assistant",
		Model:            conv.Model(),
		FinishReason:     "stop",
		PromptTokens:     len(userMessage) / 4,     // Rough estimate
		CompletionTokens: len(mockResponse) / 4,    // Rough estimate
		TotalTokens:      (len(userMessage) + len(mockResponse)) / 4,
	}, nil
}

// StreamMessage sends a mock streaming message
func (m *MockOpenAIClient) StreamMessage(ctx context.Context, conv *domain.Conversation) (<-chan StreamChunk, error) {
	if !m.enabled {
		return nil, fmt.Errorf("mock client is disabled")
	}

	logger.Info("[MOCK MODE] Processing streaming message (no actual API call)")

	// Get the last user message
	messages := conv.Messages()
	var userMessage string
	if len(messages) > 0 {
		userMessage = messages[len(messages)-1].Content()
	}

	// Generate mock response
	mockResponse := m.generateMockResponse(userMessage)

	// Create channel for streaming chunks
	chunkChan := make(chan StreamChunk)

	// Start goroutine to simulate streaming
	go func() {
		defer close(chunkChan)

		// Split response into words for streaming simulation
		words := strings.Fields(mockResponse)

		for i, word := range words {
			// Add space before word (except first word)
			if i > 0 {
				word = " " + word
			}

			// Send chunk
			chunkChan <- StreamChunk{
				Delta:        word,
				FinishReason: "",
				Done:         false,
			}

			// Simulate streaming delay
			time.Sleep(100 * time.Millisecond)
		}

		// Send final chunk
		chunkChan <- StreamChunk{
			Delta:        "",
			FinishReason: "stop",
			Done:         true,
		}
	}()

	return chunkChan, nil
}

// generateMockResponse generates a contextual mock response based on user input
func (m *MockOpenAIClient) generateMockResponse(userMessage string) string {
	lowerMsg := strings.ToLower(userMessage)

	// Context-aware responses
	switch {
	case contains(lowerMsg, "hello") || contains(lowerMsg, "hi"):
		return "Hello! I'm a mock AI assistant running in test mode. How can I help you today?"

	case contains(lowerMsg, "how are you"):
		return "I'm doing great! I'm a mock AI running in test mode, so I don't have real feelings, but I'm here to help you test your chatbox feature."

	case contains(lowerMsg, "what") && contains(lowerMsg, "name"):
		return "I'm Mock AI, a test assistant. I'm here to help you test your chatbox feature without using real OpenAI API credits."

	case contains(lowerMsg, "thank"):
		return "You're welcome! Remember, I'm running in test mode, so these are simulated responses."

	case contains(lowerMsg, "test") || contains(lowerMsg, "mock"):
		return "Yes, I'm running in MOCK/TEST mode. This means no actual API calls are being made to OpenAI, and you're not being charged. Perfect for development and testing!"

	case contains(lowerMsg, "math") || contains(lowerMsg, "calculate"):
		return "In test mode, I can simulate math responses. For example: 2 + 2 = 4. For real calculations, you'll need to use the actual OpenAI API."

	case contains(lowerMsg, "code") || contains(lowerMsg, "program"):
		return "Mock response: Here's a simple example:\n\nfunc main() {\n    fmt.Println(\"Hello, World!\")\n}\n\nThis is a simulated code response for testing purposes."

	case strings.Contains(lowerMsg, "?"):
		return fmt.Sprintf("That's an interesting question about '%s'. In test mode, I provide simulated responses. For real AI capabilities, configure a valid OpenAI API key.", userMessage)

	default:
		return fmt.Sprintf("This is a mock response to your message: '%s'. I'm running in test mode to help you develop without API costs. When you're ready for production, configure a real OpenAI API key.", userMessage)
	}
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
