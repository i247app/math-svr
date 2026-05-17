package semester

import (
	"context"

	"github.com/google/uuid"
	"math-ai.com/math-ai/internal/shared/enum"
	"math-ai.com/math-ai/internal/shared/pagination"
)

type IRepository interface {
	ListSemesters(ctx context.Context, params *ListSemestersParams) ([]*Semester, *pagination.Pagination, error)
	// ListSemestersByIds resolves a set of semesters in one query. Returns nil
	// slice on empty input; caller maps by SemesterId().
	ListSemestersByIds(ctx context.Context, ids []uuid.UUID, language enum.LanguageType) ([]*Semester, error)
}

type ListSemestersParams struct {
	Language enum.LanguageType
	Page     int64
	Limit    int64
	TakeAll  bool
}
