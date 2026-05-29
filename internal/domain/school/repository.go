package school

import (
	"context"

	"math-ai.com/math-ai/internal/shared/pagination"
)

// ListSchoolsParams narrows the listing/search query. Search is matched
// case-insensitively against name + district + province; District and
// Province narrow by exact match when set. TakeAll bypasses pagination
// for admin exports.
type ListSchoolsParams struct {
	Search   *string
	District *string
	Province *string
	Page     int64
	Limit    int64
	TakeAll  bool
}

// IRepository owns ma_schools persistence. There is no translation
// surface — name/description/district/province are single-language
// at the schema level today.
type IRepository interface {
	FindBySchoolId(ctx context.Context, schoolId string) (*School, error)
	ListSchools(ctx context.Context, params *ListSchoolsParams) ([]*School, *pagination.Pagination, error)
	// ListSchoolsByIds resolves a set of schools in one query. Returns nil
	// slice on empty input; caller maps by SchoolId(). Used by parent
	// aggregates (profile) to embed school responses without N+1.
	ListSchoolsByIds(ctx context.Context, ids []string) ([]*School, error)
	Create(ctx context.Context, s *School) (*School, error)
	Update(ctx context.Context, s *School) error
	SoftDeleteBySchoolId(ctx context.Context, schoolId string) error
	ForceDeleteBySchoolId(ctx context.Context, schoolId string) error
}
