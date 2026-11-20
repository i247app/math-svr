package grade

import "math-ai.com/math-ai/internal/domain/shared/status"

// Status codes of this aggregate (4400–4499). They are wire contract: the
// client reads them from "mstatus", so never renumber or reuse one.
// Moved from domain/shared/status (docs/features/014-status-registry).
const (
	GradeNotFound            status.StatusCode = 4401
	GradeMissingID           status.StatusCode = 4403
	GradeMissingLabel        status.StatusCode = 4404
	GradeMissingDescription  status.StatusCode = 4405
	GradeInvalidDisplayOrder status.StatusCode = 4406
)

func init() {
	status.Register(GradeNotFound, "Grade not found", "Không tìm thấy lớp")
	status.Register(GradeMissingID, "Grade id is required", "Mã lớp là bắt buộc")
	status.Register(GradeMissingLabel, "Grade label is required", "Tên lớp là bắt buộc")
	status.Register(GradeMissingDescription, "Grade description is required", "Mô tả lớp là bắt buộc")
	status.Register(GradeInvalidDisplayOrder, "Grade display order is invalid", "Thứ tự hiển thị của lớp không hợp lệ")
}
