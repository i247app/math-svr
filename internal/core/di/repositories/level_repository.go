package di

import (
	"context"
	"database/sql"

	domain "math-ai.com/math-ai/internal/core/domain/level"
	"math-ai.com/math-ai/internal/shared/utils/pagination"
)

type ListLevelsParams struct {
	Search    string
	Page      int64
	Limit     int64
	OrderBy   string
	OrderDesc bool
	TakeAll   bool
}

type ILevelRepository interface {
	List(ctx context.Context, params ListLevelsParams) ([]*domain.Level, *pagination.Pagination, error)
	FindByID(ctx context.Context, id string) (*domain.Level, error)
	FindByLabel(ctx context.Context, label string) (*domain.Level, error)
	Create(ctx context.Context, tx *sql.Tx, level *domain.Level) (int64, error)
	Update(ctx context.Context, level *domain.Level) (int64, error)
	Delete(ctx context.Context, id string) error
	ForceDelete(ctx context.Context, tx *sql.Tx, id string) error
}
