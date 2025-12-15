package di

import (
	"context"
	"database/sql"

	domain "math-ai.com/math-ai/internal/core/domain/term"
	"math-ai.com/math-ai/internal/shared/utils/pagination"
)

type ListTermsParams struct {
	Search    string
	Page      int64
	Limit     int64
	OrderBy   string
	OrderDesc bool
	TakeAll   bool
}

type ITermRepository interface {
	List(ctx context.Context, params ListTermsParams) ([]*domain.Term, *pagination.Pagination, error)
	FindByID(ctx context.Context, id string) (*domain.Term, error)
	FindByName(ctx context.Context, name string) (*domain.Term, error)
	Create(ctx context.Context, tx *sql.Tx, term *domain.Term) (int64, error)
	Update(ctx context.Context, term *domain.Term) (int64, error)
	Delete(ctx context.Context, id string) error
	ForceDelete(ctx context.Context, tx *sql.Tx, id string) error
}
