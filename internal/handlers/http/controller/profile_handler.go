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

type ProfileController struct {
	appResources *resources.AppResource
	service      di.IProfileService
}

func NewProfileController(appResources *resources.AppResource, service di.IProfileService) *ProfileController {
	return &ProfileController{
		appResources: appResources,
		service:      service,
	}
}

// POST - /profiles/fetch
func (c *ProfileController) HandlerFetchProfile(w http.ResponseWriter, r *http.Request) {
	var req dto.FetchProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, r.Context(), nil, fmt.Errorf("invalid parameters"), status.FAIL)
		return
	}

	statusCode, profile, err := c.service.FetchProfile(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, r.Context(), nil, err, statusCode)
		return
	}

	res := &dto.FetchProfileResponse{
		Profile: profile,
	}

	response.WriteJson(w, r.Context(), res, nil, statusCode)
}

// POST - /profiles/create
func (c *ProfileController) HandlerCreateProfile(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, r.Context(), nil, fmt.Errorf("invalid parameters"), status.FAIL)
		return
	}

	statusCode, profile, err := c.service.CreateProfile(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, r.Context(), nil, err, statusCode)
		return
	}

	res := &dto.CreateProfileResponse{
		Profile: profile,
	}

	response.WriteJson(w, r.Context(), res, nil, statusCode)
}

// POST - /profiles/update
func (c *ProfileController) HandlerUpdateProfile(w http.ResponseWriter, r *http.Request) {
	var req dto.UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, r.Context(), nil, fmt.Errorf("invalid parameters"), status.FAIL)
		return
	}

	statusCode, profile, err := c.service.UpdateProfile(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, r.Context(), nil, err, statusCode)
		return
	}

	res := &dto.UpdateProfileResponse{
		Profile: profile,
	}

	response.WriteJson(w, r.Context(), res, nil, statusCode)
}
