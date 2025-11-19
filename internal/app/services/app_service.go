package services

import (
	"math-ai.com/math-ai/internal/app/resources"
	"math-ai.com/math-ai/internal/shared/logger"
)

type ServiceContainer struct{}

func SetupServiceContainer(res *resources.AppResource) (*ServiceContainer, error) {
	logger.Info("Initializing services")

	return &ServiceContainer{}, nil
}
