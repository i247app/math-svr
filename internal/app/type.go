package app

import (
	"context"

	"github.com/i247app/gex"
	"math-ai.com/math-ai/internal/app/resources"
	"math-ai.com/math-ai/internal/app/services"
	domain "math-ai.com/math-ai/internal/core/domain/jobs"
	"math-ai.com/math-ai/internal/shared/db"
)

type App struct {
	Server     *gex.Server
	Services   *services.ServiceContainer
	JobManager *domain.JobManager

	Resource *resources.AppResource
	Database db.IDatabase
}

func NewApp(resource *resources.AppResource) *App {
	return &App{
		Resource:   resource,
		JobManager: domain.NewJobManager(context.Background()),
	}
}
