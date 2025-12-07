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
	"math-ai.com/math-ai/internal/shared/utils/response"
)

type SemesterController struct {
	appResources *resources.AppResource
	service      di.ISemesterService
}

func NewSemesterController(appResources *resources.AppResource, service di.ISemesterService) *SemesterController {
	return &SemesterController{
		appResources: appResources,
		service:      service,
	}
}

// GET - /semesters/list
func (c *SemesterController) HandlerGetListSemesters(w http.ResponseWriter, r *http.Request) {
	var req dto.ListSemesterRequest

	// Parse query parameters
	query := r.URL.Query()
	if search := query.Get("search"); search != "" {
		req.Search = search
	}

	statusCode, semesters, pagination, err := c.service.ListSemesters(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, r.Context(), nil, err, statusCode)
		return
	}

	res := &dto.ListSemesterResponse{
		Items:      semesters,
		Pagination: pagination,
	}

	response.WriteJson(w, r.Context(), res, nil, statusCode)
}

// GET - /semesters/{id}
func (c *SemesterController) HandlerGetSemester(w http.ResponseWriter, r *http.Request) {
	semesterID := r.PathValue("id")

	statusCode, semester, err := c.service.GetSemesterByID(r.Context(), semesterID)
	if err != nil {
		response.WriteJson(w, r.Context(), nil, err, statusCode)
		return
	}

	res := &dto.GetSemesterResponse{
		Semester: semester,
	}

	response.WriteJson(w, r.Context(), res, nil, statusCode)
}

// GET - /semesters/name/{name}
func (c *SemesterController) HandlerGetSemesterByName(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")

	statusCode, semester, err := c.service.GetSemesterByName(r.Context(), name)
	if err != nil {
		response.WriteJson(w, r.Context(), nil, err, statusCode)
		return
	}

	res := &dto.GetSemesterResponse{
		Semester: semester,
	}

	response.WriteJson(w, r.Context(), res, nil, statusCode)
}

// POST - /semesters/create
func (c *SemesterController) HandlerCreateSemester(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateSemesterRequest

	// Multipart form request (with icon)
	if err := r.ParseMultipartForm(MaxAvatarUploadSize); err != nil {
		response.WriteJson(w, r.Context(), nil, fmt.Errorf("invalid form data"), status.FAIL)
		return
	}

	// Parse form fields
	req.Name = r.FormValue("name")
	description := r.FormValue("description")
	req.Description = &description
	displayOrder, _ := strconv.ParseInt(r.FormValue("display_order"), 10, 8)
	req.DisplayOrder = int8(displayOrder)

	// Handle icon file
	file, header, err := r.FormFile("image")
	if err == nil {
		defer file.Close()
		req.IconFile = file
		req.IconFilename = header.Filename
		req.IconContentType = header.Header.Get("Content-Type")
	}

	statusCode, semester, err := c.service.CreateSemester(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, r.Context(), nil, err, statusCode)
		return
	}

	res := &dto.CreateSemesterResponse{
		Semester: semester,
	}

	response.WriteJson(w, r.Context(), res, nil, statusCode)
}

// POST - /semesters/update
func (c *SemesterController) HandlerUpdateSemester(w http.ResponseWriter, r *http.Request) {
	var req dto.UpdateSemesterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, r.Context(), nil, fmt.Errorf("invalid parameters"), status.FAIL)
		return
	}

	statusCode, semester, err := c.service.UpdateSemester(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, r.Context(), nil, err, statusCode)
		return
	}

	res := &dto.UpdateSemesterResponse{
		Semester: semester,
	}

	response.WriteJson(w, r.Context(), res, nil, statusCode)
}

// POST - /semesters/delete
func (c *SemesterController) HandlerDeleteSemester(w http.ResponseWriter, r *http.Request) {
	var req dto.DeleteSemesterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, r.Context(), nil, fmt.Errorf("invalid parameters"), status.FAIL)
		return
	}

	statusCode, err := c.service.DeleteSemester(r.Context(), req.ID)
	if err != nil {
		response.WriteJson(w, r.Context(), nil, err, statusCode)
		return
	}

	response.WriteJson(w, r.Context(), "Semester deleted successfully", nil, statusCode)
}

// POST - /semesters/force-delete
func (c *SemesterController) HandlerForceDeleteSemester(w http.ResponseWriter, r *http.Request) {
	var req dto.DeleteSemesterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, r.Context(), nil, fmt.Errorf("invalid parameters"), status.FAIL)
		return
	}

	statusCode, err := c.service.ForceDeleteSemester(r.Context(), req.ID)
	if err != nil {
		response.WriteJson(w, r.Context(), nil, err, statusCode)
		return
	}

	response.WriteJson(w, r.Context(), "Semester permanently deleted successfully", nil, statusCode)
}
