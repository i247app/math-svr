package llm

import "math-ai.com/math-ai/internal/domain/shared/status"

// Status codes of this aggregate (10000–10099). They are wire contract: the
// client reads them from "mstatus", so never renumber or reuse one.
// Moved from domain/shared/status (docs/features/014-status-registry).
const (
	LLMConnectFailed   status.StatusCode = 10001
	LLMConfigInvalid   status.StatusCode = 10002
	LLMOpFailed        status.StatusCode = 10003
	LLMSerializeFailed status.StatusCode = 10004
	LLMInvalidPrompt   status.StatusCode = 10005
	LLMContextTooLarge status.StatusCode = 10006
	LLMRateLimited     status.StatusCode = 10007
	LLMUnsupportedOp   status.StatusCode = 10008
	// LLMContentBlocked is the provider's safety / recitation filter
	// refusing to answer. Distinct from LLM_OP_FAILED on purpose: that one
	// reads as transient and invites a retry, whereas a blocked prompt
	// returns the same block every time and needs the prompt changed.
	// Gemini surfaces this as finishReason SAFETY / RECITATION /
	// PROHIBITED_CONTENT or promptFeedback.blockReason.
	LLMContentBlocked status.StatusCode = 10009
)

func init() {
	status.Register(LLMConnectFailed, "AI provider connection failed", "Kết nối nhà cung cấp AI thất bại")
	status.Register(LLMConfigInvalid, "AI configuration is invalid", "Cấu hình AI không hợp lệ")
	status.Register(LLMOpFailed, "AI operation failed", "Thao tác AI thất bại")
	status.Register(LLMSerializeFailed, "Failed to decode AI response", "Không thể giải mã phản hồi AI")
	status.Register(LLMInvalidPrompt, "AI prompt is invalid", "Nội dung yêu cầu AI không hợp lệ")
	status.Register(LLMContextTooLarge, "AI context window exceeded", "Vượt quá giới hạn ngữ cảnh của mô hình AI")
	status.Register(LLMRateLimited, "AI provider rate limit exceeded", "Đã vượt quá giới hạn tần suất gọi AI")
	status.Register(LLMUnsupportedOp, "AI operation is not supported by the configured provider", "Thao tác AI không được nhà cung cấp hiện tại hỗ trợ")
	status.Register(LLMContentBlocked, "The AI provider blocked this content", "Nội dung đã bị nhà cung cấp AI chặn")
}
