package enum

// PaginationType selects how a list endpoint pages its results. A list that
// supports both takes it as `pagination_type` in the request body; the
// default for an empty value is the list's own decision (/users/list:
// CURSOR).
//
//   - OFFSET: page + size, with total_count / total_pages. Can jump to any
//     page — the admin-table shape. Costs a COUNT(*) and an OFFSET scan.
//   - CURSOR: size + an opaque cursor from the previous page (next or
//     previous). Stable while rows are inserted or deleted, no COUNT — the
//     mobile infinite-scroll shape.
type PaginationType string

const (
	PaginationTypeOffset PaginationType = "OFFSET"
	PaginationTypeCursor PaginationType = "CURSOR"
)

func (t PaginationType) String() string {
	return string(t)
}

func (t PaginationType) IsValid() bool {
	switch t {
	case PaginationTypeOffset, PaginationTypeCursor:
		return true
	default:
		return false
	}
}

func ListPaginationTypes() []string {
	return []string{
		PaginationTypeOffset.String(),
		PaginationTypeCursor.String(),
	}
}
