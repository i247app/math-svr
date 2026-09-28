package program

import (
	"encoding/json"
	"net/http"

	dto "math-ai.com/math-ai/internal/application/dto/program"
	"math-ai.com/math-ai/internal/shared/response"
)

type ProgramHandler struct {
	service *Service
}

func NewProgramHandler(service *Service) *ProgramHandler {
	return &ProgramHandler{service: service}
}

// POST /programs/list
func (h *ProgramHandler) HandleListPrograms(w http.ResponseWriter, r *http.Request) {
	var req dto.ListProgramsReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	res, err := h.service.ListPrograms(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	response.WriteJson(w, res, nil)
}

// POST /programs/create
func (h *ProgramHandler) HandleCreateProgram(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateProgramReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	res, err := h.service.CreateProgram(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	response.WriteJson(w, res, nil)
}

// POST /programs/update
func (h *ProgramHandler) HandleUpdateProgram(w http.ResponseWriter, r *http.Request) {
	var req dto.UpdateProgramReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	res, err := h.service.UpdateProgram(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	response.WriteJson(w, res, nil)
}

// POST /programs/soft-delete
func (h *ProgramHandler) HandleSoftDeleteProgram(w http.ResponseWriter, r *http.Request) {
	var req dto.DeleteProgramReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	res, err := h.service.SoftDeleteProgram(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	response.WriteJson(w, res, nil)
}

// POST /programs/force-delete
func (h *ProgramHandler) HandleForceDeleteProgram(w http.ResponseWriter, r *http.Request) {
	var req dto.DeleteProgramReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	res, err := h.service.ForceDeleteProgram(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	response.WriteJson(w, res, nil)
}

// POST /programs/detail
func (h *ProgramHandler) HandleGetProgram(w http.ResponseWriter, r *http.Request) {
	var req dto.GetProgramReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	res, err := h.service.GetProgram(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	response.WriteJson(w, res, nil)
}
