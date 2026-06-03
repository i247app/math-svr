package classroomexercise

import (
	"encoding/json"
	"net/http"
	"strconv"

	dto "math-ai.com/math-ai/internal/application/dto/classroomexercise"
	"math-ai.com/math-ai/internal/application/resource"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/response"
)

type ClassroomExerciseSubmissionHandler struct {
	appResource *resource.Resource
	svc         *Service
}

func NewClassroomExerciseSubmissionHandler(appResource *resource.Resource, svc *Service) *ClassroomExerciseSubmissionHandler {
	return &ClassroomExerciseSubmissionHandler{appResource: appResource, svc: svc}
}

// sessionUID extracts the authenticated user's id from the request
// session. Mirrors the exercise handler's helper so the §0 Q1 contract
// stays consistent.
func (h *ClassroomExerciseSubmissionHandler) sessionUID(r *http.Request) (int64, error) {
	sess, err := h.appResource.GetRequestSession(r)
	if err != nil {
		return 0, err
	}
	uid, ok := sess.UID()
	if !ok {
		return 0, errs.NewError(r.Context(), status.UNAUTHORIZED, nil,
			ErrUidNotFoundFromSession)
	}
	return uid, nil
}

// POST /classroom-exercise-submissions/submit
func (h *ClassroomExerciseSubmissionHandler) HandleSubmitExerciseAnswers(w http.ResponseWriter, r *http.Request) {
	var req dto.SubmitExerciseAnswersReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	uid, err := h.sessionUID(r)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	res, err := h.svc.SubmitExerciseAnswers(r.Context(), &req, uid)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// GET /classroom-exercise-submissions/{id}
func (h *ClassroomExerciseSubmissionHandler) HandleGetSubmission(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id == 0 {
		response.WriteJson(w, nil, errs.NewError(r.Context(),
			status.CLASSROOM_EXERCISE_SUBMISSION_MISSING_ID, nil, err))
		return
	}
	uid, err := h.sessionUID(r)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	res, err := h.svc.GetSubmission(r.Context(), &dto.GetSubmissionReq{
		ClassroomExerciseSubmissionID: id,
	}, uid)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// POST /classroom-exercise-submissions/list
//
// Flexible list endpoint. profile_id is optional — see
// dto.ListSubmissionsReq for the full permission matrix.
func (h *ClassroomExerciseSubmissionHandler) HandleListSubmissions(w http.ResponseWriter, r *http.Request) {
	var req dto.ListSubmissionsReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	uid, err := h.sessionUID(r)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	res, err := h.svc.ListSubmissions(r.Context(), &req, uid)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// POST /classroom-exercise-submissions/by-exercise
func (h *ClassroomExerciseSubmissionHandler) HandleListSubmissionsByExercise(w http.ResponseWriter, r *http.Request) {
	var req dto.ListSubmissionsByExerciseReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	uid, err := h.sessionUID(r)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	res, err := h.svc.ListSubmissionsByExercise(r.Context(), &req, uid)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// POST /classroom-exercises/submissions/submitted-members
//
// Teacher-side roster of members who already submitted the exercise.
func (h *ClassroomExerciseSubmissionHandler) HandleListSubmittedMembers(w http.ResponseWriter, r *http.Request) {
	var req dto.ListAudienceMembersReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	uid, err := h.sessionUID(r)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	res, err := h.svc.ListSubmittedMembers(r.Context(), &req, uid)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// POST /classroom-exercises/submissions/non-submitted-members
//
// Teacher-side roster of ACTIVE classroom members who have NOT
// submitted the exercise.
func (h *ClassroomExerciseSubmissionHandler) HandleListNonSubmittedMembers(w http.ResponseWriter, r *http.Request) {
	var req dto.ListAudienceMembersReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	uid, err := h.sessionUID(r)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	res, err := h.svc.ListNonSubmittedMembers(r.Context(), &req, uid)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// POST /classroom-exercise-submissions/soft-delete
func (h *ClassroomExerciseSubmissionHandler) HandleSoftDeleteSubmission(w http.ResponseWriter, r *http.Request) {
	var req dto.DeleteSubmissionReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	uid, err := h.sessionUID(r)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	res, err := h.svc.SoftDeleteSubmission(r.Context(), &req, uid)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}
