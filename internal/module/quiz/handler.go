package quiz

import (
	"encoding/json"
	"net/http"

	dto "math-ai.com/math-ai/internal/application/dto/quiz"
	"math-ai.com/math-ai/internal/shared/response"
	"math-ai.com/math-ai/internal/shared/utils"
)

type QuizHandler struct {
	quizSvc *Service
}

func NewQuizHandler(quizSvc *Service) *QuizHandler {
	return &QuizHandler{quizSvc: quizSvc}
}

// POST /quizzes/generate
func (h *QuizHandler) HandleGenerateQuiz(w http.ResponseWriter, r *http.Request) {
	var req dto.GenerateQuizReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}

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
	quizID, err := utils.StringToUUID(idStr)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	res, err := h.quizSvc.GetQuizByQuizId(r.Context(), &dto.GetQuizByQuizIdReq{QuizID: quizID})
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// POST /quizzes/list
func (h *QuizHandler) HandleListQuizzes(w http.ResponseWriter, r *http.Request) {
	var req dto.ListQuizzesByProfileIdReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	res, err := h.quizSvc.ListQuizzesByProfileId(r.Context(), &req)
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
