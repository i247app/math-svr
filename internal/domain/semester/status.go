package semester

import "math-ai.com/math-ai/internal/domain/shared/status"

// Status codes of this aggregate (4500–4599). They are wire contract: the
// client reads them from "mstatus", so never renumber or reuse one.
// Moved from domain/shared/status (docs/features/014-status-registry).
const (
	SemesterNotFound            status.StatusCode = 4501
	SemesterMissingID           status.StatusCode = 4503
	SemesterMissingName         status.StatusCode = 4504
	SemesterMissingDescription  status.StatusCode = 4505
	SemesterInvalidDisplayOrder status.StatusCode = 4506
)

func init() {
	status.Register(SemesterNotFound, "Semester not found", "Không tìm thấy học kỳ")
	status.Register(SemesterMissingID, "Semester id is required", "Mã học kỳ là bắt buộc")
	status.Register(SemesterMissingName, "Semester name is required", "Tên học kỳ là bắt buộc")
	status.Register(SemesterMissingDescription, "Semester description is required", "Mô tả học kỳ là bắt buộc")
	status.Register(SemesterInvalidDisplayOrder, "Semester display order is invalid", "Thứ tự hiển thị của học kỳ không hợp lệ")
}
