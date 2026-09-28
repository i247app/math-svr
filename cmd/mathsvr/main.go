package main

import (
	"fmt"
	"log"
	"os"

	// Embed the IANA timezone database in the binary. Cron schedules
	// resolve their Hour/Minute against a *time.Location, and
	// /jobs/schedule/update lets an operator name any IANA zone — both
	// go through time.LoadLocation, which otherwise reads
	// /usr/share/zoneinfo from the host. A deploy target without tzdata
	// installed would make every zone silently resolve to UTC and fire
	// jobs at the wrong local time. ~450KB for a deterministic answer on
	// every host.
	_ "time/tzdata"

	"math-ai.com/math-ai/internal/bootstrap"
	"math-ai.com/math-ai/internal/module/server"
)

const envPath = ".env"

func main() {
	// Pre-flight child of POST /server/reload: validate .env and exit.
	if server.IsEnvCheckRun() {
		os.Exit(server.RunEnvCheck(envPath))
	}

	// Surface startup failures. run() wraps every fatal error with context
	// (config load, DB connect, resource setup); previously main swallowed it
	// with a bare return, so the process exited with no reason logged.
	if err := run(); err != nil {
		log.Printf("startup failed: %+v", err)
		return
	}

	// POST /server/reload: the graceful shutdown has run (hooks, app.Close),
	// so start over as a fresh process. A failed exec exits non-zero so
	// systemd's Restart=on-failure brings the server back instead.
	if server.ReloadRequested() {
		log.Println("Reloading server...")
		if err := server.Reexec(); err != nil {
			log.Printf("reload failed: %v", err)
			os.Exit(1)
		}
	}
}

func run() error {
	// Initialize app
	app, err := bootstrap.NewFromEnv(envPath)
	if err != nil {
		return fmt.Errorf("failed to initialize app: %w", err)
	}
	// The shutdown hook already calls Close after a normal shutdown; this
	// covers Start failing before any hook ran (e.g. the port is taken), so
	// the log file is still flushed. Close runs only once.
	defer app.Close()

	// Start app
	log.Println("Starting server...")
	if err := app.Start(); err != nil {
		return fmt.Errorf("failed to start app: %w", err)
	}

	return nil
}
