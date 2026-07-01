package misc

import (
	"encoding/json"
	"net/http"

	dto "math-ai.com/math-ai/internal/application/dto/misc"
	"math-ai.com/math-ai/internal/shared/response"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{
		svc: svc,
	}
}

func (h *Handler) LogsTimeFormat(w http.ResponseWriter, r *http.Request) {
	var req dto.LogTimeFormatReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	res, err := h.svc.LogsTimeFormat(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	response.WriteJson(w, res, nil)
}
