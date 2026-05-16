package bootstrap

import (
	"math-ai.com/math-ai/internal/bootstrap/container"
	"math-ai.com/math-ai/internal/infrastructure/logger"

	"github.com/i247app/gex"
	"math-ai.com/math-ai/internal/application/resource"
)

type App struct {
	Server   *gex.Server
	Services *container.ServiceContainer
	Resource *resource.Resource
	Logger   *logger.Provider
}

func NewApp(resource *resource.Resource) *App {
	return &App{
		Resource: resource,
	}
}
