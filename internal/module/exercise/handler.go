package exercise

import (
	"encoding/json"
	"net/http"

	dto "math-ai.com/math-ai/internal/application/dto/exercise"
	"math-ai.com/math-ai/internal/application/resource"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/response"
	"math-ai.com/math-ai/internal/shared/utils"
)

type ClassroomExerciseHandler struct {
	appResource *resource.Resource
	svc         *Service
}

func NewClassroomExerciseHandler(appResource *resource.Resource, svc *Service) *ClassroomExerciseHandler {
	return &ClassroomExerciseHandler{appResource: appResource, svc: svc}
}

// sessionUID extracts the authenticated user's id from the request
// session. Mirrors the classroom handler's helper so the §0 Q1 contract
// (caller's profile must belong to the session user) is uniformly
// enforced across both modules.
func (h *ClassroomExerciseHandler) sessionUID(r *http.Request) (int64, error) {
	sess, err := h.appResource.GetRequestSession(r)
	if err != nil {
		return 0, err
	}
	uid, ok := sess.UID()
	if !ok {
		return 0, errs.NewError(r.Context(), status.UNAUTHORIZED, nil,
			ErrUIDNotFoundFromSession)
	}
	return uid, nil
}

// POST /classroom-exercises/create
func (h *ClassroomExerciseHandler) HandleCreateExercise(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateExerciseReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	uid, err := h.sessionUID(r)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	res, err := h.svc.CreateExercise(r.Context(), &req, uid)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	response.WriteJson(w, res, nil)
}

// POST /classroom-exercises/update
func (h *ClassroomExerciseHandler) HandleUpdateExercise(w http.ResponseWriter, r *http.Request) {
	var req dto.UpdateExerciseReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	uid, err := h.sessionUID(r)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	res, err := h.svc.UpdateExercise(r.Context(), &req, uid)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	response.WriteJson(w, res, nil)
}

// GET /classroom-exercises/{id}
func (h *ClassroomExerciseHandler) HandleGetExercise(w http.ResponseWriter, r *http.Request) {
	id := utils.StringToInt64(r.PathValue("id"), 0)

	uid, err := h.sessionUID(r)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	res, err := h.svc.GetExercise(r.Context(), &dto.GetExerciseReq{ClassroomExerciseID: id}, uid)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	response.WriteJson(w, res, nil)
}

// POST /classroom-exercises/list
func (h *ClassroomExerciseHandler) HandleListExercises(w http.ResponseWriter, r *http.Request) {
	var req dto.ListExercisesReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	uid, err := h.sessionUID(r)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	res, err := h.svc.ListExercises(r.Context(), &req, uid)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	response.WriteJson(w, res, nil)
}

// POST /classroom-exercises/soft-delete
func (h *ClassroomExerciseHandler) HandleSoftDeleteExercise(w http.ResponseWriter, r *http.Request) {
	var req dto.DeleteExerciseReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	uid, err := h.sessionUID(r)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	res, err := h.svc.SoftDeleteExercise(r.Context(), &req, uid)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	response.WriteJson(w, res, nil)
}
