package semester

import (
	"encoding/json"
	"net/http"

	dto "math-ai.com/math-ai/internal/application/dto/semester"
	"math-ai.com/math-ai/internal/shared/response"
)

type SemesterHandler struct {
	semesterSvc *Service
}

func NewSemesterHandler(semesterSvc *Service) *SemesterHandler {
	return &SemesterHandler{semesterSvc: semesterSvc}
}

// POST /semesters/list
func (h *SemesterHandler) HandleListSemesters(w http.ResponseWriter, r *http.Request) {
	var req dto.ListSemestersReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	res, err := h.semesterSvc.ListSemesters(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	response.WriteJson(w, res, nil)
}
