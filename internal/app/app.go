package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"slices"

	"github.com/i247app/gex"
	"math-ai.com/math-ai/internal/app/resources"
	"math-ai.com/math-ai/internal/app/routes"
	"math-ai.com/math-ai/internal/app/services"
	"math-ai.com/math-ai/internal/driven-adapter/jobs"
	"math-ai.com/math-ai/internal/handlers/http/middleware"
	"math-ai.com/math-ai/internal/session"
	"math-ai.com/math-ai/internal/shared/config"
	"math-ai.com/math-ai/internal/shared/constant/status"
	"math-ai.com/math-ai/internal/shared/db"
	"math-ai.com/math-ai/internal/shared/utils/response"
)

func NewFromEnv(envPath string) (*App, error) {
	// Load configuration
	env, err := config.NewEnv(envPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Initialize logger
	// logger.Initialize(env.HostConfig.ServerMode)

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
	app.Database = database

	routes.SetUpHttpRoutes(app.Server, &resources, app.Services)

	return app, nil
}

func (a *App) Init() error {
	services, err := services.SetupServiceContainer(a.Resource)
	if err != nil {
		return fmt.Errorf("failed to setup services: %w", err)
	}
	a.Services = services
	a.Resource.SessionManager = services.SessionManager

	defaultRouteHandler := func(w http.ResponseWriter, r *http.Request) {
		response.WriteJson(w, r.Context(), nil, fmt.Errorf("route not found"), status.NOT_FOUND)
	}
	a.Server = gex.NewServer(a.Resource.HostConfig, defaultRouteHandler)

	// Register middlewares
	a.setupMiddleware(a.Server, services)

	// Setup jobs
	a.setupJobs(a.Server, a.Services)

	// Setup shutdown hooks
	a.setupShutdownHooks(a.Server, services)

	// Reload sessions
	a.reloadSessions()

	return nil
}

func (a *App) Start() error {
	return a.Server.Start()
}

func (a *App) setupJobs(_ *gex.Server, _ *services.ServiceContainer) {
	testJob := jobs.NewTestJob()

	a.JobManager.RegisterJob(testJob)

	// Start the job manager
	a.JobManager.Start()
}

func (a *App) setupShutdownHooks(gexServer *gex.Server, _ *services.ServiceContainer) {
	gexServer.OnShutdown(func() {
		sessionFile := a.Resource.Env.SerializedSessionFile
		if sessionFile == "" {
			return
		}

		log.Println("Serializing sessions...")
		err := a.SerializeSessions(sessionFile)
		if err != nil {
			log.Printf("Failed to serialize sessions: %v\n", err)
		} else {
			log.Println("Sessions serialized!")
		}
	})
}

// Setup middlewares
func (a *App) setupMiddleware(gexSvr *gex.Server, services *services.ServiceContainer) {
	middlewares := []gex.Middleware{
		// Start-->
		middleware.GexSessionMiddleware(services.SessionProvider, session.SessionContextKey),
		middleware.LoggerMiddleware(a.Resource.Env.LogFile),
		middleware.ValidateSessionMiddleware,
		middleware.LogRequestMiddleware,
		middleware.LocaleMiddleware("en"),
		// -->End
	}

	slices.Reverse(middlewares) // Reverse the middleware order so that the first middleware in the slice is the first to run
	for _, middleware := range middlewares {
		gexSvr.RegisterMiddleware(middleware)
	}

	gexSvr.SetupServerCORS()
}

func (a *App) Close() error {
	return a.Database.Close()
}

func (a *App) reloadSessions() {
	// Reload old sessions
	if a.Resource.Env.SerializedSessionFile != "" {
		log.Println("Reloading old sessions...")
		if err := a.ReloadSessions(a.Resource.Env.SerializedSessionFile); err != nil {
			log.Printf("Failed to reload old sessions: %v\n", err)
		}
	}
}
