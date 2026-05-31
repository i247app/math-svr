package semester

import (
	"encoding/json"
	"net/http"

	dto "math-ai.com/math-ai/internal/application/dto/semester"
	"math-ai.com/math-ai/internal/infrastructure/metadata"
	"math-ai.com/math-ai/internal/shared/response"
	"math-ai.com/math-ai/internal/shared/utils"
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

// POST /semesters/create
func (h *SemesterHandler) HandleCreateSemester(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateSemesterReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	res, err := h.semesterSvc.CreateSemester(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// POST /semesters/update
func (h *SemesterHandler) HandleUpdateSemester(w http.ResponseWriter, r *http.Request) {
	var req dto.UpdateSemesterReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	res, err := h.semesterSvc.UpdateSemester(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// POST /semesters/soft-delete
func (h *SemesterHandler) HandleSoftDeleteSemester(w http.ResponseWriter, r *http.Request) {
	var req dto.DeleteSemesterReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	res, err := h.semesterSvc.SoftDeleteSemester(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// POST /semesters/force-delete
func (h *SemesterHandler) HandleForceDeleteSemester(w http.ResponseWriter, r *http.Request) {
	var req dto.DeleteSemesterReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	res, err := h.semesterSvc.ForceDeleteSemester(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// GET /semesters/{id}
func (h *SemesterHandler) HandleGetSemester(w http.ResponseWriter, r *http.Request) {
	req := dto.GetSemesterReq{
		SemesterID: utils.StringToInt64(r.PathValue("id"), 0),
		Language:   metadata.GetClientLanguage(r.Context()).ToEnumLanguage(),
	}

	res, err := h.semesterSvc.GetSemester(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}
