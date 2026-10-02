package pagination

import "math"

const (
	DefaultPageSize int64 = 20
	MaxPageSize     int64 = 1000
)

type Pagination struct {
	Page        int64 `json:"page"`
	Size        int64 `json:"size"`
	TakeAll     bool  `json:"take_all"`
	Skip        int64 `json:"skip"`
	TotalCount  int64 `json:"total_count"`
	TotalPages  int64 `json:"total_pages"`
	HasPrevious bool  `json:"has_previous"`
	HasNext     bool  `json:"has_next"`
}

// ClampSize is the page-size rule shared by offset and cursor paging:
// anything outside [1, MaxPageSize] falls back to DefaultPageSize.
func ClampSize(size int64) int64 {
	if size > 0 && size <= MaxPageSize {
		return size
	}
	return DefaultPageSize
}

func NewPagination(page int64, size int64, total int64) *Pagination {
	var pageInfo Pagination
	pageInfo.Size = ClampSize(size)

	totalPage := int64(math.Ceil(float64(total) / float64(pageInfo.Size)))
	pageInfo.TotalCount = total
	pageInfo.TotalPages = totalPage
	if page < 1 || totalPage == 0 {
		page = 1
	}

	pageInfo.Page = page
	pageInfo.Skip = (page - 1) * pageInfo.Size
	pageInfo.HasPrevious = page > 1
	pageInfo.HasNext = page < totalPage

	return &pageInfo
}
