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

type UserQuizPracticesController struct {
	appResources *resources.AppResource
	service      di.IUserQuizPracticesService
}

func NewUserQuizPracticesController(
	appResources *resources.AppResource,
	service di.IUserQuizPracticesService,
) *UserQuizPracticesController {
	return &UserQuizPracticesController{
		appResources: appResources,
		service:      service,
	}
}

// POST /quiz-practices/generate
func (c *UserQuizPracticesController) HandleGenerateQuizPractices(w http.ResponseWriter, r *http.Request) {
	var req dto.GenerateQuizRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, r.Context(), nil, fmt.Errorf("invalid request body"), status.FAIL)
		return
	}

	// Send message to service
	statusCode, res, err := c.service.GenerateQuiz(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, r.Context(), nil, err, statusCode)
		return
	}

	response.WriteJson(w, r.Context(), res, nil, statusCode)
}

// POST /quiz-practices/submit
func (c *UserQuizPracticesController) HandleSubmitQuizParctices(w http.ResponseWriter, r *http.Request) {
	var req dto.SubmitQuizRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, r.Context(), nil, fmt.Errorf("invalid request body"), status.FAIL)
		return
	}

	// Send message to service
	statusCode, res, err := c.service.SubmitQuiz(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, r.Context(), nil, err, statusCode)
		return
	}

	response.WriteJson(w, r.Context(), res, nil, statusCode)
}

// POST /quiz-practices/reinforce
func (c *UserQuizPracticesController) HandleReinforceQuizPractices(w http.ResponseWriter, r *http.Request) {
	var req dto.GenerateQuizPracticeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, r.Context(), nil, fmt.Errorf("invalid request body"), status.FAIL)
		return
	}

	// Send message to service
	statusCode, res, err := c.service.GenerateQuizPractice(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, r.Context(), nil, err, statusCode)
		return
	}

	response.WriteJson(w, r.Context(), res, nil, statusCode)
}
