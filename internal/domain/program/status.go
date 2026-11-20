package program

import "math-ai.com/math-ai/internal/domain/shared/status"

// Status codes of this aggregate (4300–4399). They are wire contract: the
// client reads them from "mstatus", so never renumber or reuse one.
// Moved from domain/shared/status (docs/features/014-status-registry).
const (
	ProgramNotFound            status.StatusCode = 4301
	ProgramMissingID           status.StatusCode = 4303
	ProgramMissingLabel        status.StatusCode = 4304
	ProgramMissingDescription  status.StatusCode = 4305
	ProgramInvalidDisplayOrder status.StatusCode = 4306
)

func init() {
	status.Register(ProgramNotFound, "Program not found", "Không tìm thấy chương trình học")
	status.Register(ProgramMissingID, "Program id is required", "Mã chương trình học là bắt buộc")
	status.Register(ProgramMissingLabel, "Program label is required", "Tên chương trình học là bắt buộc")
	status.Register(ProgramMissingDescription, "Program description is required", "Mô tả chương trình học là bắt buộc")
	status.Register(ProgramInvalidDisplayOrder, "Program display order is invalid", "Thứ tự hiển thị của chương trình học không hợp lệ")
}
