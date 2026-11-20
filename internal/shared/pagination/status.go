package pagination

import "math-ai.com/math-ai/internal/domain/shared/status"

// Status codes of this aggregate (14000–14099). They are wire contract: the
// client reads them from "mstatus", so never renumber or reuse one.
// Moved from domain/shared/status (docs/features/014-status-registry).
const (
	// Pagination — 14000-14099. Shared by every list endpoint that supports
	// both offset and cursor paging (enum.PaginationType), so a list adding
	// cursor paging reuses these instead of minting its own.
	PaginationInvalidType    status.StatusCode = 14000 // pagination_type is neither OFFSET nor CURSOR
	PaginationInvalidCursor  status.StatusCode = 14001 // cursor was not issued by this list
	PaginationParamsConflict status.StatusCode = 14002 // page with CURSOR, next/previous with OFFSET, or next AND previous
)

func init() {
	status.Register(PaginationInvalidType, "Invalid pagination type", "Kiểu phân trang không hợp lệ")
	status.Register(PaginationInvalidCursor, "Invalid pagination cursor", "Con trỏ phân trang không hợp lệ")
	status.Register(PaginationParamsConflict, "Send page only with OFFSET pagination, and at most one of next / previous only with CURSOR pagination", "Chỉ gửi page khi phân trang OFFSET; khi phân trang CURSOR chỉ gửi một trong hai next hoặc previous")
}
