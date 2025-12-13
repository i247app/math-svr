package services

import (
	"context"
	"log"

	aiProvider "math-ai.com/math-ai/internal/driven-adapter/external/ai_provider"
	"math-ai.com/math-ai/internal/shared/config"
)

func DetermineAIProvider(ctx context.Context, env config.Env) aiProvider.IChatBoxClient {
	var chatBoxClient aiProvider.IChatBoxClient

	provider := env.ChatBoxProvider
	if env.ChatBoxTestMode {
		provider = "mock"
	}

	switch provider {
	case "google", "gemini":
		log.Println("ChatBox using GOOGLE GEMINI provider (free tier available)")
		googleClient, googleErr := aiProvider.NewGoogleGeminiClient(context.Background(), env.ChatBoxAPIKey)
		if googleErr != nil {
			chatBoxClient = aiProvider.NewMockOpenAIClient()
		} else {
			chatBoxClient = googleClient
		}

	case "openai":
		log.Println("ChatBox using OPENAI provider")
		chatBoxClient = aiProvider.NewOpenAIClient(env.ChatBoxAPIKey)

	case "langchain":
		log.Println("ChatBox using LANGCHAIN provider")
		langchainConfig := aiProvider.LangChainConfig{
			Provider:  env.ChatBoxLangChainProvider,
			APIKey:    env.ChatBoxAPIKey,
			ModelName: env.ChatBoxModelName,
		}
		langchainClient, langchainErr := aiProvider.NewLangChainClient(context.Background(), langchainConfig)
		if langchainErr != nil {
			chatBoxClient = aiProvider.NewMockOpenAIClient()
		} else {
			log.Printf("LangChain initialized with sub-provider: %s", env.ChatBoxLangChainProvider)
			chatBoxClient = langchainClient
		}

	case "mock", "test":
		log.Println("ChatBox using MOCK provider (test mode - no API calls)")
		chatBoxClient = aiProvider.NewMockOpenAIClient()

	default:
		log.Printf("Unknown ChatBox provider '%s', defaulting to MOCK mode", provider)
		chatBoxClient = aiProvider.NewMockOpenAIClient()
	}

	return chatBoxClient
}
