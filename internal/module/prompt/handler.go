package prompt

import (
	"encoding/json"
	"net/http"

	dto "math-ai.com/math-ai/internal/application/dto/prompt"
	"math-ai.com/math-ai/internal/shared/response"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// POST /prompts/list — admin only; every prompt, full text included.
func (h *Handler) HandleListPrompts(w http.ResponseWriter, r *http.Request) {
	var req dto.ListPromptsReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	res, err := h.service.ListPrompts(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// POST /prompts/detail — admin only; one prompt by prompt_id.
func (h *Handler) HandleGetPrompt(w http.ResponseWriter, r *http.Request) {
	var req dto.GetPromptReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	res, err := h.service.GetPrompt(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// POST /prompts/update — admin only; overwrites one prompt's text.
func (h *Handler) HandleUpdatePrompt(w http.ResponseWriter, r *http.Request) {
	var req dto.UpdatePromptReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	res, err := h.service.UpdatePrompt(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}
