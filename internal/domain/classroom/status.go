package classroom

import "math-ai.com/math-ai/internal/domain/shared/status"

// Status codes of this aggregate (12200–12599). They are wire contract: the
// client reads them from "mstatus", so never renumber or reuse one.
// Moved from domain/shared/status (docs/features/014-status-registry).
const (
	ClassroomNotFound                 status.StatusCode = 12201
	ClassroomMissingID                status.StatusCode = 12202
	ClassroomMissingName              status.StatusCode = 12203
	ClassroomNameTooLong              status.StatusCode = 12204
	ClassroomDescriptionTooLong       status.StatusCode = 12205
	ClassroomMissingOwnerProfileID    status.StatusCode = 12206
	ClassroomInvalidOwnerRole         status.StatusCode = 12207
	ClassroomInvalidProgram           status.StatusCode = 12208
	ClassroomInvalidGrade             status.StatusCode = 12209
	ClassroomAlreadyArchived          status.StatusCode = 12210
	ClassroomNotArchived              status.StatusCode = 12211
	ClassroomAlreadyDeleted           status.StatusCode = 12212
	ClassroomMaxMembersReached        status.StatusCode = 12213
	ClassroomCodeInvalid              status.StatusCode = 12214
	ClassroomCodeExpired              status.StatusCode = 12215
	ClassroomCodeDisabled             status.StatusCode = 12216
	ClassroomCodeGenerationFailed     status.StatusCode = 12217
	ClassroomPermissionDenied         status.StatusCode = 12218
	ClassroomOwnerCannotLeave         status.StatusCode = 12219
	ClassroomOwnerTransferToNonMember status.StatusCode = 12220
	ClassroomInvalidMaxMembers        status.StatusCode = 12221
	ClassroomInvalidSchool            status.StatusCode = 12222
	ClassroomCodeTaken                status.StatusCode = 12223
	ClassroomProgramDuplicate         status.StatusCode = 12224
	// Profile-learning-progress (002-profile-learning-progress).
	ClassroomProgressInvalidDateRange        status.StatusCode = 12225
	ClassroomProgressInvalidTz               status.StatusCode = 12226
	ClassroomProgressInvalidPurpose          status.StatusCode = 12227
	ClassroomMemberNotFound                  status.StatusCode = 12301
	ClassroomMemberMissingID                 status.StatusCode = 12302
	ClassroomMemberMissingClassroomID        status.StatusCode = 12303
	ClassroomMemberMissingProfileID          status.StatusCode = 12304
	ClassroomMemberAlreadyMember             status.StatusCode = 12305
	ClassroomMemberNotMember                 status.StatusCode = 12306
	ClassroomMemberInvalidRole               status.StatusCode = 12307
	ClassroomMemberCannotRemoveOwner         status.StatusCode = 12308
	ClassroomMemberCannotDemoteOwner         status.StatusCode = 12309
	ClassroomInvitationNotFound              status.StatusCode = 12401
	ClassroomInvitationMissingID             status.StatusCode = 12402
	ClassroomInvitationMissingTarget         status.StatusCode = 12403
	ClassroomInvitationInvalidIdentifier     status.StatusCode = 12404
	ClassroomInvitationInvalidIdentifierType status.StatusCode = 12405
	ClassroomInvitationInvalidRole           status.StatusCode = 12406
	ClassroomInvitationAlreadyResponded      status.StatusCode = 12407
	ClassroomInvitationExpired               status.StatusCode = 12408
	ClassroomInvitationNotPending            status.StatusCode = 12409
	ClassroomInvitationNotInvitee            status.StatusCode = 12410
	ClassroomInvitationAlreadyInvited        status.StatusCode = 12411
	ClassroomInvitationTokenInvalid          status.StatusCode = 12412
	ClassroomInvitationTokenGenerationFailed status.StatusCode = 12413
	ClassroomInvitationPermissionDenied      status.StatusCode = 12414
	// Classroom join request — user-initiated counterpart to the
	// teacher-initiated invitation flow. Same ma_classroom_members
	// table, distinguished by member_status = PENDING_REQUEST.
	ClassroomJoinRequestNotFound         status.StatusCode = 12501
	ClassroomJoinRequestNotPending       status.StatusCode = 12502
	ClassroomJoinRequestAlreadyPending   status.StatusCode = 12503
	ClassroomJoinRequestPermissionDenied status.StatusCode = 12504
)

