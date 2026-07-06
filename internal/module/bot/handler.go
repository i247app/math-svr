package bot

import (
	"encoding/json"
	"net/http"

	dto "math-ai.com/math-ai/internal/application/dto/bot"
	"math-ai.com/math-ai/internal/shared/response"
)

// Handler exposes the AI connection warm-up over HTTP.
type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// HandleShake serves POST /ai/shake.
//
// Best-effort priming of the LLM connection pool so the first real
// quiz/exercise call is fast. Safe to call repeatedly — a process-global TTL
// + single-flight coalesce calls. Always returns a Success envelope; check
// the body's `shaked` flag.
//
// ?force=true bypasses the warm-up cache so the call really hits the vendor —
// useful for observing connection reuse (see the langchain.conn log line).
func (h *Handler) HandleShake(w http.ResponseWriter, r *http.Request) {
	var req dto.ShakeReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	force := r.URL.Query().Get("force") == "true"
	res := h.svc.Shake(r.Context(), req, force)
	response.WriteJson(w, res, nil)
}
