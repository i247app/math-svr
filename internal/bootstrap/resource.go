package bootstrap

import (
	"context"
	"fmt"

	"math-ai.com/math-ai/internal/adapter/email"
	"math-ai.com/math-ai/internal/adapter/sms"
	"math-ai.com/math-ai/internal/adapter/storage"
	"math-ai.com/math-ai/internal/application/resource"
	"math-ai.com/math-ai/internal/infrastructure/logger"
)

func SetupResource(res *resource.Resource) error {
	log := logger.From(context.Background())

	env := res.Env

	log.Info("> Setup Resource...")
	log.Info("> Setup EmailAdapter...")
	log.Infof("> Email Provider: %s", env.EmailConfig.EmailProvider)
	emailAdapter, err := email.NewFromConfig(context.Background(), env.EmailConfig)
	if err != nil {
		return fmt.Errorf("failed to setup email adapter: %w", err)
	}
	res.EmailProvider = emailAdapter

	log.Info("> Setup SMSAdapter...")
	log.Infof("> SMS Provider: %s", env.SMSConfig.SMSProvider)
	smsAdapter, err := sms.NewFromConfig(context.Background(), env.SMSConfig)
	if err != nil {
		return fmt.Errorf("failed to setup sms adapter: %w", err)
	}
	res.SMSProvider = smsAdapter

	log.Info("> Setup StorageAdapter...")
	log.Infof("> Storage Provider: %s", env.StorageConfig.Provider)
	storageAdapter, err := storage.NewFromConfig(context.Background(), env.StorageConfig)
	if err != nil {
		return fmt.Errorf("failed to setup storage adapter: %w", err)
	}
	res.StorageProvider = storageAdapter

	return nil
}
