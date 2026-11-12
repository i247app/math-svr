package storage

import (
	"context"

	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/infrastructure/config"
	"math-ai.com/math-ai/internal/pkg/s3"

	errs "math-ai.com/math-ai/internal/domain/shared/error"
)

func NewFromConfig(ctx context.Context, cfg config.StorageConfig) (*Registry, error) {
	switch cfg.Provider {
	case string(ProviderS3), "":
		client := s3.NewClient(s3.Config{
			AccessKey: cfg.AccessKey,
			SecretKey: cfg.SecretKey,
			Region:    cfg.Region,
			Bucket:    cfg.Bucket,
		})

		reg := NewRegistry()
		reg.Register(NewS3Provider(client))
		// Register auto-elects the first provider as default; explicit
		// SetDefault here would be redundant.
		return reg, nil

	default:
		return nil, errs.NewError(ctx, status.STORAGE_CONFIG_INVALID,
			map[string]any{"provider": cfg.Provider}, nil)
	}
}
