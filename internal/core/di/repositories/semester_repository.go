package di

import (
	"context"
	"database/sql"

	domain "math-ai.com/math-ai/internal/core/domain/semester"
	"math-ai.com/math-ai/internal/shared/db"
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
	DoTransaction(ctx context.Context, handler db.HanderlerWithTx) error

	List(ctx context.Context, params ListSemestersParams) ([]*domain.Semester, *pagination.Pagination, error)
	FindByID(ctx context.Context, id string) (*domain.Semester, error)
	FindByName(ctx context.Context, name string) (*domain.Semester, error)
	Create(ctx context.Context, tx *sql.Tx, semester *domain.Semester) (int64, error)
	Update(ctx context.Context, tx *sql.Tx, semester *domain.Semester) (int64, error)
	Delete(ctx context.Context, tx *sql.Tx, id string) error
	ForceDelete(ctx context.Context, tx *sql.Tx, id string) error
	CreateSemesterTranslation(ctx context.Context, tx *sql.Tx, semesterTranslation *domain.SemesterTranslation) (int64, error)
	UpdateSemesterTranslation(ctx context.Context, tx *sql.Tx, semesterTranslation *domain.SemesterTranslation) (int64, error)
	DeleteSemesterTranslationsBySemesterID(ctx context.Context, tx *sql.Tx, semesterID string) error
	ForceDeleteSemesterTranslationsBySemesterID(ctx context.Context, tx *sql.Tx, semesterID string) error
}
