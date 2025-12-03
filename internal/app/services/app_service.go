package services

import (
	"context"
	"fmt"
	"time"

	"github.com/i247app/gex/jwtutil"
	gexsess "github.com/i247app/gex/session"
	"github.com/i247app/gex/sessionprovider"
	"math-ai.com/math-ai/internal/app/resources"
	"math-ai.com/math-ai/internal/applications/services"
	"math-ai.com/math-ai/internal/applications/validators"
	di "math-ai.com/math-ai/internal/core/di/services"
	"math-ai.com/math-ai/internal/driven-adapter/persistence/repositories"
	"math-ai.com/math-ai/internal/session"
	"math-ai.com/math-ai/internal/shared/logger"
)

type ServiceContainer struct {
	SessionManager  *session.SessionManager
	SessionProvider sessionprovider.SessionProvider
	JwtHelper       jwtutil.JwtHelper

	LoginService          di.ILoginService
	UserService           di.IUserService
	DeviceService         di.IDeviceService
	ChatBoxService        di.IChatBoxService
	GradeService          di.IGradeService
	LevelService          di.ILevelService
	ProfileService        di.IProfileService
	UserLatestQuizService di.IUserLatestQuizService
	StorageService        di.IStorageService
}

const (
	sessionTTL = 14 * 24 * time.Hour // 14 days
)

func SetupServiceContainer(res *resources.AppResource) (*ServiceContainer, error) {
	env := res.Env

	logger.Info("Initializing repository")
	loginRepo := repositories.NewloginRepository(res.Db)
	userRepo := repositories.NewUserRepository(res.Db)
	deviceRepo := repositories.NewDeviceRepository(res.Db)
	gradeRepo := repositories.NewGradeRepository(res.Db)
	levelRepo := repositories.NewLevelRepository(res.Db)
	profileRepo := repositories.NewProfileRepository(res.Db)
	userLatestQuizRepo := repositories.NewUserLatestQuizRepository(res.Db)

	logger.Info("Initializing services")

	logger.Info("> sessionManager...")
	sessionManager := session.NewSessionManager()

	logger.Info("> jwtHelper...")
	var jwtHelper jwtutil.JwtHelper
	if env.SharedKeyBytes != nil {
		helper, err := jwtutil.NewHmacJwtHelper(env.SharedKeyBytes)
		if err != nil {
			panic("failed to create jwt toolkit from env shared key")
		}
		jwtHelper = helper
	} else {
		return nil, fmt.Errorf("unable to determine jwt helper from env")
	}

	// Build the session provider
	logger.Info("> sessionProvider...")
	var sessionProvider sessionprovider.SessionProvider
	{
		defaultSessFactory := func() gexsess.SessionStorer {
			// Create the basic session that all new sessions are based on
			return session.NewSession()
		}
		if env.GexSessionDriver == "xwt" {
			sessionProvider = sessionprovider.NewXwtSessionProvider(
				sessionManager.Container(),
				jwtHelper,
				defaultSessFactory,
				sessionTTL,
			)
		} else {
			sessionProvider = sessionprovider.NewJwtSessionProvider(
				sessionManager.Container(),
				jwtHelper,
				defaultSessFactory,
				sessionTTL,
			)
		}
	}

	logger.Info("> loginSvc...")
	var userValidator = validators.NewUserValidator()
	var userSvc = services.NewUserService(userValidator, userRepo, loginRepo, profileRepo)

	logger.Info("> loginSvc...")
	var loginValidator = validators.NewLoginValidator()
	var loginSvc = services.NewLoginService(loginValidator, loginRepo, userRepo)

	logger.Info("> deviceSvc...")
	var deviceValidator = validators.NewDeviceValidator()
	var deviceSvc = services.NewDeviceService(deviceValidator, deviceRepo)

	logger.Info("> gradeSvc...")
	var gradeValidator = validators.NewGradeValidator()
	var gradeSvc = services.NewGradeService(gradeValidator, gradeRepo)

	logger.Info("> levelSvc...")
	var levelValidator = validators.NewLevelValidator()
	var levelSvc = services.NewLevelService(levelValidator, levelRepo)

	logger.Info("> profileSvc...")
	var profileValidator = validators.NewProfileValidator()
	var profileSvc = services.NewProfileService(profileValidator, profileRepo)

	logger.Info("> userLatestQuizSvc...")
	var userLatestQuizValidator = validators.NewUserLatestQuizValidator()
	var userLatestQuizSvc = services.NewUserLatestQuizService(userLatestQuizValidator, userLatestQuizRepo)

	logger.Info("> chatBoxSvc...")
	chatBoxClient := DetermineAIProvider(context.Background(), *res.Env)

	var chatBoxValidator = validators.NewChatboxValidator()
	var chatBoxSvc = services.NewChatBoxService(chatBoxClient, chatBoxValidator, profileSvc, userLatestQuizSvc)

	logger.Info("> storageSvc...")
	var storageSvc = services.NewStorageService(res.Env.S3Config)

	return &ServiceContainer{
		SessionManager:        sessionManager,
		SessionProvider:       sessionProvider,
		JwtHelper:             jwtHelper,
		LoginService:          loginSvc,
		UserService:           userSvc,
		DeviceService:         deviceSvc,
		ChatBoxService:        chatBoxSvc,
		GradeService:          gradeSvc,
		LevelService:          levelSvc,
		ProfileService:        profileSvc,
		UserLatestQuizService: userLatestQuizSvc,
		StorageService:        storageSvc,
	}, nil
}
