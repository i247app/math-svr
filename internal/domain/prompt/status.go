package prompt

import "math-ai.com/math-ai/internal/domain/shared/status"

// Status codes of this aggregate (14300–14399). They are wire contract: the
// client reads them from "mstatus", so never renumber or reuse one.
// Moved from domain/shared/status (docs/features/014-status-registry).
const (
	// Prompts — 14300-14399. The static half of every LLM call (ma_prompts,
	// /prompts/*).
	// PromptNotFound: no live row for the scope an LLM call needs, or for
	// the prompt_id a /prompts route names. Generation fails rather than
	// falling back to a prompt in code.
	PromptNotFound status.StatusCode = 14300
	// PromptMissingID: /prompts/detail or /prompts/update named no prompt_id.
	PromptMissingID status.StatusCode = 14301
	// PromptEmpty: /prompts/update sent no text (or only whitespace).
	PromptEmpty status.StatusCode = 14302
	// PromptTooLong: the text is over the 50,000-character limit.
	PromptTooLong status.StatusCode = 14303
	// PromptMissingJSONKey: the text does not name every JSON key the
	// reply parser of that prompt_code reads; debug lists the missing ones.
	PromptMissingJSONKey status.StatusCode = 14304
)

func init() {
	status.Register(PromptNotFound, "No prompt is configured for this request", "Chưa cấu hình prompt cho yêu cầu này")
	status.Register(PromptMissingID, "prompt_id is required", "Thiếu prompt_id")
	status.Register(PromptEmpty, "The prompt cannot be empty", "Prompt không được để trống")
	status.Register(PromptTooLong, "The prompt is too long (at most 50,000 characters)", "Prompt quá dài (tối đa 50.000 ký tự)")
	status.Register(PromptMissingJSONKey, "The prompt must name every JSON key of its output structure", "Prompt phải nêu đủ các khóa JSON của cấu trúc kết quả")
}
