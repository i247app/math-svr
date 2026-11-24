package game

import (
	"encoding/json"
	"net/http"

	"math-ai.com/math-ai/internal/application/dto"
	"math-ai.com/math-ai/internal/application/resource"
	"math-ai.com/math-ai/internal/shared/response"
)

type Handler struct {
	appResource *resource.Resource
	service     *Service
}

func NewHandler(appResource *resource.Resource, service *Service) *Handler {
	return &Handler{appResource: appResource, service: service}
}

// actorID is the signed-in user's id, nil when the session holds none (the
// routes are auth-gated; the service then refuses).
func (h *Handler) actorID(r *http.Request) *int64 {
	sess, err := h.appResource.GetRequestSession(r)
	if err != nil || sess == nil {
		return nil
	}
	uid, ok := sess.UID()
	if !ok {
		return nil
	}
	return &uid
}

// POST /games/start
func (h *Handler) HandleStartGame(w http.ResponseWriter, r *http.Request) {
	var req dto.StartGameReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	req.UID = h.actorID(r)

	res, err := h.service.StartGame(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// POST /games/submit
func (h *Handler) HandleSubmitGame(w http.ResponseWriter, r *http.Request) {
	var req dto.SubmitGameReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	req.UID = h.actorID(r)

	res, err := h.service.SubmitGame(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// POST /games/list
func (h *Handler) HandleListGames(w http.ResponseWriter, r *http.Request) {
	var req dto.ListGamesReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	req.UID = h.actorID(r)

	res, err := h.service.ListGames(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// POST /games/detail
func (h *Handler) HandleGetGame(w http.ResponseWriter, r *http.Request) {
	var req dto.GetGameReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	req.UID = h.actorID(r)

	res, err := h.service.GetGame(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}
