package chatbox

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"

	domain "math-ai.com/math-ai/internal/core/domain/chatbox"
)

// GoogleGeminiClient wraps the Google Gemini API client
type GoogleGeminiClient struct {
	client *genai.Client
	apiKey string
	model  string
}

// NewGoogleGeminiClient creates a new Google Gemini client
func NewGoogleGeminiClient(ctx context.Context, apiKey string) (*GoogleGeminiClient, error) {
	if apiKey == "" {
		//logger.Warn("Google Gemini API key is empty. ChatBox functionality will not work.")
		return nil, errors.New("Google Gemini API key is required")
	}

	////logger.Infof("Initializing Google Gemini client with model: gemini-2.5-flash")
	////logger.Infof("API Key: %s", apiKey)

	// Create Gemini client
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		////logger.Errorf("Failed to create Google Gemini client: %v", err)
		return nil, fmt.Errorf("failed to create Google Gemini client: %w", err)
	}

	return &GoogleGeminiClient{
		client: client,
		apiKey: apiKey,
		model:  "gemini-2.5-flash", // Free tier model - stable and widely available
	}, nil
}

// Close closes the Google Gemini client
func (c *GoogleGeminiClient) Close() error {
	return c.client.Close()
}

// SendMessage sends a conversation to Google Gemini and returns the response
func (c *GoogleGeminiClient) SendMessage(ctx context.Context, conv *domain.Conversation) (*ChatCompletionResponse, error) {
	if c.apiKey == "" {
		return nil, errors.New("Google Gemini API key is not configured")
	}

	// Get the model
	model := c.client.GenerativeModel(c.model)

	// Configure generation parameters
	model.Temperature = floatPtr(conv.Temperature())
	model.MaxOutputTokens = int32Ptr(int32(conv.MaxTokens()))

	// Set system instruction if provided
	if conv.SystemPrompt() != nil && *conv.SystemPrompt() != "" {
		model.SystemInstruction = &genai.Content{
			Parts: []genai.Part{genai.Text(*conv.SystemPrompt())},
		}
	}

	// Start chat session
	chat := model.StartChat()

	// Build chat history (exclude the last message as it will be sent separately)
	messages := conv.Messages()
	if len(messages) > 1 {
		for i := 0; i < len(messages)-1; i++ {
			msg := messages[i]
			role := "user"
			if msg.Role() == "assistant" {
				role = "model"
			}

			chat.History = append(chat.History, &genai.Content{
				Parts: []genai.Part{genai.Text(msg.Content())},
				Role:  role,
			})
		}
	}

	// Send the last message
	var prompt string
	if len(messages) > 0 {
		prompt = messages[len(messages)-1].Content()
	}

	// Generate response
	////logger.Infof("Sending message to Google Gemini: %s", prompt)
	resp, err := chat.SendMessage(ctx, genai.Text(prompt))
	if err != nil {
		////logger.Errorf("Failed to send message to Google Gemini: %v", err)
		return nil, fmt.Errorf("failed to send message to Google Gemini: %w", err)
	}

	// Extract response text
	if len(resp.Candidates) == 0 {
		return nil, errors.New("no response from Google Gemini")
	}

	candidate := resp.Candidates[0]
	if candidate.Content == nil || len(candidate.Content.Parts) == 0 {
		return nil, errors.New("empty response from Google Gemini")
	}

	// Extract text from parts
	var responseText string
	for _, part := range candidate.Content.Parts {
		if txt, ok := part.(genai.Text); ok {
			responseText += string(txt)
		}
	}

	// Get token counts
	var promptTokens, completionTokens int
	if resp.UsageMetadata != nil {
		promptTokens = int(resp.UsageMetadata.PromptTokenCount)
		completionTokens = int(resp.UsageMetadata.CandidatesTokenCount)
	}

	response := &ChatCompletionResponse{
		Message:          responseText,
		Role:             "assistant",
		Model:            c.model,
		FinishReason:     string(candidate.FinishReason),
		PromptTokens:     promptTokens,
		CompletionTokens: completionTokens,
		TotalTokens:      promptTokens + completionTokens,
	}

	return response, nil
}

// StreamMessage sends a conversation to Google Gemini and streams the response
func (c *GoogleGeminiClient) StreamMessage(ctx context.Context, conv *domain.Conversation) (<-chan StreamChunk, error) {
	if c.apiKey == "" {
		return nil, errors.New("Google Gemini API key is not configured")
	}

	// Get the model
	model := c.client.GenerativeModel(c.model)

	// Configure generation parameters
	model.Temperature = floatPtr(conv.Temperature())
	model.MaxOutputTokens = int32Ptr(int32(conv.MaxTokens()))

	// Set system instruction if provided
	if conv.SystemPrompt() != nil && *conv.SystemPrompt() != "" {
		model.SystemInstruction = &genai.Content{
			Parts: []genai.Part{genai.Text(*conv.SystemPrompt())},
		}
	}

	// Start chat session
	chat := model.StartChat()

	// Build chat history (exclude the last message)
	messages := conv.Messages()
	if len(messages) > 1 {
		for i := 0; i < len(messages)-1; i++ {
			msg := messages[i]
			role := "user"
			if msg.Role() == "assistant" {
				role = "model"
			}

			chat.History = append(chat.History, &genai.Content{
				Parts: []genai.Part{genai.Text(msg.Content())},
				Role:  role,
			})
		}
	}

	// Send the last message
	var prompt string
	if len(messages) > 0 {
		prompt = messages[len(messages)-1].Content()
	}

	// Create streaming iterator
	iter := chat.SendMessageStream(ctx, genai.Text(prompt))

	// Create channel for streaming chunks
	chunkChan := make(chan StreamChunk)

	// Start goroutine to read stream
	go func() {
		defer close(chunkChan)

		for {
			resp, err := iter.Next()
			if errors.Is(err, iterator.Done) {
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

			// Extract text from response
			if len(resp.Candidates) > 0 {
				candidate := resp.Candidates[0]
				if candidate.Content != nil {
					for _, part := range candidate.Content.Parts {
						if txt, ok := part.(genai.Text); ok {
							chunkChan <- StreamChunk{
								Delta:        string(txt),
								FinishReason: string(candidate.FinishReason),
								Done:         false,
							}
						}
					}
				}
			}
		}
	}()

	return chunkChan, nil
}

// Helper functions to create pointers
func floatPtr(f float32) *float32 {
	return &f
}

func int32Ptr(i int32) *int32 {
	return &i
}
