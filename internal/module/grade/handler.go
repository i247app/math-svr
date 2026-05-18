package grade

import (
	"encoding/json"
	"net/http"

	dto "math-ai.com/math-ai/internal/application/dto/grade"
	"math-ai.com/math-ai/internal/shared/response"
)

type GradeHandler struct {
	gradeSvc *Service
}

func NewGradeHandler(gradeSvc *Service) *GradeHandler {
	return &GradeHandler{gradeSvc: gradeSvc}
}

// POST /grades/list
func (h *GradeHandler) HandleListGrades(w http.ResponseWriter, r *http.Request) {
	var req dto.ListGradesReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	res, err := h.gradeSvc.ListGrades(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	response.WriteJson(w, res, nil)
}
