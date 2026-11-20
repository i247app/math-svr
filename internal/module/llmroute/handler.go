package llmroute

import (
	"encoding/json"
	"net/http"

	"math-ai.com/math-ai/internal/application/dto"
	"math-ai.com/math-ai/internal/shared/response"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// POST /llm-routes/list — admin only; every route and what it resolves to.
func (h *Handler) HandleListLLMRoutes(w http.ResponseWriter, r *http.Request) {
	var req dto.ListLLMRoutesReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	res, err := h.service.ListLLMRoutes(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// POST /llm-routes/update — admin only; retargets one route.
func (h *Handler) HandleUpdateLLMRoute(w http.ResponseWriter, r *http.Request) {
	var req dto.UpdateLLMRouteReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	res, err := h.service.UpdateLLMRoute(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// POST /llm-routes/models — admin only; the models each registered provider
// lists, for picking llm_model. Body {provider?}.
func (h *Handler) HandleListLLMModels(w http.ResponseWriter, r *http.Request) {
	var req dto.ListLLMModelsReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	res, err := h.service.ListLLMModels(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}
