package app

import (
	"fmt"

	"math-ai.com/math-ai/internal/shared/config"
	"math-ai.com/math-ai/internal/shared/logger"
)

func NewFromEnv(envPath string) (*App, error) {
	// Load configuration
	env, err := config.NewEnv(envPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	logger.Initialize(env.HostConfig.ServerMode)

	return nil, nil
}

func (a *App) Init() error {
	return nil
}

func (a *App) Start() error {
	logger.Info("Server running on port ")
	return nil
}

func (a *App) Close() error {
	return nil
}
