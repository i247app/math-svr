package app

import (
	"context"
	"fmt"
	"net/http"
	"slices"

	"github.com/i247app/gex"
	"math-ai.com/math-ai/internal/app/resources"
	"math-ai.com/math-ai/internal/app/routes"
	"math-ai.com/math-ai/internal/app/services"
	"math-ai.com/math-ai/internal/handlers/http/middleware"
	"math-ai.com/math-ai/internal/shared/config"
	"math-ai.com/math-ai/internal/shared/constant/status"
	"math-ai.com/math-ai/internal/shared/db"
	"math-ai.com/math-ai/internal/shared/logger"
	"math-ai.com/math-ai/internal/shared/utils/response"
)

func NewFromEnv(envPath string) (*App, error) {
	// Load configuration
	env, err := config.NewEnv(envPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Initialize logger
	logger.Initialize(env.HostConfig.ServerMode)

	// Initialize database connection
	database, err := db.NewDatabase(env.DBEnv)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to the database: %w", err)
	}

	// Ping the database
	if err := database.PingContext(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping the database: %w", err)
	}

	// Build app resource
	hostConfig := gex.HostConfig{
		ServerHost: env.HostConfig.ServerHost,
		ServerPort: env.HostConfig.ServerPort,
	}
	if env.HostConfig.HttpsCertFile != nil {
		hostConfig.HttpsCertFile = *env.HostConfig.HttpsCertFile
	}
	if env.HostConfig.HttpsKeyFile != nil {
		hostConfig.HttpsKeyFile = *env.HostConfig.HttpsKeyFile
	}
	resources := resources.AppResource{
		Env:        env,
		HostConfig: hostConfig,
		Db:         database,
	}

	app := NewApp(&resources)
	if err := app.Init(); err != nil {
		return nil, fmt.Errorf("failed to init app: %w", err)
	}

	routes.SetUpHttpRoutes(app.Server, &resources, app.Services)

	return app, nil
}

func (a *App) Init() error {
	services, err := services.SetupServiceContainer(a.Resource)
	if err != nil {
		return fmt.Errorf("failed to setup services: %w", err)
	}
	a.Services = services

	defaultRouteHandler := func(w http.ResponseWriter, r *http.Request) {
		response.WriteJson(w, r.Context(), nil, fmt.Errorf("route not found"), status.NOT_FOUND)
	}
	a.Server = gex.NewServer(a.Resource.HostConfig, defaultRouteHandler)

	// Register middlewares
	a.setupMiddleware(a.Server, services)

	return nil
}

func (a *App) Start() error {
	logger.Infof("Starting server on %s:%s", a.Resource.HostConfig.ServerHost, a.Resource.HostConfig.ServerPort)
	return a.Server.Start()
}

// Setup middlewares
func (a *App) setupMiddleware(gexSvr *gex.Server, _ *services.ServiceContainer) {
	logger.Info("Setup middlewares...")
	// Middleware are run in order of declaration
	// The first middleware in the slice runs first
	middlewares := []gex.Middleware{
		// Start-->
		middleware.LocaleMiddleware("en"),
		middleware.LogRequestMiddleware,
		// -->End
	}

	slices.Reverse(middlewares) // Reverse the middleware order so that the first middleware in the slice is the first to run
	for _, middleware := range middlewares {
		gexSvr.RegisterMiddleware(middleware)
	}

	gexSvr.SetupServerCORS()
}

func (a *App) Close() error {
	return nil
}
