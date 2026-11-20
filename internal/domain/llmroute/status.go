package llmroute

import "math-ai.com/math-ai/internal/domain/shared/status"

// Status codes of this aggregate (14600–14699). They are wire contract: the
// client reads them from "mstatus", so never renumber or reuse one.
// Moved from domain/shared/status (docs/features/014-status-registry).
const (
	// LLMRouteNotFound: no live row has that route_code.
	LLMRouteNotFound status.StatusCode = 14600
	// LLMRouteMissingCode: route_code is empty.
	LLMRouteMissingCode status.StatusCode = 14601
	// LLMRouteInvalidProvider: llm_provider is not a provider registered
	// on the server (unknown name, or its API key is not configured).
	LLMRouteInvalidProvider status.StatusCode = 14602
	// LLMRouteInvalidModel: llm_model is too long, or set without an
	// llm_provider.
	LLMRouteInvalidModel status.StatusCode = 14603
)

func init() {
	status.Register(LLMRouteNotFound, "LLM route not found", "Không tìm thấy cấu hình định tuyến LLM")
	status.Register(LLMRouteMissingCode, "LLM route code is required", "Mã định tuyến LLM là bắt buộc")
	status.Register(LLMRouteInvalidProvider, "The LLM provider is not configured on the server", "Nhà cung cấp LLM chưa được cấu hình trên máy chủ")
	status.Register(LLMRouteInvalidModel, "The LLM model is invalid", "Model LLM không hợp lệ")
}
