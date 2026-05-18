package bootstrap

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"slices"

	"math-ai.com/math-ai/internal/bootstrap/container"
	"math-ai.com/math-ai/internal/bootstrap/middleware"
	"math-ai.com/math-ai/internal/bootstrap/routes"
	"math-ai.com/math-ai/internal/infrastructure/config"
	"math-ai.com/math-ai/internal/infrastructure/database"
	"math-ai.com/math-ai/internal/infrastructure/logger"
	"math-ai.com/math-ai/internal/shared/response"

	"github.com/i247app/gex"
	"math-ai.com/math-ai/internal/application/resource"
)

func NewFromEnv(envPath string) (*App, error) {
	// Load configuration
	env, err := config.NewEnv(envPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	logProvider, err := logger.NewProvider(logger.Options{Level: slog.LevelInfo, WithColor: true}, env.LogFile)
	if err != nil {
		return nil, fmt.Errorf("failed to setup logger: %w", err)
	}
	logger.SetDefault(logProvider)

	sqlDB, err := database.NewSqlDB(env.DBConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// // Run migrations
	// if err := database.Migrate(context.Background(), sqlDB, "migrations"); err != nil {
	// 	sqlDB.Close()
	// 	_ = logProvider.Close()
	// 	return nil, fmt.Errorf("failed to run migrations: %w", err)
	// }

	db := database.NewDatabaseWithLogs(sqlDB)

	// Build app resource
	hostConfig := gex.HostConfig{
		ServerHost: env.ServerHost,
		ServerPort: env.ServerPort,
	}
	if env.HttpsCertFile != nil {
		hostConfig.HttpsCertFile = *env.HttpsCertFile
	}
	if env.HttpsKeyFile != nil {
		hostConfig.HttpsKeyFile = *env.HttpsKeyFile
	}
	resource := resource.Resource{
		Env:        env,
		HostConfig: hostConfig,
		DB:         db,
	}

	if err := SetupResource(&resource); err != nil {
		return nil, fmt.Errorf("failed to setup resource: %w", err)
	}

	// Initialize app
	app := NewApp(&resource)
	app.Logger = logProvider
	if err := app.Init(); err != nil {
		db.Close()
		_ = logProvider.Close()
		return nil, fmt.Errorf("failed to init app: %w", err)
	}

	routes.SetupHttpRoutes(app.Server, &resource, app.Services)

	return app, nil
}

func (a *App) Init() error {
	// Build service container
	services, err := container.SetupServiceContainer(a.Resource)
	if err != nil {
		return fmt.Errorf("failed to setup services: %w", err)
	}
	a.Services = services

	// Setup gex.Server
	defaultRouteHandler := func(w http.ResponseWriter, r *http.Request) {
		response.WriteJson(w, nil, fmt.Errorf("route not found: %v", r.URL.Path))
	}

	a.Server = gex.NewServer(
		a.Resource.HostConfig,
		defaultRouteHandler,
	)

	// Register middlewares
	a.setupMiddleware(a.Server, services)

	return nil
}

// Setup middlewares
func (a *App) setupMiddleware(gexSvr *gex.Server, _ *container.ServiceContainer) {
	middlewares := []gex.Middleware{
		// Start-->
		middleware.LoggerMiddleware(a.Logger),
		middleware.LogRequestMiddleware,
		middleware.MetadataMiddleware(),
		middleware.RecoveryMiddleware,
		middleware.GzipMiddleware,
		// -->End
	}

	slices.Reverse(middlewares)
	for _, mid := range middlewares {
		gexSvr.RegisterMiddleware(mid)
	}

	gexSvr.SetupServerCORS()
}

func (a *App) Start() error {
	return a.Server.Start()
}

func (a *App) Close() error {
	if a.Resource.DB != nil {
		log.Println("[APP] Closing database connection...")
		if err := a.Resource.DB.Close(); err != nil {
			log.Printf("[APP] database close error: %v", err)
		}
	}

	if a.Logger != nil {
		log.Println("[APP] Closing log file...")
		if err := a.Logger.Close(); err != nil {
			log.Printf("[APP] logger close error: %v", err)
		}
	}
	return nil
}
