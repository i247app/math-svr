package semester

import (
	"context"

	"math-ai.com/math-ai/internal/shared/pagination"
)

// IRepository owns ma_semesters persistence.
type IRepository interface {
	FindBySemesterId(ctx context.Context, semesterId int64) (*Semester, error)
	ListSemesters(ctx context.Context, params *ListSemestersParams) ([]*Semester, *pagination.Pagination, error)
	// ListSemestersByIds resolves a set of semesters in one query. Returns nil
	// slice on empty input; caller maps by SemesterId().
	ListSemestersByIds(ctx context.Context, ids []int64) ([]*Semester, error)
	Create(ctx context.Context, s *Semester) (*Semester, error)
	Update(ctx context.Context, s *Semester) error
	SoftDeleteBySemesterId(ctx context.Context, semesterId int64) error
	ForceDeleteBySemesterId(ctx context.Context, semesterId int64) error
}

type ListSemestersParams struct {
	Page    int64
	Limit   int64
	TakeAll bool
}
