package classroom

import (
	"encoding/json"
	"net/http"

	dto "math-ai.com/math-ai/internal/application/dto/classroomprogress"
	"math-ai.com/math-ai/internal/shared/response"
)

// HandleProgressScoresOverTime is POST /classrooms/progress/scores-over-time.
// Body shape lives in dto.ScoresOverTimeReq; the service runs validation
// + permission gating + the aggregation query.
func (h *ClassroomHandler) HandleProgressScoresOverTime(w http.ResponseWriter, r *http.Request) {
	var req dto.ScoresOverTimeReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	uid, err := h.sessionUID(r)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	res, err := h.classroomSvc.GetScoresOverTime(r.Context(), &req, uid)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// HandleProgressStudents is POST /classrooms/progress/students.
// Returns the three summary cards + the paginated per-student rows
// driving the screenshot's table.
func (h *ClassroomHandler) HandleProgressStudents(w http.ResponseWriter, r *http.Request) {
	var req dto.StudentsProgressReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	uid, err := h.sessionUID(r)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	res, err := h.classroomSvc.GetStudentsProgress(r.Context(), &req, uid)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}
