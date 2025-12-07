package di

import (
	"context"
	"database/sql"

	domain "math-ai.com/math-ai/internal/core/domain/grade"
	"math-ai.com/math-ai/internal/shared/utils/pagination"
)

type ListGradesParams struct {
	Search    string
	Page      int64
	Limit     int64
	OrderBy   string
	OrderDesc bool
	TakeAll   bool
}

type IGradeRepository interface {
	List(ctx context.Context, params ListGradesParams) ([]*domain.Grade, *pagination.Pagination, error)
	FindByID(ctx context.Context, id string) (*domain.Grade, error)
	FindByLabel(ctx context.Context, label string) (*domain.Grade, error)
	Create(ctx context.Context, tx *sql.Tx, grade *domain.Grade) (int64, error)
	Update(ctx context.Context, grade *domain.Grade) (int64, error)
	Delete(ctx context.Context, id string) error
	ForceDelete(ctx context.Context, tx *sql.Tx, id string) error
}
