package program

import (
	"encoding/json"
	"net/http"

	dto "math-ai.com/math-ai/internal/application/dto/program"
	"math-ai.com/math-ai/internal/shared/response"
)

type ProgramHandler struct {
	programSvc *Service
}

func NewProgramHandler(programSvc *Service) *ProgramHandler {
	return &ProgramHandler{programSvc: programSvc}
}

// POST /programs/list
func (h *ProgramHandler) ListPrograms(w http.ResponseWriter, r *http.Request) {
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
