package profile

import "math-ai.com/math-ai/internal/domain/shared/status"

// Status codes of this aggregate (4100–4199). They are wire contract: the
// client reads them from "mstatus", so never renumber or reuse one.
// Moved from domain/shared/status (docs/features/014-status-registry).
const (
	ProfileNotFound               status.StatusCode = 4101
	ProfileMissingName            status.StatusCode = 4102
	ProfileMissingUserID          status.StatusCode = 4103
	ProfileMissingGradeID         status.StatusCode = 4104
	ProfileMissingSemesterID      status.StatusCode = 4105
	ProfileInvalidDob             status.StatusCode = 4106
	ProfileAvatarInvalidFile      status.StatusCode = 4107
	ProfileAvatarUploadFailed     status.StatusCode = 4108
	ProfileMissingProgramID       status.StatusCode = 4109
	ProfileInvalidLanguage        status.StatusCode = 4110
	ProfileAvatarConflict         status.StatusCode = 4111
	ProfileAvatarInvalidReference status.StatusCode = 4112
	ProfileCodeTaken              status.StatusCode = 4113
	ProfileCodeGenerationFailed   status.StatusCode = 4114
)

func init() {
	status.Register(ProfileNotFound, "Profile not found", "Không tìm thấy hồ sơ")
	status.Register(ProfileMissingName, "Please enter the child's name", "Vui lòng nhập tên của con")
	status.Register(ProfileMissingUserID, "User id is required", "Mã người dùng là bắt buộc")
	status.Register(ProfileMissingGradeID, "Grade id is required", "Mã lớp là bắt buộc")
	status.Register(ProfileMissingSemesterID, "Semester id is required", "Mã học kỳ là bắt buộc")
	status.Register(ProfileInvalidDob, "Date of birth is invalid", "Ngày sinh không hợp lệ")
	status.Register(ProfileAvatarInvalidFile, "Avatar file is invalid", "Tệp ảnh đại diện không hợp lệ")
	status.Register(ProfileAvatarUploadFailed, "Avatar upload failed", "Tải ảnh đại diện thất bại")
	status.Register(ProfileMissingProgramID, "Program id is required", "Mã chương trình học là bắt buộc")
	status.Register(ProfileInvalidLanguage, "Language is invalid", "Ngôn ngữ không hợp lệ")
	status.Register(ProfileAvatarConflict, "Provide either an avatar file or an avatar reference, not both", "Chỉ chọn một trong hai: tệp ảnh hoặc đường dẫn ảnh")
	status.Register(ProfileAvatarInvalidReference, "Avatar reference is invalid", "Đường dẫn ảnh đại diện không hợp lệ")
	status.Register(ProfileCodeTaken, "Profile code is already taken", "Mã hồ sơ đã được sử dụng")
	status.Register(ProfileCodeGenerationFailed, "Could not generate a unique profile code", "Không thể tạo mã hồ sơ duy nhất")
}
