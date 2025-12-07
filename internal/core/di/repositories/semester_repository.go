package di

import (
	"context"
	"database/sql"

	domain "math-ai.com/math-ai/internal/core/domain/semester"
	"math-ai.com/math-ai/internal/shared/utils/pagination"
)

type ListSemestersParams struct {
	Search    string
	Page      int64
	Limit     int64
	OrderBy   string
	OrderDesc bool
	TakeAll   bool
}

type ISemesterRepository interface {
	List(ctx context.Context, params ListSemestersParams) ([]*domain.Semester, *pagination.Pagination, error)
	FindByID(ctx context.Context, id string) (*domain.Semester, error)
	FindByName(ctx context.Context, name string) (*domain.Semester, error)
	Create(ctx context.Context, tx *sql.Tx, semester *domain.Semester) (int64, error)
	Update(ctx context.Context, semester *domain.Semester) (int64, error)
	Delete(ctx context.Context, id string) error
	ForceDelete(ctx context.Context, tx *sql.Tx, id string) error
}
