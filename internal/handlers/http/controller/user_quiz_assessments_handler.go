package controller

import (
	"encoding/json"
	"fmt"
	"net/http"

	"math-ai.com/math-ai/internal/app/resources"
	"math-ai.com/math-ai/internal/applications/dto"
	di "math-ai.com/math-ai/internal/core/di/services"
	"math-ai.com/math-ai/internal/shared/constant/status"
	"math-ai.com/math-ai/internal/shared/utils/response"
)

type UserQuizAssessmentsController struct {
	appResources *resources.AppResource
	service      di.IUserQuizAssessmentService
}

func NewUserQuizAssessmentsController(
	appResources *resources.AppResource,
	service di.IUserQuizAssessmentService,
) *UserQuizAssessmentsController {
	return &UserQuizAssessmentsController{
		appResources: appResources,
		service:      service,
	}
}

// POST /quiz-assessments/generate
func (c *UserQuizAssessmentsController) HandleGenerateQuizAssessments(w http.ResponseWriter, r *http.Request) {
	var req dto.GenerateQuizAssessmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, r.Context(), nil, fmt.Errorf("invalid request body"), status.FAIL)
		return
	}

	// Send message to service
	statusCode, res, err := c.service.GenerateQuizAssessment(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, r.Context(), nil, err, statusCode)
		return
	}

	response.WriteJson(w, r.Context(), res, nil, statusCode)
}

// POST /quiz-assessments/submit
func (c *UserQuizAssessmentsController) HandleSubmitQuizAssessments(w http.ResponseWriter, r *http.Request) {
	var req dto.SubmitQuizAssessmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, r.Context(), nil, fmt.Errorf("invalid request body"), status.FAIL)
		return
	}

	// Send message to service
	statusCode, res, err := c.service.SubmitQuizAssessment(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, r.Context(), nil, err, statusCode)
		return
	}

	response.WriteJson(w, r.Context(), res, nil, statusCode)
}

// POST /quiz-assessments/reinforce
func (c *UserQuizAssessmentsController) HandleReinforceQuizAssessments(w http.ResponseWriter, r *http.Request) {
	var req dto.ReinforceQuizAssessmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, r.Context(), nil, fmt.Errorf("invalid request body"), status.FAIL)
		return
	}

	// Send message to service
	statusCode, res, err := c.service.ReinforceQuizAssessment(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, r.Context(), nil, err, statusCode)
		return
	}

	response.WriteJson(w, r.Context(), res, nil, statusCode)
}

// POST /quiz-assessments/submit-reinforce
func (c *UserQuizAssessmentsController) HandleSubmitReinforceQuizAssessments(w http.ResponseWriter, r *http.Request) {
	var req dto.SubmitReinforceQuizAssessmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, r.Context(), nil, fmt.Errorf("invalid request body"), status.FAIL)
		return
	}

	// Send message to service
	statusCode, res, err := c.service.SubmitReinforceQuizAssessment(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, r.Context(), nil, err, statusCode)
		return
	}

	response.WriteJson(w, r.Context(), res, nil, statusCode)
}

// POST /quiz-assessments/history
func (c *UserQuizAssessmentsController) HandleGetUserQuizAssessmentsHistory(w http.ResponseWriter, r *http.Request) {
	var req dto.GetUserQuizAssessmentsHistoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, r.Context(), nil, fmt.Errorf("invalid request body"), status.FAIL)
		return
	}

	// Send message to service
	statusCode, res, err := c.service.GetUserQuizAssessmentsHistory(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, r.Context(), nil, err, statusCode)
		return
	}

	response.WriteJson(w, r.Context(), res, nil, statusCode)
}
