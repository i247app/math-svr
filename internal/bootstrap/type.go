package bootstrap

import (
	"context"
	"net/http"

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

	// MetricsServer serves /metrics on its own middleware-free listener
	// (OBS_METRICS_ADDR). Nil when metrics are disabled.
	MetricsServer *http.Server

	// TracerShutdown flushes and stops the OpenTelemetry TracerProvider.
	// Always non-nil (a no-op when tracing is disabled); called from Close.
	TracerShutdown func(context.Context) error
}

func NewApp(resource *resource.Resource) *App {
	return &App{
		Resource: resource,
	}
}
