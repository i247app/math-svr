package ai_provider

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/sashabaranov/go-openai"
	domain "math-ai.com/math-ai/internal/core/domain/chatbox"
)

// OpenAIClient wraps the OpenAI API client
type OpenAIClient struct {
	client *openai.Client
	apiKey string
}

// NewOpenAIClient creates a new OpenAI client
func NewOpenAIClient(apiKey string) *OpenAIClient {
	if apiKey == "" {
		//logger.Warn("OpenAI API key is empty. ChatBox functionality will not work.")
	}

	////logger.Infof("Initializing OpenAI client with API key: %s", apiKey)
	client := openai.NewClient(apiKey)

	return &OpenAIClient{
		client: client,
		apiKey: apiKey,
	}
}

// ChatCompletionResponse represents the response from OpenAI
type ChatCompletionResponse struct {
	Message          string
	Role             string
	Model            string
	FinishReason     string
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
}

// SendMessage sends a conversation to OpenAI and returns the response
func (c *OpenAIClient) SendMessage(ctx context.Context, conv *domain.Conversation) (*ChatCompletionResponse, error) {
	if c.apiKey == "" {
		return nil, errors.New("OpenAI API key is not configured")
	}

	// Build messages for OpenAI API
	messages := c.buildOpenAIMessages(conv)

	// Create chat completion request
	req := openai.ChatCompletionRequest{
		Model:       conv.Model(),
		Messages:    messages,
		Temperature: conv.Temperature(),
		MaxTokens:   conv.MaxTokens(),
	}

	// Send request to OpenAI
	resp, err := c.client.CreateChatCompletion(ctx, req)
	if err != nil {
		////logger.Errorf("Failed to create chat completion: %v", err)
		return nil, fmt.Errorf("failed to create chat completion: %w", err)
	}

	// Check if we got a response
	if len(resp.Choices) == 0 {
		return nil, errors.New("no response from OpenAI")
	}

	// Extract the response
	choice := resp.Choices[0]
	response := &ChatCompletionResponse{
		Message:          choice.Message.Content,
		Role:             choice.Message.Role,
		Model:            resp.Model,
		FinishReason:     string(choice.FinishReason),
		PromptTokens:     resp.Usage.PromptTokens,
		CompletionTokens: resp.Usage.CompletionTokens,
		TotalTokens:      resp.Usage.TotalTokens,
	}

	return response, nil
}

// StreamMessage sends a conversation to OpenAI and streams the response
func (c *OpenAIClient) StreamMessage(ctx context.Context, conv *domain.Conversation) (<-chan StreamChunk, error) {
	if c.apiKey == "" {
		return nil, errors.New("OpenAI API key is not configured")
	}

	// Build messages for OpenAI API
	messages := c.buildOpenAIMessages(conv)

	// Create chat completion stream request
	req := openai.ChatCompletionRequest{
		Model:       conv.Model(),
		Messages:    messages,
		Temperature: conv.Temperature(),
		MaxTokens:   conv.MaxTokens(),
		Stream:      true,
	}

	// Send request to OpenAI
	stream, err := c.client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		////logger.Errorf("Failed to create chat completion stream: %v", err)
		return nil, fmt.Errorf("failed to create chat completion stream: %w", err)
	}

	// Create channel for streaming chunks
	chunkChan := make(chan StreamChunk)

	// Start goroutine to read stream
	go func() {
		defer close(chunkChan)
		defer stream.Close()

		for {
			response, err := stream.Recv()
			if errors.Is(err, io.EOF) {
				// Stream finished successfully
				chunkChan <- StreamChunk{
					Delta:        "",
					FinishReason: "stop",
					Done:         true,
				}
				return
			}

			if err != nil {
				////logger.Errorf("Stream error: %v", err)
				chunkChan <- StreamChunk{
					Error: err,
					Done:  true,
				}
				return
			}

			// Send chunk to channel
			if len(response.Choices) > 0 {
				choice := response.Choices[0]
				chunkChan <- StreamChunk{
					Delta:        choice.Delta.Content,
					FinishReason: string(choice.FinishReason),
					Done:         false,
				}
			}
		}
	}()

	return chunkChan, nil
}

// StreamChunk represents a chunk of streaming data
type StreamChunk struct {
	Delta        string
	FinishReason string
	Done         bool
	Error        error
}

// buildOpenAIMessages converts domain messages to OpenAI message format
func (c *OpenAIClient) buildOpenAIMessages(conv *domain.Conversation) []openai.ChatCompletionMessage {
	messages := make([]openai.ChatCompletionMessage, 0)

	// Add system prompt if provided
	if conv.SystemPrompt() != nil && *conv.SystemPrompt() != "" {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: *conv.SystemPrompt(),
		})
	}

	// Add conversation messages
	for _, msg := range conv.Messages() {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    msg.Role(),
			Content: msg.Content(),
		})
	}

	return messages
}
