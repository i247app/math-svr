package repositories

import (
	"context"
	"database/sql"

	domain "math-ai.com/math-ai/internal/core/domain/profile"
)

type IProfileRepository interface {
	FindByID(ctx context.Context, id string) (*domain.Profile, error)
	FindByUID(ctx context.Context, uid string) (*domain.Profile, error)
	Create(ctx context.Context, tx *sql.Tx, profile *domain.Profile) (int64, error)
	Update(ctx context.Context, profile *domain.Profile) (int64, error)
	DeleteByUID(ctx context.Context, tx *sql.Tx, uid string) error
	ForceDeleteByUID(ctx context.Context, tx *sql.Tx, uid string) error
}
