package semester

import (
	"encoding/json"
	"net/http"

	dto "math-ai.com/math-ai/internal/application/dto/semester"
	"math-ai.com/math-ai/internal/shared/response"
)

type SemesterHandler struct {
	service *Service
}

func NewSemesterHandler(service *Service) *SemesterHandler {
	return &SemesterHandler{service: service}
}

// POST /semesters/list
func (h *SemesterHandler) HandleListSemesters(w http.ResponseWriter, r *http.Request) {
	var req dto.ListSemestersReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	res, err := h.service.ListSemesters(r.Context(), &req)
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

	res, err := h.service.CreateSemester(r.Context(), &req)
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

	res, err := h.service.UpdateSemester(r.Context(), &req)
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

	res, err := h.service.SoftDeleteSemester(r.Context(), &req)
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

	res, err := h.service.ForceDeleteSemester(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	response.WriteJson(w, res, nil)
}

// POST /semesters/detail
func (h *SemesterHandler) HandleGetSemester(w http.ResponseWriter, r *http.Request) {
	var req dto.GetSemesterReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	res, err := h.service.GetSemester(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	response.WriteJson(w, res, nil)
}
