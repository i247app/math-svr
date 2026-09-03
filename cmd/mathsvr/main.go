package main

import (
	"fmt"
	"log"

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
)

func main() {
	// Surface startup failures. run() wraps every fatal error with context
	// (config load, DB connect, resource setup); previously main swallowed it
	// with a bare return, so the process exited with no reason logged.
	if err := run(); err != nil {
		log.Printf("startup failed: %+v", err)
		return
	}
}

func run() error {
	// Initialize app
	app, err := bootstrap.NewFromEnv(".env")
	if err != nil {
		return fmt.Errorf("failed to initialize app: %w", err)
	}
	defer app.Close()

	// Start app
	log.Println("Starting server...")
	if err := app.Start(); err != nil {
		return fmt.Errorf("failed to start app: %w", err)
	}

	return nil
}
