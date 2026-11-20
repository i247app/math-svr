package seq

import "math-ai.com/math-ai/internal/domain/shared/status"

// Status codes of this aggregate (12100–12199). They are wire contract: the
// client reads them from "mstatus", so never renumber or reuse one.
// Moved from domain/shared/status (docs/features/014-status-registry).
const (
	SeqNotFound         status.StatusCode = 12101
	SeqMissingName      status.StatusCode = 12102
	SeqGenerationFailed status.StatusCode = 12103
)

func init() {
	status.Register(SeqNotFound, "Sequence not found", "Không tìm thấy sequence")
	status.Register(SeqMissingName, "Sequence name is required", "Tên sequence là bắt buộc")
	status.Register(SeqGenerationFailed, "Failed to generate next id", "Sinh id mới thất bại")
}
