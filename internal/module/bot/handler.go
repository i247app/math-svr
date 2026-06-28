package bot

import (
	"net/http"

	"math-ai.com/math-ai/internal/shared/response"
)

// Handler exposes the AI initial handhshake over HTTP.
type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// HandleShake serves POST /ai/shake.
//
// It is a best-effort priming of the LLM connection pool: the frontend
// calls it early (e.g. on entering the home screen) so the cold-connection
// latency is paid here instead of on the first real quiz/exercise call.
//
// Safe to call repeatedly and from any screen — a process-global TTL +
// single-flight in the service coalesce calls so the upstream LLM is hit at
// most once per handshake window regardless of caller volume. The response is
// always a Success envelope; check the body's `shaked` flag.
func (h *Handler) HandleShake(w http.ResponseWriter, r *http.Request) {
	res := h.svc.Shake(r.Context())
	response.WriteJson(w, res, nil)
}
