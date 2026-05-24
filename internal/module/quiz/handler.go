package quiz

import (
	"encoding/json"
	"errors"
	"net/http"

	dto "math-ai.com/math-ai/internal/application/dto/quiz"
	"math-ai.com/math-ai/internal/application/resource"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/response"
)

type QuizHandler struct {
	appResource *resource.Resource
	quizSvc     *Service
}

func NewQuizHandler(appResource *resource.Resource, quizSvc *Service) *QuizHandler {
	return &QuizHandler{
		appResource: appResource,
		quizSvc:     quizSvc,
	}
}

// POST /quizzes/generate
func (h *QuizHandler) HandleGenerateQuiz(w http.ResponseWriter, r *http.Request) {
	var req dto.GenerateQuizReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	session, err := h.appResource.GetRequestSession(r)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	uid, ok := session.UID()
	if !ok {
		response.WriteJson(w, nil, errs.NewError(r.Context(), status.UNAUTHORIZED, nil, errors.New("uid not found from session")))
		return
	}

	req.UserID = &uid

	res, err := h.quizSvc.GenerateQuiz(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// POST /quizzes/submit
func (h *QuizHandler) HandleSubmitQuizAnswers(w http.ResponseWriter, r *http.Request) {
	var req dto.SubmitQuizAnswersReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	res, err := h.quizSvc.SubmitQuizAnswers(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// GET /quizzes/{id}
func (h *QuizHandler) HandleGetQuiz(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")

	res, err := h.quizSvc.GetQuizByQuizId(r.Context(), &dto.GetQuizByQuizIdReq{QuizID: idStr})
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// POST /quizzes/list
func (h *QuizHandler) HandleListQuizzes(w http.ResponseWriter, r *http.Request) {
	var req dto.ListQuizzesReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	res, err := h.quizSvc.ListQuizzes(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// POST /quizzes/soft-delete
func (h *QuizHandler) HandleSoftDeleteQuiz(w http.ResponseWriter, r *http.Request) {
	var req dto.DeleteQuizReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	res, err := h.quizSvc.SoftDeleteQuiz(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}
