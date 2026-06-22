package home

import (
	"encoding/json"
	"fmt"
	"net/http"

	dto "math-ai.com/math-ai/internal/application/dto/home"
	"math-ai.com/math-ai/internal/application/resource"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/response"
)

// Handler is the home module's HTTP surface — a single endpoint that
// renders the role-aware home dashboard for one of the session user's
// profiles.
type Handler struct {
	appResource *resource.Resource
	homeSvc     *Service
}

func NewHandler(appResource *resource.Resource, homeSvc *Service) *Handler {
	return &Handler{appResource: appResource, homeSvc: homeSvc}
}

// sessionUID pulls the authenticated user's id off the session so the
// service can confirm the requested profile_id belongs to that user.
func (h *Handler) sessionUID(r *http.Request) (int64, error) {
	sess, err := h.appResource.GetRequestSession(r)
	if err != nil {
		return 0, err
	}
	uid, ok := sess.UID()
	if !ok {
		return 0, errs.NewError(r.Context(), status.UNAUTHORIZED, nil, ErrUIDNotFoundFromSession)
	}
	return uid, nil
}

// POST /home/layout
func (h *Handler) HandleGetHomeLayout(w http.ResponseWriter, r *http.Request) {
	var req dto.HomeLayoutReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, fmt.Errorf("invalid parameters"))
		return
	}

	uid, err := h.sessionUID(r)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	res, err := h.homeSvc.GetHomeLayout(r.Context(), &req, uid)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	response.WriteJson(w, res, nil)
}
