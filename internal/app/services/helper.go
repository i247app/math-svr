package services

import (
	"context"

	chatbox "math-ai.com/math-ai/internal/driven-adapter/external/chat-box"
	"math-ai.com/math-ai/internal/shared/config"
	"math-ai.com/math-ai/internal/shared/logger"
)

func DetermineAIProvider(ctx context.Context, env config.Env) chatbox.IChatBoxClient {
	var chatBoxClient chatbox.IChatBoxClient

	provider := env.ChatBoxProvider
	if env.ChatBoxTestMode {
		provider = "mock"
	}

	switch provider {
	case "google", "gemini":
		logger.Info("ChatBox using GOOGLE GEMINI provider (free tier available)")
		googleClient, googleErr := chatbox.NewGoogleGeminiClient(context.Background(), env.ChatBoxAPIKey)
		if googleErr != nil {
			logger.Errorf("Failed to initialize Google Gemini client: %v", googleErr)
			logger.Warn("Falling back to MOCK mode")
			chatBoxClient = chatbox.NewMockOpenAIClient()
		} else {
			chatBoxClient = googleClient
		}

	case "openai":
		logger.Info("ChatBox using OPENAI provider")
		chatBoxClient = chatbox.NewOpenAIClient(env.ChatBoxAPIKey)

	case "langchain":
		logger.Info("ChatBox using LANGCHAIN provider")
		langchainConfig := chatbox.LangChainConfig{
			Provider:  env.ChatBoxLangChainProvider,
			APIKey:    env.ChatBoxAPIKey,
			ModelName: env.ChatBoxModelName,
		}
		langchainClient, langchainErr := chatbox.NewLangChainClient(context.Background(), langchainConfig)
		if langchainErr != nil {
			logger.Errorf("Failed to initialize LangChain client: %v", langchainErr)
			logger.Warn("Falling back to MOCK mode")
			chatBoxClient = chatbox.NewMockOpenAIClient()
		} else {
			logger.Infof("LangChain initialized with sub-provider: %s", env.ChatBoxLangChainProvider)
			chatBoxClient = langchainClient
		}

	case "mock", "test":
		logger.Info("ChatBox using MOCK provider (test mode - no API calls)")
		chatBoxClient = chatbox.NewMockOpenAIClient()

	default:
		logger.Warnf("Unknown ChatBox provider '%s', defaulting to MOCK mode", provider)
		chatBoxClient = chatbox.NewMockOpenAIClient()
	}

	return chatBoxClient
}
