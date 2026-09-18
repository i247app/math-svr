package exam

import (
	"encoding/json"
	"net/http"

	dto "math-ai.com/math-ai/internal/application/dto/exam"
	"math-ai.com/math-ai/internal/application/resource"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/response"
)

// ExamHandler decodes, injects the session's user id, and delegates. The
// user id is taken from the session on every route and never from the
// body: the body is the client's claim, the session is the server's.
type ExamHandler struct {
	appResource *resource.Resource
	examSvc     *Service
}

func NewExamHandler(appResource *resource.Resource, examSvc *Service) *ExamHandler {
	return &ExamHandler{appResource: appResource, examSvc: examSvc}
}

// uid pulls the authenticated user id out of the request's session.
func (h *ExamHandler) uid(w http.ResponseWriter, r *http.Request) (*int64, bool) {
	session, err := h.appResource.GetRequestSession(r)
	if err != nil {
		response.WriteJson(w, nil, err)
		return nil, false
	}
	id, ok := session.UID()
	if !ok {
		response.WriteJson(w, nil, errs.NewError(r.Context(), status.UNAUTHORIZED, nil, ErrUidNotFoundFromSession))
		return nil, false
	}
	return &id, true
}

// POST /exams/generate
func (h *ExamHandler) HandleGenerateExam(w http.ResponseWriter, r *http.Request) {
	var req dto.GenerateExamReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	uid, ok := h.uid(w, r)
	if !ok {
		return
	}
	req.UserID = uid

	res, err := h.examSvc.GenerateExam(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// POST /exams/submit
func (h *ExamHandler) HandleSubmitExam(w http.ResponseWriter, r *http.Request) {
	var req dto.SubmitExamReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	uid, ok := h.uid(w, r)
	if !ok {
		return
	}
	req.UserID = uid

	res, err := h.examSvc.SubmitExam(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// POST /exams/detail
func (h *ExamHandler) HandleGetExam(w http.ResponseWriter, r *http.Request) {
	var req dto.GetExamReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	uid, ok := h.uid(w, r)
	if !ok {
		return
	}
	req.UserID = uid

	res, err := h.examSvc.GetExam(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// POST /exams/list
func (h *ExamHandler) HandleListExams(w http.ResponseWriter, r *http.Request) {
	var req dto.ListExamsReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	uid, ok := h.uid(w, r)
	if !ok {
		return
	}
	req.UserID = uid

	res, err := h.examSvc.ListExams(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// POST /exams/stats
func (h *ExamHandler) HandleGetExamStats(w http.ResponseWriter, r *http.Request) {
	var req dto.GetExamStatsReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	uid, ok := h.uid(w, r)
	if !ok {
		return
	}
	req.UserID = uid

	res, err := h.examSvc.GetExamStats(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// POST /exams/analytics/progress
func (h *ExamHandler) HandleGetExamProgress(w http.ResponseWriter, r *http.Request) {
	var req dto.ExamProgressReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	uid, ok := h.uid(w, r)
	if !ok {
		return
	}
	req.UserID = uid

	res, err := h.examSvc.GetExamProgress(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// POST /exams/journey/progress
func (h *ExamHandler) HandleGetJourneyProgress(w http.ResponseWriter, r *http.Request) {
	var req dto.JourneyProgressReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	uid, ok := h.uid(w, r)
	if !ok {
		return
	}
	req.UserID = uid

	res, err := h.examSvc.GetJourneyProgress(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// POST /exams/journeys/mark
func (h *ExamHandler) HandleMarkExamJourney(w http.ResponseWriter, r *http.Request) {
	var req dto.MarkExamJourneyReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	uid, ok := h.uid(w, r)
	if !ok {
		return
	}
	req.UserID = uid

	res, err := h.examSvc.MarkExamJourney(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}
