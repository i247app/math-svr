package bootstrap

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"slices"
	"time"

	"github.com/i247app/gex"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"math-ai.com/math-ai/internal/application/resource"
	"math-ai.com/math-ai/internal/bootstrap/container"
	"math-ai.com/math-ai/internal/bootstrap/middleware"
	"math-ai.com/math-ai/internal/bootstrap/routes"
	"math-ai.com/math-ai/internal/infrastructure/config"
	"math-ai.com/math-ai/internal/infrastructure/database"
	"math-ai.com/math-ai/internal/infrastructure/logger"
	"math-ai.com/math-ai/internal/infrastructure/session"
	"math-ai.com/math-ai/internal/infrastructure/tracing"
	"math-ai.com/math-ai/internal/jobs"
	"math-ai.com/math-ai/internal/shared/response"
)

func NewFromEnv(envPath string) (*App, error) {
	// Load configuration
	env, err := config.NewEnv(envPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Log encodings are env-driven and per-destination. Default split:
	// console=text (readable while running `make run`), file=json (tailed by
	// Alloy → Loki). LOG_FORMAT, if set, overrides both (uniform mode).
	consoleFmt := env.LogConsoleFormat
	fileFmt := env.LogFileFormat
	if env.LogFormat != "" {
		consoleFmt = env.LogFormat
		fileFmt = env.LogFormat
	}
	logOpts := logger.Options{
		Level:         slog.LevelInfo,
		WithColor:     true,
		ConsoleFormat: logger.Format(consoleFmt),
		FileFormat:    logger.Format(fileFmt),
	}
	logProvider, err := logger.NewProvider(logOpts, env.LogFile)
	if err != nil {
		return nil, fmt.Errorf("failed to setup logger: %w", err)
	}
	logger.SetDefault(logProvider)

	// Initialise distributed tracing early so the logger trace-id bridge is
	// active before any span is created. Best-effort: a failure here (e.g.
	// bad OTLP endpoint) must not stop the server — log and continue with
	// tracing effectively off.
	traceShutdown, err := tracing.Init(context.Background(), env.ObservabilityConfig)
	if err != nil {
		logProvider.Background(context.Background()).
			Warnf("tracing init failed, continuing without tracing: %v", err)
		traceShutdown = func(context.Context) error { return nil }
	}

	// Breadcrumb: this is the last step before the "[DB] Connected" line. If the
	// process stalls here, DB connect is hanging (unreachable host or TLS
	// mismatch — the driver forces TLS 1.2), not erroring.
	logProvider.Background(context.Background()).
		Infof("connecting to database host=%s:%s db=%s user=%s ...",
			env.DBConfig.DBHost, env.DBConfig.DBPort, env.DBConfig.DBName, env.DBConfig.DBUser)

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
	app.TracerShutdown = traceShutdown
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
	a.setupMiddleware(a.Server, a.Resource, services)

	// Setup jobs
	a.setupJobs(a.Server, services)

	// Setup the dedicated metrics listener (separate from the API server, so
	// /metrics never passes through the app middleware chain).
	a.setupMetricsServer()

	// Setup shutdown hooks
	a.setupShutdownHooks(a.Server, services)

	// Reload sessions
	a.reloadSessions()

	return nil
}

// Setup middlewares
func (a *App) setupMiddleware(gexSvr *gex.Server, res *resource.Resource, _ *container.ServiceContainer) {
	middlewares := []gex.Middleware{
		// Start-->
		// Metrics is outermost so it measures the full request lifetime.
		// No-op when res.Metrics is nil (metrics disabled).
		middleware.MetricsMiddleware(res.Metrics, res.RouteClassifier),
		// Tracing runs just inside metrics and BEFORE the logger, so every
		// log line for the request can carry its trace_id. Pass-through when
		// tracing is disabled.
		middleware.TracingMiddleware(res.Env.ObservabilityConfig.TracingEnabled, res.RouteClassifier),
		// SessionTokenMiddleware runs just outside the session middleware: the
		// session token is taken ONLY from the request body's
		// metadata.authorization and lifted into the Authorization header so
		// GexSessionMiddleware resolves the session. Any client-supplied
		// Authorization header is dropped — the header is never trusted.
		middleware.SessionTokenMiddleware,
		middleware.GexSessionMiddleware(res.SessionProvider, session.SessionContextKey),
		middleware.LoggerMiddleware(a.Logger, res),
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

// setupJobs populates the JobRegistry with every concrete CronJob /
// TaskHandler this binary should run, then starts the JobRuntime. The
// catalogue itself lives in internal/jobs/RegisterAll so that adding a
// new job does not require editing bootstrap.
//
// Order matters: jobs.RegisterAll must finish before runtime.Start —
// Start reads the registry to seed its per-job state.
func (a *App) setupJobs(_ *gex.Server, _ *container.ServiceContainer) {
	jobs.RegisterAll(a.Resource.JobRegistry, jobs.Deps{
		Runtime:        a.Resource.JobRuntime,
		SessionManager: a.Resource.SessionManager,
		EmailProvider:  a.Resource.EmailProvider,
	})
	a.Resource.JobRuntime.Start(context.Background())
}

// setupMetricsServer starts a small, middleware-free HTTP server that serves
// only /metrics (and /ping for a liveness probe). Running metrics on its own
// listener keeps the Prometheus exposition format out of the app's gzip /
// session / logging middleware and off the public API port. No-op when
// metrics are disabled.
func (a *App) setupMetricsServer() {
	if a.Resource.Metrics == nil {
		return
	}
	addr := a.Resource.Env.ObservabilityConfig.MetricsAddr

	mux := http.NewServeMux()
	mux.Handle("GET /metrics", promhttp.HandlerFor(
		a.Resource.Metrics.Registry,
		// EnableOpenMetrics is REQUIRED for exemplars: trace_id exemplars are
		// only rendered in the OpenMetrics exposition format. Prometheus (with
		// --enable-feature=exemplar-storage) negotiates it automatically.
		promhttp.HandlerOpts{EnableOpenMetrics: true},
	))
	mux.HandleFunc("GET /ping", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	a.MetricsServer = &http.Server{Addr: addr, Handler: mux}

	go func() {
		log.Printf("[APP] metrics server listening on %s/metrics", addr)
		if err := a.MetricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("[APP] metrics server error: %v", err)
		}
	}()
}

func (a *App) Start() error {
	return a.Server.Start()
}

func (a *App) Close() error {
	if a.MetricsServer != nil {
		log.Println("[APP] Shutting down metrics server...")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := a.MetricsServer.Shutdown(ctx); err != nil {
			log.Printf("[APP] metrics server shutdown error: %v", err)
		}
	}

	// Flush pending spans before tearing down the DB/logger so late spans
	// still export. No-op when tracing is disabled.
	if a.TracerShutdown != nil {
		log.Println("[APP] Flushing traces...")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := a.TracerShutdown(ctx); err != nil {
			log.Printf("[APP] tracer shutdown error: %v", err)
		}
	}

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

func (a *App) reloadSessions() {
	// Reload old sessions
	if a.Resource.Env.SerializedSessionFile != "" {
		log.Println("Reloading old sessions...")
		if err := a.ReloadSessions(a.Resource.Env.SerializedSessionFile); err != nil {
			log.Printf("Failed to reload old sessions: %v\n", err)
		}
	}
}

func (a *App) setupShutdownHooks(gexSvr *gex.Server, _ *container.ServiceContainer) {
	gexSvr.OnShutdown(func() {
		// Close every WebSocket first (StatusGoingAway) so clients get a clean
		// close frame and can reconnect, instead of hanging on a dropped socket.
		if a.Resource.SocketHub != nil {
			log.Println("Closing WebSocket connections...")
			a.Resource.SocketHub.Shutdown("server shutting down")
		}

		// Drain the job runtime first: stops new schedules from
		// firing, cancels in-flight execution contexts, waits up to
		// JobRuntime.Config().DrainTimeout for graceful exit. Done
		// before session serialisation so any job that touches
		// sessions has finished before we snapshot them.
		if a.Resource.JobRuntime != nil {
			a.Resource.JobRuntime.Stop(context.Background())
		}

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
