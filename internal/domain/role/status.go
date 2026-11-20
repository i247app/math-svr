package role

import "math-ai.com/math-ai/internal/domain/shared/status"

// Status codes of this aggregate (13900–13999). They are wire contract: the
// client reads them from "mstatus", so never renumber or reuse one.
// Moved from domain/shared/status (docs/features/014-status-registry).
const (
	// Role (permission module) — 13900-13999
	RoleNotFound           status.StatusCode = 13900
	RoleMissingID          status.StatusCode = 13901
	RoleMissingCode        status.StatusCode = 13902
	RoleInvalidCode        status.StatusCode = 13903
	RoleCodeAlreadyExists  status.StatusCode = 13904
	RoleMissingName        status.StatusCode = 13905
	RoleNameTooLong        status.StatusCode = 13906
	RoleDescriptionTooLong status.StatusCode = 13907
	RoleNoteTooLong        status.StatusCode = 13908
	RoleInvalidStatus      status.StatusCode = 13909
	RoleImageInvalidFile   status.StatusCode = 13910
	RoleImageUploadFailed  status.StatusCode = 13911
	RoleImageConflict      status.StatusCode = 13912 // update sent a new role_image AND remove_role_image
)

func init() {
	status.Register(RoleNotFound, "Role not found", "Không tìm thấy vai trò")
	status.Register(RoleMissingID, "role_id is required", "Thiếu role_id")
	status.Register(RoleMissingCode, "Role code is required", "Mã vai trò là bắt buộc")
	status.Register(RoleInvalidCode, "Role code must start with a letter and contain only upper-case letters, digits and underscores (max 64 characters)", "Mã vai trò chỉ gồm chữ in hoa, chữ số và dấu gạch dưới, bắt đầu bằng chữ cái, tối đa 64 ký tự")
	status.Register(RoleCodeAlreadyExists, "Role code already exists", "Mã vai trò đã tồn tại")
	status.Register(RoleMissingName, "Role name is required", "Tên vai trò là bắt buộc")
	status.Register(RoleNameTooLong, "Role name is too long", "Tên vai trò quá dài")
	status.Register(RoleDescriptionTooLong, "Role description is too long", "Mô tả vai trò quá dài")
	status.Register(RoleNoteTooLong, "Note is too long", "Ghi chú quá dài")
	status.Register(RoleInvalidStatus, "Invalid role status", "Trạng thái vai trò không hợp lệ")
	status.Register(RoleImageInvalidFile, "Invalid role image file", "Tệp ảnh vai trò không hợp lệ")
	status.Register(RoleImageUploadFailed, "Failed to upload role image", "Tải ảnh vai trò lên thất bại")
	status.Register(RoleImageConflict, "Send either a new role image or remove_role_image, not both", "Không thể vừa tải ảnh mới vừa xóa ảnh vai trò")
}
