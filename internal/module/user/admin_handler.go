package user

import (
	"encoding/json"
	"net/http"

	dto "math-ai.com/math-ai/internal/application/dto/user"
	"math-ai.com/math-ai/internal/shared/response"
)

// POST /users/soft-delete
func (h *UserHandler) HandleSoftDeleteUser(w http.ResponseWriter, r *http.Request) {
	var req dto.DeleteUserReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	if _, err := h.userSvc.SoftDeleteUser(r.Context(), &req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJsonNoContent(w)
}

// POST /users/force-delete
func (h *UserHandler) HandleForceDeleteUser(w http.ResponseWriter, r *http.Request) {
	var req dto.DeleteUserReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	if _, err := h.userSvc.ForceDeleteUser(r.Context(), &req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJsonNoContent(w)
}
