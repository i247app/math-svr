package grade

import (
	"encoding/json"
	"net/http"

	dto "math-ai.com/math-ai/internal/application/dto/grade"
	"math-ai.com/math-ai/internal/infrastructure/metadata"
	"math-ai.com/math-ai/internal/shared/response"
	"math-ai.com/math-ai/internal/shared/utils"
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

// POST /grades/create
func (h *GradeHandler) HandleCreateGrade(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateGradeReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	res, err := h.gradeSvc.CreateGrade(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// POST /grades/update
func (h *GradeHandler) HandleUpdateGrade(w http.ResponseWriter, r *http.Request) {
	var req dto.UpdateGradeReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	res, err := h.gradeSvc.UpdateGrade(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// POST /grades/soft-delete
func (h *GradeHandler) HandleSoftDeleteGrade(w http.ResponseWriter, r *http.Request) {
	var req dto.DeleteGradeReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	res, err := h.gradeSvc.SoftDeleteGrade(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// POST /grades/force-delete
func (h *GradeHandler) HandleForceDeleteGrade(w http.ResponseWriter, r *http.Request) {
	var req dto.DeleteGradeReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	res, err := h.gradeSvc.ForceDeleteGrade(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// GET /grades/{id}
func (h *GradeHandler) HandleGetGrade(w http.ResponseWriter, r *http.Request) {
	req := dto.GetGradeReq{
		GradeID:  utils.StringToInt64(r.PathValue("id"), 0),
		Language: metadata.GetClientLanguage(r.Context()).ToEnumLanguage(),
	}

	res, err := h.gradeSvc.GetGrade(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}
