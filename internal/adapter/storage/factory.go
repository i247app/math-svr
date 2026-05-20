package storage

import (
	"context"

	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/infrastructure/config"
	"math-ai.com/math-ai/internal/libs/s3"

	errs "math-ai.com/math-ai/internal/domain/shared/error"
)

func NewFromConfig(ctx context.Context, cfg config.StorageConfig) (*Adapter, error) {
	println("AccessKey", cfg.AccessKey)
	println("SecretKey", cfg.SecretKey)
	println("Region", cfg.Region)
	println("Bucket", cfg.Bucket)

	switch cfg.Provider {
	case string(ProviderS3), "":
		client := s3.NewClient(s3.Config{
			AccessKey: cfg.AccessKey,
			SecretKey: cfg.SecretKey,
			Region:    cfg.Region,
			Bucket:    cfg.Bucket,
		})

		adapter := NewAdapter()
		adapter.Register(NewS3Provider(client))
		// Register auto-elects the first provider as default; explicit
		// SetDefault here would be redundant.
		return adapter, nil

	default:
		return nil, errs.NewError(ctx, status.STORAGE_CONFIG_INVALID,
			map[string]any{"provider": cfg.Provider}, nil)
	}
}
