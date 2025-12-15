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

type TermController struct {
	appResources *resources.AppResource
	service      di.ITermService
}

func NewTermController(appResources *resources.AppResource, service di.ITermService) *TermController {
	return &TermController{
		appResources: appResources,
		service:      service,
	}
}

// GET - /terms/list
func (c *TermController) HandlerGetListTerms(w http.ResponseWriter, r *http.Request) {
	var req dto.ListTermRequest

	// Parse query parameters
	query := r.URL.Query()
	if search := query.Get("search"); search != "" {
		req.Search = search
	}

	statusCode, terms, pagination, err := c.service.ListTerms(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, r.Context(), nil, err, statusCode)
		return
	}

	res := &dto.ListTermResponse{
		Items:      terms,
		Pagination: pagination,
	}

	response.WriteJson(w, r.Context(), res, nil, statusCode)
}

// GET - /terms/{id}
func (c *TermController) HandlerGetTerm(w http.ResponseWriter, r *http.Request) {
	termID := r.PathValue("id")

	statusCode, term, err := c.service.GetTermByID(r.Context(), termID)
	if err != nil {
		response.WriteJson(w, r.Context(), nil, err, statusCode)
		return
	}

	res := &dto.GetTermResponse{
		Term: term,
	}

	response.WriteJson(w, r.Context(), res, nil, statusCode)
}

// GET - /terms/name/{name}
func (c *TermController) HandlerGetTermByName(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")

	statusCode, term, err := c.service.GetTermByName(r.Context(), name)
	if err != nil {
		response.WriteJson(w, r.Context(), nil, err, statusCode)
		return
	}

	res := &dto.GetTermResponse{
		Term: term,
	}

	response.WriteJson(w, r.Context(), res, nil, statusCode)
}

// POST - /terms/create
func (c *TermController) HandlerCreateTerm(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateTermRequest

	// Check content type - support both JSON and multipart form
	contentType := r.Header.Get("Content-Type")

	if contentType == "application/json" {
		// JSON request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response.WriteJson(w, r.Context(), nil, fmt.Errorf("invalid parameters"), status.FAIL)
			return
		}
	} else {
		// Multipart form request (with icon)
		if err := r.ParseMultipartForm(MaxAvatarUploadSize); err != nil {
			response.WriteJson(w, r.Context(), nil, fmt.Errorf("invalid form data"), status.FAIL)
			return
		}

		// Parse form fields
		req.Name = r.FormValue("name")
		if description := r.FormValue("description"); description != "" {
			req.Description = &description
		}

		displayOrder, _ := strconv.ParseInt(r.FormValue("display_order"), 10, 8)
		req.DisplayOrder = int8(displayOrder)

		// Handle icon file
		file, header, err := r.FormFile("image")
		if err == nil {
			defer file.Close()
			req.ImageFile = file
			req.ImageFilename = header.Filename
			req.ImageContentType = header.Header.Get("Content-Type")
		}
	}

	statusCode, term, err := c.service.CreateTerm(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, r.Context(), nil, err, statusCode)
		return
	}

	res := &dto.CreateTermResponse{
		Term: term,
	}

	response.WriteJson(w, r.Context(), res, nil, statusCode)
}

// POST - /terms/update
func (c *TermController) HandlerUpdateTerm(w http.ResponseWriter, r *http.Request) {
	var req dto.UpdateTermRequest
	// Check content type - support both JSON and multipart form
	contentType := r.Header.Get("Content-Type")

	if contentType == "application/json" {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response.WriteJson(w, r.Context(), nil, fmt.Errorf("invalid parameters"), status.FAIL)
			return
		}
	} else {
		// Multipart form request (with icon)
		if err := r.ParseMultipartForm(MaxAvatarUploadSize); err != nil {
			response.WriteJson(w, r.Context(), nil, fmt.Errorf("invalid form data"), status.FAIL)
			return
		}

		// Parse form fields
		req.ID = r.FormValue("id")

		// Optional fields
		if name := r.FormValue("name"); name != "" {
			req.Name = &name
		}
		if description := r.FormValue("description"); description != "" {
			req.Description = &description
		}

		if displayOrderStr := r.FormValue("display_order"); displayOrderStr != "" {
			if displayOrder, err := strconv.ParseInt(displayOrderStr, 10, 8); err == nil {
				displayOrderInt8 := int8(displayOrder)
				req.DisplayOrder = &displayOrderInt8
			}
		}

		// Handle icon file
		file, header, err := r.FormFile("image")
		if err == nil {
			defer file.Close()
			req.ImageFile = file
			req.ImageFilename = header.Filename
			req.ImageContentType = header.Header.Get("Content-Type")
		}
	}

	statusCode, term, err := c.service.UpdateTerm(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, r.Context(), nil, err, statusCode)
		return
	}

	res := &dto.UpdateTermResponse{
		Term: term,
	}

	response.WriteJson(w, r.Context(), res, nil, statusCode)
}

// POST - /terms/delete
func (c *TermController) HandlerDeleteTerm(w http.ResponseWriter, r *http.Request) {
	var req dto.DeleteTermRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, r.Context(), nil, fmt.Errorf("invalid parameters"), status.FAIL)
		return
	}

	statusCode, err := c.service.DeleteTerm(r.Context(), req.ID)
	if err != nil {
		response.WriteJson(w, r.Context(), nil, err, statusCode)
		return
	}

	response.WriteJson(w, r.Context(), "Term deleted successfully", nil, statusCode)
}

// POST - /terms/force-delete
func (c *TermController) HandlerForceDeleteTerm(w http.ResponseWriter, r *http.Request) {
	var req dto.DeleteTermRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, r.Context(), nil, fmt.Errorf("invalid parameters"), status.FAIL)
		return
	}

	statusCode, err := c.service.ForceDeleteTerm(r.Context(), req.ID)
	if err != nil {
		response.WriteJson(w, r.Context(), nil, err, statusCode)
		return
	}

	response.WriteJson(w, r.Context(), "Term permanently deleted successfully", nil, statusCode)
}
