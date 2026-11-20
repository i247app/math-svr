package chat

import "math-ai.com/math-ai/internal/domain/shared/status"

// Status codes of this aggregate (13500–13599). They are wire contract: the
// client reads them from "mstatus", so never renumber or reuse one.
// Moved from domain/shared/status (docs/features/014-status-registry).
const (
	// Chat (realtime messaging) — 13500-13599
	ChatConversationNotFound     status.StatusCode = 13500
	ChatNotParticipant           status.StatusCode = 13501
	ChatMessageNotFound          status.StatusCode = 13502
	ChatMessageEmpty             status.StatusCode = 13503
	ChatMessageTooLong           status.StatusCode = 13504
	ChatTargetNotInClassroom     status.StatusCode = 13505
	ChatCannotMessageSelf        status.StatusCode = 13506
	ChatSeqAllocationFailed      status.StatusCode = 13507
	ChatMissingConversationID    status.StatusCode = 13508
	ChatMissingClassroomID       status.StatusCode = 13509
	ChatMissingTargetProfileID   status.StatusCode = 13510
	ChatConversationCreateFailed status.StatusCode = 13511
	ChatMessageCreateFailed      status.StatusCode = 13512
	ChatInvalidMessageType       status.StatusCode = 13513
	ChatMissingProfileID         status.StatusCode = 13514
)

func init() {
	status.Register(ChatConversationNotFound, "Conversation not found", "Không tìm thấy cuộc trò chuyện")
	status.Register(ChatNotParticipant, "You are not a participant of this conversation", "Bạn không thuộc cuộc trò chuyện này")
	status.Register(ChatMessageNotFound, "Message not found", "Không tìm thấy tin nhắn")
	status.Register(ChatMessageEmpty, "Message content is required", "Nội dung tin nhắn là bắt buộc")
	status.Register(ChatMessageTooLong, "Message content is too long", "Nội dung tin nhắn quá dài")
	status.Register(ChatTargetNotInClassroom, "This member does not belong to the classroom", "Thành viên này không thuộc lớp học")
	status.Register(ChatCannotMessageSelf, "You cannot message yourself", "Bạn không thể nhắn tin cho chính mình")
	status.Register(ChatSeqAllocationFailed, "Failed to allocate a message sequence number", "Không cấp phát được số thứ tự tin nhắn")
	status.Register(ChatMissingConversationID, "conversation_id is required", "Thiếu conversation_id")
	status.Register(ChatMissingClassroomID, "classroom_id is required", "Thiếu classroom_id")
	status.Register(ChatMissingTargetProfileID, "target_profile_id is required", "Thiếu target_profile_id")
	status.Register(ChatConversationCreateFailed, "Failed to create the conversation", "Tạo cuộc trò chuyện thất bại")
	status.Register(ChatMessageCreateFailed, "Failed to send the message", "Gửi tin nhắn thất bại")
	status.Register(ChatInvalidMessageType, "Invalid message type", "Loại tin nhắn không hợp lệ")
	status.Register(ChatMissingProfileID, "profile_id is required", "Thiếu profile_id")
}
