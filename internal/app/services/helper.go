package services

import (
	"context"
	"log"

	chatbox "math-ai.com/math-ai/internal/driven-adapter/external/chat-box"
	"math-ai.com/math-ai/internal/shared/config"
)

func DetermineAIProvider(ctx context.Context, env config.Env) chatbox.IChatBoxClient {
	var chatBoxClient chatbox.IChatBoxClient

	provider := env.ChatBoxProvider
	if env.ChatBoxTestMode {
		provider = "mock"
	}

	switch provider {
	case "google", "gemini":
		log.Println("ChatBox using GOOGLE GEMINI provider (free tier available)")
		googleClient, googleErr := chatbox.NewGoogleGeminiClient(context.Background(), env.ChatBoxAPIKey)
		if googleErr != nil {
			chatBoxClient = chatbox.NewMockOpenAIClient()
		} else {
			chatBoxClient = googleClient
		}

	case "openai":
		log.Println("ChatBox using OPENAI provider")
		chatBoxClient = chatbox.NewOpenAIClient(env.ChatBoxAPIKey)

	case "langchain":
		log.Println("ChatBox using LANGCHAIN provider")
		langchainConfig := chatbox.LangChainConfig{
			Provider:  env.ChatBoxLangChainProvider,
			APIKey:    env.ChatBoxAPIKey,
			ModelName: env.ChatBoxModelName,
		}
		langchainClient, langchainErr := chatbox.NewLangChainClient(context.Background(), langchainConfig)
		if langchainErr != nil {
			chatBoxClient = chatbox.NewMockOpenAIClient()
		} else {
			log.Printf("LangChain initialized with sub-provider: %s", env.ChatBoxLangChainProvider)
			chatBoxClient = langchainClient
		}

	case "mock", "test":
		log.Println("ChatBox using MOCK provider (test mode - no API calls)")
		chatBoxClient = chatbox.NewMockOpenAIClient()

	default:
		log.Printf("Unknown ChatBox provider '%s', defaulting to MOCK mode", provider)
		chatBoxClient = chatbox.NewMockOpenAIClient()
	}

	return chatBoxClient
}
