package bot

import (
	"net/http"

	"math-ai.com/math-ai/internal/shared/response"
)

// Handler exposes the AI warm-up over HTTP.
type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// HandleWarmup serves POST /connect/init-ai.
//
// It is a best-effort priming of the LLM connection pool: the frontend
// calls it early (e.g. on entering the home screen) so the cold-connection
// latency is paid here instead of on the first real quiz/exercise call.
//
// Safe to call repeatedly and from any screen — a process-global TTL +
// single-flight in the service coalesce calls so the upstream LLM is hit at
// most once per warm-up window regardless of caller volume. The response is
// always a Success envelope; check the body's `warmed` flag.
func (h *Handler) HandleWarmup(w http.ResponseWriter, r *http.Request) {
	res := h.svc.Warmup(r.Context())
	response.WriteJson(w, res, nil)
}
