package classroom

import (
	"encoding/json"
	"fmt"
	"net/http"

	dto "math-ai.com/math-ai/internal/application/dto/classroomprogress"
	"math-ai.com/math-ai/internal/shared/response"
)

// HandleProfileProgress serves POST /classrooms/progress/profile — the
// single-student learning-progress detail (chart series + summary cards).
// JSON-only; the caller's session UID gates which profiles they may view.
func (h *ClassroomHandler) HandleProfileProgress(w http.ResponseWriter, r *http.Request) {
	var req dto.ProfileProgressReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, fmt.Errorf("invalid parameters"))
		return
	}

	uid, err := h.sessionUID(r)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	res, err := h.classroomSvc.GetProfileProgress(r.Context(), req, uid)
	response.WriteJson(w, res, err)
}
