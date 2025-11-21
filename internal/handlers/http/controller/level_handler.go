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

type LevelController struct {
	appResources *resources.AppResource
	service      di.ILevelService
}

func NewLevelController(appResources *resources.AppResource, service di.ILevelService) *LevelController {
	return &LevelController{
		appResources: appResources,
		service:      service,
	}
}

// GET - /levels/list
func (c *LevelController) HandlerGetListLevels(w http.ResponseWriter, r *http.Request) {
	var req dto.ListLevelRequest

	// Parse query parameters
	query := r.URL.Query()
	if search := query.Get("search"); search != "" {
		req.Search = search
	}

	statusCode, levels, pagination, err := c.service.ListLevels(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, r.Context(), nil, err, statusCode)
		return
	}

	res := &dto.ListLevelResponse{
		Items:      levels,
		Pagination: pagination,
	}

	response.WriteJson(w, r.Context(), res, nil, statusCode)
}

// GET - /levels/{id}
func (c *LevelController) HandlerGetLevel(w http.ResponseWriter, r *http.Request) {
	levelID := r.PathValue("id")

	statusCode, level, err := c.service.GetLevelByID(r.Context(), levelID)
	if err != nil {
		response.WriteJson(w, r.Context(), nil, err, statusCode)
		return
	}

	res := &dto.GetLevelResponse{
		Level: level,
	}

	response.WriteJson(w, r.Context(), res, nil, statusCode)
}

// GET - /levels/label/{label}
func (c *LevelController) HandlerGetLevelByLabel(w http.ResponseWriter, r *http.Request) {
	label := r.PathValue("label")

	statusCode, level, err := c.service.GetLevelByLabel(r.Context(), label)
	if err != nil {
		response.WriteJson(w, r.Context(), nil, err, statusCode)
		return
	}

	res := &dto.GetLevelResponse{
		Level: level,
	}

	response.WriteJson(w, r.Context(), res, nil, statusCode)
}

// POST - /levels/create
func (c *LevelController) HandlerCreateLevel(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateLevelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, r.Context(), nil, fmt.Errorf("invalid parameters"), status.BAD_REQUEST)
		return
	}

	statusCode, level, err := c.service.CreateLevel(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, r.Context(), nil, err, statusCode)
		return
	}

	res := &dto.CreateLevelResponse{
		Level: level,
	}

	response.WriteJson(w, r.Context(), res, nil, statusCode)
}

// POST - /levels/update
func (c *LevelController) HandlerUpdateLevel(w http.ResponseWriter, r *http.Request) {
	var req dto.UpdateLevelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, r.Context(), nil, fmt.Errorf("invalid parameters"), status.BAD_REQUEST)
		return
	}

	statusCode, level, err := c.service.UpdateLevel(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, r.Context(), nil, err, statusCode)
		return
	}

	res := &dto.UpdateLevelResponse{
		Level: level,
	}

	response.WriteJson(w, r.Context(), res, nil, statusCode)
}

// POST - /levels/delete
func (c *LevelController) HandlerDeleteLevel(w http.ResponseWriter, r *http.Request) {
	var req dto.DeleteLevelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, r.Context(), nil, fmt.Errorf("invalid parameters"), status.BAD_REQUEST)
		return
	}

	statusCode, err := c.service.DeleteLevel(r.Context(), req.ID)
	if err != nil {
		response.WriteJson(w, r.Context(), nil, err, statusCode)
		return
	}

	response.WriteJson(w, r.Context(), "Level deleted successfully", nil, statusCode)
}

// POST - /levels/force-delete
func (c *LevelController) HandlerForceDeleteLevel(w http.ResponseWriter, r *http.Request) {
	var req dto.DeleteLevelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, r.Context(), nil, fmt.Errorf("invalid parameters"), status.BAD_REQUEST)
		return
	}

	statusCode, err := c.service.ForceDeleteLevel(r.Context(), req.ID)
	if err != nil {
		response.WriteJson(w, r.Context(), nil, err, statusCode)
		return
	}

	response.WriteJson(w, r.Context(), "Level permanently deleted successfully", nil, statusCode)
}
