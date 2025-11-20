package services

import (
	"math-ai.com/math-ai/internal/app/resources"
	"math-ai.com/math-ai/internal/applications/services"
	di "math-ai.com/math-ai/internal/core/di/services"
	"math-ai.com/math-ai/internal/driven-adapter/persistence/repositories"
	"math-ai.com/math-ai/internal/shared/logger"
)

type ServiceContainer struct {
	UserService di.IUserService
}

func SetupServiceContainer(res *resources.AppResource) (*ServiceContainer, error) {
	logger.Info("Initializing services")

	logger.Info("> userSvc...")
	userRepo := repositories.NewUserRepository(res.Db)
	var userSvc = services.NewUserService(userRepo)

	return &ServiceContainer{
		UserService: userSvc,
	}, nil
}
