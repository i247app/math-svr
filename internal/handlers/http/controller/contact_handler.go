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

// GET - /contact/{id}
func (c *ContactController) HandlerGetContactById(w http.ResponseWriter, r *http.Request) {
	contactID := r.PathValue("id")

	statusCode, contact, err := c.service.GetContactById(r.Context(), contactID)
	if err != nil {
		response.WriteJson(w, r.Context(), nil, err, statusCode)
		return
	}

	res := &dto.GetContactByIdRepsonse{
		Contact: contact,
	}

	response.WriteJson(w, r.Context(), res, nil, statusCode)
}

// GET - /contact
func (c *ContactController) HandlerListContacts(w http.ResponseWriter, r *http.Request) {
	var req dto.ListContactsRequest

	statusCode, contacts, pagination, err := c.service.ListContacts(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, r.Context(), nil, err, statusCode)
		return
	}

	res := &dto.ListContactResponse{
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
func (c *ContactController) HandlerMarkReadContact(w http.ResponseWriter, r *http.Request) {
	var req dto.MarkReadContactRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, r.Context(), nil, fmt.Errorf("invalid parameters"), status.FAIL)
		return
	}

	statusCode, contactRes, err := c.service.MarkReadContact(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, r.Context(), nil, err, statusCode)
		return
	}

	res := &dto.MarkReadContactResponse{
		Contact: contactRes,
	}

	response.WriteJson(w, r.Context(), res, nil, statusCode)
}
