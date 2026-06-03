package program

import (
	"encoding/json"
	"net/http"

	dto "math-ai.com/math-ai/internal/application/dto/program"
	"math-ai.com/math-ai/internal/infrastructure/metadata"
	"math-ai.com/math-ai/internal/shared/response"
	"math-ai.com/math-ai/internal/shared/utils"
)

type ProgramHandler struct {
	programSvc *Service
}

func NewProgramHandler(programSvc *Service) *ProgramHandler {
	return &ProgramHandler{programSvc: programSvc}
}

// POST /programs/list
func (h *ProgramHandler) HandleListPrograms(w http.ResponseWriter, r *http.Request) {
	var req dto.ListProgramsReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	res, err := h.programSvc.ListPrograms(r.Context(), &req)
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

	res, err := h.programSvc.CreateProgram(r.Context(), &req)
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

	res, err := h.programSvc.UpdateProgram(r.Context(), &req)
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

	res, err := h.programSvc.SoftDeleteProgram(r.Context(), &req)
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

	res, err := h.programSvc.ForceDeleteProgram(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	response.WriteJson(w, res, nil)
}

// GET /programs/{id}
func (h *ProgramHandler) HandleGetProgram(w http.ResponseWriter, r *http.Request) {
	req := dto.GetProgramReq{
		ProgramID: utils.StringToInt64(r.PathValue("id"), 0),
		Language:  metadata.GetClientLanguage(r.Context()).ToEnumLanguage(),
	}

	res, err := h.programSvc.GetProgram(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	response.WriteJson(w, res, nil)
}
