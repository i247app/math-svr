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

type ContactController struct {
	appResources *resources.AppResource
	service      di.IContactService
}

func NewContactController(appResources *resources.AppResource, service di.IContactService) *ContactController {
	return &ContactController{
		appResources: appResources,
		service:      service,
	}
}

// GET - /contact
func (c *ContactController) HandlerListContacts(w http.ResponseWriter, r *http.Request) {
	var req dto.ListContactsRequest

	// Parse query parameters
	if err := r.ParseForm(); err != nil {
		response.WriteJson(w, r.Context(), nil, fmt.Errorf("invalid query parameters"), status.FAIL)
		return
	}

	if pageStr := r.FormValue("page"); pageStr != "" {
		if page, err := strconv.ParseInt(pageStr, 10, 64); err == nil {
			req.Page = page
		}
	}
	if req.Page == 0 {
		req.Page = 1
	}

	if limitStr := r.FormValue("limit"); limitStr != "" {
		if limit, err := strconv.ParseInt(limitStr, 10, 64); err == nil {
			req.Limit = limit
		}
	}

	// Parse take_all
	if takeAllStr := r.FormValue("take_all"); takeAllStr == "true" {
		req.TakeAll = true
	}

	statusCode, contacts, pagination, err := c.service.ListGrades(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, r.Context(), nil, err, statusCode)
		return
	}

	res := &dto.GetContactsResponse{
		Items:      contacts,
		Pagination: pagination,
	}

	response.WriteJson(w, r.Context(), res, nil, statusCode)
}

// POST - /contact/submit
func (c *ContactController) HandlerCreateContact(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateContactRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, r.Context(), nil, fmt.Errorf("invalid parameters"), status.FAIL)
		return
	}

	uid, _ := c.appResources.GetRequestUID(r)

	req.UID = &uid

	statusCode, contactRes, err := c.service.SubmitContact(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, r.Context(), nil, err, statusCode)
		return
	}

	res := &dto.CreateContactResponse{
		Contact: contactRes,
	}

	response.WriteJson(w, r.Context(), res, nil, statusCode)
}

// POST - /contact/mark-read
func (c *ContactController) HandlerCheckReadContact(w http.ResponseWriter, r *http.Request) {
	var req dto.MarkReadContactRequest

	// Decode request body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, r.Context(), nil, fmt.Errorf("invalid parameters"), status.FAIL)
		return
	}

	// Call service to mark contact as read
	statusCode, contactRes, err := c.service.MarkReadContact(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, r.Context(), nil, err, statusCode)
		return
	}

	res := &dto.MarkReadContactResponse{
		Contact: contactRes,
	}

	// Return success response with updated contact
	response.WriteJson(w, r.Context(), res, nil, statusCode)
}
