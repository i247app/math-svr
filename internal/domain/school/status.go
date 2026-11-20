package school

import "math-ai.com/math-ai/internal/domain/shared/status"

// Status codes of this aggregate (4900–4999). They are wire contract: the
// client reads them from "mstatus", so never renumber or reuse one.
// Moved from domain/shared/status (docs/features/014-status-registry).
const (
	SchoolNotFound           status.StatusCode = 4901
	SchoolMissingID          status.StatusCode = 4902
	SchoolMissingName        status.StatusCode = 4903
	SchoolNameTooLong        status.StatusCode = 4904
	SchoolInvalidStatus      status.StatusCode = 4905
	SchoolDescriptionTooLong status.StatusCode = 4906
	SchoolProfileLinkFailed  status.StatusCode = 4907
)

func init() {
	status.Register(SchoolNotFound, "School not found", "Không tìm thấy trường học")
	status.Register(SchoolMissingID, "School id is required", "Mã trường học là bắt buộc")
	status.Register(SchoolMissingName, "School name is required", "Tên trường học là bắt buộc")
	status.Register(SchoolNameTooLong, "School name is too long", "Tên trường học quá dài")
	status.Register(SchoolInvalidStatus, "School status is invalid", "Trạng thái trường học không hợp lệ")
	status.Register(SchoolDescriptionTooLong, "School description is too long", "Mô tả trường học quá dài")
	status.Register(SchoolProfileLinkFailed, "Failed to link profile with school", "Liên kết hồ sơ với trường học thất bại")
}
