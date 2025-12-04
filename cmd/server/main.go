package main

import (
	"fmt"

	"math-ai.com/math-ai/internal/app"
)

func main() {
	if err := run(); err != nil {
		//logger.Fatalf("application error: %w", err)
	}
}

func run() error {
	// Initialize app
	app, err := app.NewFromEnv(".env")
	if err != nil {
		////logger.Errorf("failed to initialize app: %w", err)
		return fmt.Errorf("failed to initialize app: %w", err)
	}
	defer app.Close()

	// Start app
	//logger.Info("Starting server...")
	if err := app.Start(); err != nil {
		return fmt.Errorf("failed to start app: %w", err)
	}

	return nil
}