func init() {
	status.Register(ClassroomNotFound, "Classroom not found", "Không tìm thấy lớp học")
	status.Register(ClassroomMissingID, "Classroom id is required", "Mã lớp học là bắt buộc")
	status.Register(ClassroomMissingName, "Classroom name is required", "Tên lớp học là bắt buộc")
	status.Register(ClassroomNameTooLong, "Classroom name is too long", "Tên lớp học quá dài")
	status.Register(ClassroomDescriptionTooLong, "Classroom description is too long", "Mô tả lớp học quá dài")
	status.Register(ClassroomMissingOwnerProfileID, "Owner profile id is required", "Hồ sơ chủ lớp là bắt buộc")
	status.Register(ClassroomInvalidOwnerRole, "Only a teacher profile can own a classroom", "Chỉ hồ sơ giáo viên mới có thể làm chủ lớp học")
	status.Register(ClassroomInvalidProgram, "Program is invalid", "Chương trình không hợp lệ")
	status.Register(ClassroomInvalidGrade, "Grade is invalid", "Khối lớp không hợp lệ")
	status.Register(ClassroomAlreadyArchived, "Classroom is already archived", "Lớp học đã được lưu trữ")
	status.Register(ClassroomNotArchived, "Classroom is not archived", "Lớp học chưa được lưu trữ")
	status.Register(ClassroomAlreadyDeleted, "Classroom is already deleted", "Lớp học đã bị xoá")
	status.Register(ClassroomMaxMembersReached, "Classroom has reached its member limit", "Lớp học đã đạt giới hạn thành viên")
	status.Register(ClassroomCodeInvalid, "Invite code is invalid", "Mã mời không hợp lệ")
	status.Register(ClassroomCodeExpired, "Invite code has expired", "Mã mời đã hết hạn")
	status.Register(ClassroomCodeDisabled, "Invite code is disabled", "Mã mời đã bị vô hiệu hoá")
	status.Register(ClassroomCodeGenerationFailed, "Failed to generate invite code", "Tạo mã mời thất bại")
	status.Register(ClassroomPermissionDenied, "You do not have permission to perform this action on the classroom", "Bạn không có quyền thực hiện thao tác này trên lớp học")
	status.Register(ClassroomOwnerCannotLeave, "Owner must transfer ownership before leaving the classroom", "Chủ lớp phải chuyển quyền sở hữu trước khi rời lớp")
	status.Register(ClassroomOwnerTransferToNonMember, "Ownership can only be transferred to an active member", "Chỉ có thể chuyển quyền sở hữu cho thành viên đang hoạt động")
	status.Register(ClassroomInvalidMaxMembers, "Max members must be a positive number", "Giới hạn thành viên phải là số dương")
	status.Register(ClassroomInvalidSchool, "School not found", "Không tìm thấy trường học")
	status.Register(ClassroomCodeTaken, "Invite code is already in use", "Mã mời đã được sử dụng")
	status.Register(ClassroomProgramDuplicate, "Duplicate program id in the list", "Danh sách chương trình bị trùng")
	status.Register(ClassroomProgressInvalidDateRange, "Invalid date range.", "Khoảng thời gian không hợp lệ.")
	status.Register(ClassroomProgressInvalidTz, "Invalid time zone.", "Múi giờ không hợp lệ.")
	status.Register(ClassroomProgressInvalidPurpose, "Invalid exercise type.", "Loại bài không hợp lệ.")
	status.Register(ClassroomMemberNotFound, "Classroom member not found", "Không tìm thấy thành viên lớp học")
	status.Register(ClassroomMemberMissingID, "Member id is required", "Mã thành viên là bắt buộc")
	status.Register(ClassroomMemberMissingClassroomID, "Classroom id is required", "Mã lớp học là bắt buộc")
	status.Register(ClassroomMemberMissingProfileID, "Profile id is required", "Mã hồ sơ là bắt buộc")
	status.Register(ClassroomMemberAlreadyMember, "Profile is already a member of this classroom", "Hồ sơ đã là thành viên của lớp học")
	status.Register(ClassroomMemberNotMember, "Profile is not a member of this classroom", "Hồ sơ không phải là thành viên của lớp học")
	status.Register(ClassroomMemberInvalidRole, "Member role is invalid", "Vai trò thành viên không hợp lệ")
	status.Register(ClassroomMemberCannotRemoveOwner, "Owner cannot be removed", "Không thể xoá chủ lớp")
	status.Register(ClassroomMemberCannotDemoteOwner, "Owner role cannot be changed; transfer ownership first", "Không thể đổi vai trò chủ lớp; hãy chuyển quyền sở hữu trước")
	status.Register(ClassroomInvitationNotFound, "Invitation not found", "Không tìm thấy lời mời")
	status.Register(ClassroomInvitationMissingID, "Invitation id is required", "Mã lời mời là bắt buộc")
	status.Register(ClassroomInvitationMissingTarget, "Invitation must target a profile or an identifier", "Lời mời phải có hồ sơ hoặc định danh người nhận")
	status.Register(ClassroomInvitationInvalidIdentifier, "Invitation identifier is invalid", "Định danh người nhận không hợp lệ")
	status.Register(ClassroomInvitationInvalidIdentifierType, "Invitation identifier type is invalid", "Loại định danh người nhận không hợp lệ")
	status.Register(ClassroomInvitationInvalidRole, "Proposed role is invalid", "Vai trò đề xuất không hợp lệ")
	status.Register(ClassroomInvitationAlreadyResponded, "Invitation has already been responded to", "Lời mời đã được phản hồi")
	status.Register(ClassroomInvitationExpired, "Invitation has expired", "Lời mời đã hết hạn")
	status.Register(ClassroomInvitationNotPending, "Invitation is not in a pending state", "Lời mời không ở trạng thái chờ")
	status.Register(ClassroomInvitationNotInvitee, "You are not the invitee of this invitation", "Bạn không phải người được mời")
	status.Register(ClassroomInvitationAlreadyInvited, "A pending invitation already exists for this invitee", "Đã tồn tại lời mời đang chờ cho người này")
	status.Register(ClassroomInvitationTokenInvalid, "Invitation token is invalid", "Token lời mời không hợp lệ")
	status.Register(ClassroomInvitationTokenGenerationFailed, "Failed to generate invitation token", "Tạo token lời mời thất bại")
	status.Register(ClassroomInvitationPermissionDenied, "You do not have permission to manage this invitation", "Bạn không có quyền thao tác trên lời mời này")
	status.Register(ClassroomJoinRequestNotFound, "Join request not found", "Không tìm thấy yêu cầu tham gia")
	status.Register(ClassroomJoinRequestNotPending, "Join request is not in a pending state", "Yêu cầu tham gia không ở trạng thái chờ duyệt")
	status.Register(ClassroomJoinRequestAlreadyPending, "A pending join request already exists for this classroom", "Đã tồn tại yêu cầu tham gia đang chờ duyệt cho lớp học này")
	status.Register(ClassroomJoinRequestPermissionDenied, "You do not have permission to manage this join request", "Bạn không có quyền thao tác trên yêu cầu tham gia này")
}
