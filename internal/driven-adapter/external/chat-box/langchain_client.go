package chatbox

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/googleai"
	"github.com/tmc/langchaingo/llms/openai"

	domain "math-ai.com/math-ai/internal/core/domain/chatbox"
	"math-ai.com/math-ai/internal/shared/logger"
)

// LangChainClient wraps LangChain Go with support for multiple LLM providers
type LangChainClient struct {
	llm       llms.Model
	provider  string
	modelName string
	apiKey    string
}

// LangChainConfig holds configuration for LangChain client
type LangChainConfig struct {
	Provider  string // "openai" or "google"
	APIKey    string
	ModelName string
}

// NewLangChainClient creates a new LangChain client
func NewLangChainClient(ctx context.Context, config LangChainConfig) (*LangChainClient, error) {
	if config.APIKey == "" {
		logger.Warn("LangChain API key is empty")
		return nil, errors.New("LangChain API key is required")
	}

	var llmModel llms.Model
	var err error
	var modelName string

	switch strings.ToLower(config.Provider) {
	case "openai":
		modelName = config.ModelName
		if modelName == "" {
			modelName = "gpt-3.5-turbo"
		}
		logger.Infof("Initializing LangChain with OpenAI provider, model: %s", modelName)
		llmModel, err = openai.New(
			openai.WithToken(config.APIKey),
			openai.WithModel(modelName),
		)
		if err != nil {
			logger.Errorf("Failed to create LangChain OpenAI client: %v", err)
			return nil, fmt.Errorf("failed to create LangChain OpenAI client: %w", err)
		}

	case "google", "googleai":
		modelName = config.ModelName
		if modelName == "" {
			modelName = "gemini-2.5-flash"
		}
		logger.Infof("Initializing LangChain with Google AI provider, model: %s", modelName)
		llmModel, err = googleai.New(
			ctx,
			googleai.WithAPIKey(config.APIKey),
			googleai.WithDefaultModel(modelName),
		)
		if err != nil {
			logger.Errorf("Failed to create LangChain Google AI client: %v", err)
			return nil, fmt.Errorf("failed to create LangChain Google AI client: %w", err)
		}

	default:
		return nil, fmt.Errorf("unsupported LangChain provider: %s (supported: openai, google)", config.Provider)
	}

	return &LangChainClient{
		llm:       llmModel,
		provider:  config.Provider,
		modelName: modelName,
		apiKey:    config.APIKey,
	}, nil
}

// SendMessage sends a conversation to LangChain and returns the response
func (c *LangChainClient) SendMessage(ctx context.Context, conv *domain.Conversation) (*ChatCompletionResponse, error) {
	if c.llm == nil {
		return nil, errors.New("LangChain LLM is not initialized")
	}

	// Build messages for LangChain
	messages := c.buildMessages(conv)

	// Create generation options
	opts := []llms.CallOption{
		llms.WithTemperature(float64(conv.Temperature())),
		llms.WithMaxTokens(conv.MaxTokens()),
	}

	// Generate content
	logger.Infof("[LangChain] Generating content with %d messages", len(messages))
	resp, err := c.llm.GenerateContent(ctx, messages, opts...)
	if err != nil {
		logger.Errorf("Failed to generate content with LangChain: %v", err)
		return nil, fmt.Errorf("failed to generate content: %w", err)
	}

	// Extract response
	if len(resp.Choices) == 0 {
		return nil, errors.New("no response from LangChain")
	}

	choice := resp.Choices[0]
	responseText := choice.Content

	logger.Infof("[LangChain] Received response: %s", responseText)

	// Calculate token counts (approximate)
	promptTokens := c.estimateTokens(c.buildPromptText(conv))
	completionTokens := c.estimateTokens(responseText)

	response := &ChatCompletionResponse{
		Message:          responseText,
		Role:             "assistant",
		Model:            c.modelName,
		FinishReason:     string(choice.StopReason),
		PromptTokens:     promptTokens,
		CompletionTokens: completionTokens,
		TotalTokens:      promptTokens + completionTokens,
	}

	return response, nil
}

// StreamMessage sends a conversation to LangChain and streams the response
func (c *LangChainClient) StreamMessage(ctx context.Context, conv *domain.Conversation) (<-chan StreamChunk, error) {
	if c.llm == nil {
		return nil, errors.New("LangChain LLM is not initialized")
	}

	// Build messages for LangChain
	messages := c.buildMessages(conv)

	// Create generation options
	opts := []llms.CallOption{
		llms.WithTemperature(float64(conv.Temperature())),
		llms.WithMaxTokens(conv.MaxTokens()),
		llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
			// This will be handled by the stream callback
			return nil
		}),
	}

	// Create channel for streaming chunks
	chunkChan := make(chan StreamChunk)

	// Start goroutine to handle streaming
	go func() {
		defer close(chunkChan)

		// Create streaming callback
		var responseBuilder strings.Builder
		streamOpts := append(opts, llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
			text := string(chunk)
			responseBuilder.WriteString(text)

			// Send chunk
			chunkChan <- StreamChunk{
				Delta:        text,
				FinishReason: "",
				Done:         false,
			}
			return nil
		}))

		// Generate content with streaming
		_, err := c.llm.GenerateContent(ctx, messages, streamOpts...)
		if err != nil {
			logger.Errorf("Stream error: %v", err)
			chunkChan <- StreamChunk{
				Error: err,
				Done:  true,
			}
			return
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

// buildMessages converts domain conversation to LangChain message format
func (c *LangChainClient) buildMessages(conv *domain.Conversation) []llms.MessageContent {
	messages := make([]llms.MessageContent, 0)

	// Add system prompt if provided
	if conv.SystemPrompt() != nil && *conv.SystemPrompt() != "" {
		messages = append(messages, llms.MessageContent{
			Role: llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{
				llms.TextPart(*conv.SystemPrompt()),
			},
		})
	}

	// Add conversation messages
	for _, msg := range conv.Messages() {
		var role llms.ChatMessageType
		switch msg.Role() {
		case "user":
			role = llms.ChatMessageTypeHuman
		case "assistant":
			role = llms.ChatMessageTypeAI
		default:
			role = llms.ChatMessageTypeGeneric
		}

		messages = append(messages, llms.MessageContent{
			Role: role,
			Parts: []llms.ContentPart{
				llms.TextPart(msg.Content()),
			},
		})
	}

	return messages
}

// buildPromptText builds a text representation of the conversation for token estimation
func (c *LangChainClient) buildPromptText(conv *domain.Conversation) string {
	var builder strings.Builder

	if conv.SystemPrompt() != nil && *conv.SystemPrompt() != "" {
		builder.WriteString(*conv.SystemPrompt())
		builder.WriteString("\n")
	}

	for _, msg := range conv.Messages() {
		builder.WriteString(msg.Content())
		builder.WriteString("\n")
	}

	return builder.String()
}

// estimateTokens provides a rough estimate of token count
func (c *LangChainClient) estimateTokens(text string) int {
	// Rough estimation: ~4 characters per token
	return len(text) / 4
}
