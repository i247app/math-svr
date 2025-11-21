package services

import (
	"context"

	"math-ai.com/math-ai/internal/app/resources"
	"math-ai.com/math-ai/internal/applications/services"
	di "math-ai.com/math-ai/internal/core/di/services"
	chatbox "math-ai.com/math-ai/internal/driven-adapter/external/chat-box"
	"math-ai.com/math-ai/internal/driven-adapter/persistence/repositories"
	"math-ai.com/math-ai/internal/shared/logger"
)

type ServiceContainer struct {
	LoginService   di.ILoginService
	UserService    di.IUserService
	DeviceService  di.IDeviceService
	ChatBoxService di.IChatBoxService
	GradeService   di.IGradeService
	LevelService   di.ILevelService
	ProfileService di.IProfileService
}

func SetupServiceContainer(res *resources.AppResource) (*ServiceContainer, error) {
	logger.Info("Initializing repository")
	loginRepo := repositories.NewloginRepository(res.Db)
	userRepo := repositories.NewUserRepository(res.Db)
	deviceRepo := repositories.NewDeviceRepository(res.Db)
	gradeRepo := repositories.NewGradeRepository(res.Db)
	levelRepo := repositories.NewLevelRepository(res.Db)
	profileRepo := repositories.NewProfileRepository(res.Db)

	logger.Info("Initializing services")
	logger.Info("> loginSvc...")
	var userSvc = services.NewUserService(userRepo, loginRepo)

	logger.Info("> loginSvc...")
	var loginSvc = services.NewLoginService(loginRepo, userRepo)

	logger.Info("> deviceSvc...")
	var deviceSvc = services.NewDeviceService(deviceRepo)

	logger.Info("> gradeSvc...")
	var gradeSvc = services.NewGradeService(gradeRepo)

	logger.Info("> levelSvc...")
	var levelSvc = services.NewLevelService(levelRepo)

	logger.Info("> profileSvc...")
	var profileSvc = services.NewProfileService(profileRepo)

	logger.Info("> chatBoxSvc...")
	var chatBoxClient chatbox.IChatBoxClient

	// Determine which provider to use
	provider := res.Env.ChatBoxProvider
	if res.Env.ChatBoxTestMode {
		provider = "mock"
	}

	switch provider {
	case "google", "gemini":
		logger.Info("ChatBox using GOOGLE GEMINI provider (free tier available)")
		googleClient, googleErr := chatbox.NewGoogleGeminiClient(context.Background(), res.Env.ChatBoxAPIKey)
		if googleErr != nil {
			logger.Errorf("Failed to initialize Google Gemini client: %v", googleErr)
			logger.Warn("Falling back to MOCK mode")
			chatBoxClient = chatbox.NewMockOpenAIClient()
		} else {
			chatBoxClient = googleClient
		}

	case "openai":
		logger.Info("ChatBox using OPENAI provider")
		chatBoxClient = chatbox.NewOpenAIClient(res.Env.ChatBoxAPIKey)

	case "langchain":
		logger.Info("ChatBox using LANGCHAIN provider")
		langchainConfig := chatbox.LangChainConfig{
			Provider:  res.Env.ChatBoxLangChainProvider,
			APIKey:    res.Env.ChatBoxAPIKey,
			ModelName: res.Env.ChatBoxModelName,
		}
		langchainClient, langchainErr := chatbox.NewLangChainClient(context.Background(), langchainConfig)
		if langchainErr != nil {
			logger.Errorf("Failed to initialize LangChain client: %v", langchainErr)
			logger.Warn("Falling back to MOCK mode")
			chatBoxClient = chatbox.NewMockOpenAIClient()
		} else {
			logger.Infof("LangChain initialized with sub-provider: %s", res.Env.ChatBoxLangChainProvider)
			chatBoxClient = langchainClient
		}

	case "mock", "test":
		logger.Info("ChatBox using MOCK provider (test mode - no API calls)")
		chatBoxClient = chatbox.NewMockOpenAIClient()

	default:
		logger.Warnf("Unknown ChatBox provider '%s', defaulting to MOCK mode", provider)
		chatBoxClient = chatbox.NewMockOpenAIClient()
	}

	var chatBoxSvc = services.NewChatBoxService(chatBoxClient)

	return &ServiceContainer{
		LoginService:   loginSvc,
		UserService:    userSvc,
		DeviceService:  deviceSvc,
		ChatBoxService: chatBoxSvc,
		GradeService:   gradeSvc,
		LevelService:   levelSvc,
		ProfileService: profileSvc,
	}, nil
}
