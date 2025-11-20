package services

import (
	"math-ai.com/math-ai/internal/app/resources"
	"math-ai.com/math-ai/internal/applications/services"
	di "math-ai.com/math-ai/internal/core/di/services"
	"math-ai.com/math-ai/internal/driven-adapter/persistence/repositories"
	"math-ai.com/math-ai/internal/shared/logger"
)

type ServiceContainer struct {
	LoginService di.ILoginService
	UserService  di.IUserService
}

func SetupServiceContainer(res *resources.AppResource) (*ServiceContainer, error) {
	logger.Info("Initializing repository")
	loginRepo := repositories.NewloginRepository(res.Db)
	userRepo := repositories.NewUserRepository(res.Db)

	logger.Info("Initializing services")
	logger.Info("> loginSvc...")
	var userSvc = services.NewUserService(userRepo, loginRepo)

	logger.Info("> loginSvc...")
	var loginSvc = services.NewLoginService(loginRepo, userRepo)

	return &ServiceContainer{
		LoginService: loginSvc,
		UserService:  userSvc,
	}, nil
}
