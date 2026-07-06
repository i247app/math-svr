package bot

import (
	"encoding/json"
	"net/http"

	dto "math-ai.com/math-ai/internal/application/dto/bot"
	"math-ai.com/math-ai/internal/application/resource"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/response"
)

// Handler exposes the AI handshake / session-init over HTTP.
type Handler struct {
	appResource *resource.Resource
	svc         *Service
}

func NewHandler(appResource *resource.Resource, svc *Service) *Handler {
	return &Handler{appResource: appResource, svc: svc}
}

// HandleShake serves POST /ai/shake (authenticated).
//
// It does two things for the logged-in user: (1) warms the shared LLM
// connection (handshake + keep-alive priming, globally throttled), and
// (2) initializes that user's AI session — ensuring their CHAT conversation
// thread and returning its conversation_id. The client reuses that id on
// /ai/conversations/send so the server keeps per-user context across turns.
//
// Note: identity is server-side (the session). The LLM provider is stateless
// and does not recognise end users.
func (h *Handler) HandleShake(w http.ResponseWriter, r *http.Request) {
	var req dto.ShakeReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	session, err := h.appResource.GetRequestSession(r)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	uid, ok := session.UID()
	if !ok {
		response.WriteJson(w, nil, errs.NewError(r.Context(), status.UNAUTHORIZED, nil, nil))
		return
	}

	// ?force=true bypasses the warm-up cache so the call really hits the
	// vendor — useful for observing connection reuse across two calls.
	force := r.URL.Query().Get("force") == "true"

	req.UID = uid
	res := h.svc.Shake(r.Context(), req, force)
	response.WriteJson(w, res, nil)
}
