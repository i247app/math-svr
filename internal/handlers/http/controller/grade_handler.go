package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"math-ai.com/math-ai/internal/app/resources"
	"math-ai.com/math-ai/internal/applications/dto"
	di "math-ai.com/math-ai/internal/core/di/services"
	"math-ai.com/math-ai/internal/shared/constant/status"
	"math-ai.com/math-ai/internal/shared/logger"
	"math-ai.com/math-ai/internal/shared/utils/response"
)

type GradeController struct {
	appResources *resources.AppResource
	service      di.IGradeService
}

func NewGradeController(appResources *resources.AppResource, service di.IGradeService) *GradeController {
	return &GradeController{
		appResources: appResources,
		service:      service,
	}
}

// GET - /grades/list
func (c *GradeController) HandlerGetListGrades(w http.ResponseWriter, r *http.Request) {
	var req dto.ListGradeRequest

	// Parse query parameters
	query := r.URL.Query()
	if search := query.Get("search"); search != "" {
		req.Search = search
	}

	statusCode, grades, pagination, err := c.service.ListGrades(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, r.Context(), nil, err, statusCode)
		return
	}

	res := &dto.ListGradeResponse{
		Items:      grades,
		Pagination: pagination,
	}

	response.WriteJson(w, r.Context(), res, nil, statusCode)
}

// GET - /grades/{id}
func (c *GradeController) HandlerGetGrade(w http.ResponseWriter, r *http.Request) {
	gradeID := r.PathValue("id")

	statusCode, grade, err := c.service.GetGradeByID(r.Context(), gradeID)
	if err != nil {
		response.WriteJson(w, r.Context(), nil, err, statusCode)
		return
	}

	res := &dto.GetGradeResponse{
		Grade: grade,
	}

	response.WriteJson(w, r.Context(), res, nil, statusCode)
}

// GET - /grades/label/{label}
func (c *GradeController) HandlerGetGradeByLabel(w http.ResponseWriter, r *http.Request) {
	label := r.PathValue("label")

	statusCode, grade, err := c.service.GetGradeByLabel(r.Context(), label)
	if err != nil {
		response.WriteJson(w, r.Context(), nil, err, statusCode)
		return
	}

	res := &dto.GetGradeResponse{
		Grade: grade,
	}

	response.WriteJson(w, r.Context(), res, nil, statusCode)
}

// POST - /grades/create
func (c *GradeController) HandlerCreateGrade(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateGradeRequest

	// Multipart form request (with avatar)
	if err := r.ParseMultipartForm(MaxAvatarUploadSize); err != nil {
		logger.Errorf("Failed to parse multipart form: %v", err)
		response.WriteJson(w, r.Context(), nil, fmt.Errorf("invalid form data"), status.BAD_REQUEST)
		return
	}

	// Parse form fields
	req.Label = r.FormValue("label")
	req.Description = r.FormValue("description")
	displayOrder, _ := strconv.ParseInt(r.FormValue("display_order"), 10, 8)
	req.DisplayOrder = int8(displayOrder)

	// Parse role

	// Handle avatar file
	file, header, err := r.FormFile("icon")
	if err == nil {
		defer file.Close()
		req.IconFile = file
		req.IconFilename = header.Filename
		req.IconContentType = header.Header.Get("Content-Type")
	}

	statusCode, grade, err := c.service.CreateGrade(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, r.Context(), nil, err, statusCode)
		return
	}

	res := &dto.CreateGradeResponse{
		Grade: grade,
	}

	response.WriteJson(w, r.Context(), res, nil, statusCode)
}

// POST - /grades/update
func (c *GradeController) HandlerUpdateGrade(w http.ResponseWriter, r *http.Request) {
	var req dto.UpdateGradeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, r.Context(), nil, fmt.Errorf("invalid parameters"), status.BAD_REQUEST)
		return
	}

	statusCode, grade, err := c.service.UpdateGrade(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, r.Context(), nil, err, statusCode)
		return
	}

	res := &dto.UpdateGradeResponse{
		Grade: grade,
	}

	response.WriteJson(w, r.Context(), res, nil, statusCode)
}

// POST - /grades/delete
func (c *GradeController) HandlerDeleteGrade(w http.ResponseWriter, r *http.Request) {
	var req dto.DeleteGradeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, r.Context(), nil, fmt.Errorf("invalid parameters"), status.BAD_REQUEST)
		return
	}

	statusCode, err := c.service.DeleteGrade(r.Context(), req.ID)
	if err != nil {
		response.WriteJson(w, r.Context(), nil, err, statusCode)
		return
	}

	response.WriteJson(w, r.Context(), "Grade deleted successfully", nil, statusCode)
}

// POST - /grades/force-delete
func (c *GradeController) HandlerForceDeleteGrade(w http.ResponseWriter, r *http.Request) {
	var req dto.DeleteGradeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, r.Context(), nil, fmt.Errorf("invalid parameters"), status.BAD_REQUEST)
		return
	}

	statusCode, err := c.service.ForceDeleteGrade(r.Context(), req.ID)
	if err != nil {
		response.WriteJson(w, r.Context(), nil, err, statusCode)
		return
	}

	response.WriteJson(w, r.Context(), "Grade permanently deleted successfully", nil, statusCode)
}
