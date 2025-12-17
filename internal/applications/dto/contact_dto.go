package dto

import (
	domain "math-ai.com/math-ai/internal/core/domain/contact"
	"math-ai.com/math-ai/internal/shared/utils/pagination"
)

type ListContactsRequest struct {
	Search    string `json:"search,omitempty" form:"search"`
	Page      int64  `json:"page" form:"page"`
	Limit     int64  `json:"size" form:"size"`
	OrderBy   string `json:"order_by" form:"order_by"`
	OrderDesc bool   `json:"order_desc" form:"order_desc"`
	TakeAll   bool   `json:"take_all" form:"take_all"`
}

type GetContactsResponse struct {
	Items      []*ContactResponse     `json:"items"`
	Pagination *pagination.Pagination `json:"metadata"`
}

type ContactResponse struct {
	ID             string  `json:"id"`
	UID            *string `json:"uid"`
	ContactName    string  `json:"contact_name"`
	ContactEmail   *string `json:"contact_email"`
	ContactPhone   *string `json:"contact_phone"`
	ContactMessage string  `json:"contact_message"`
	IsRead         bool    `json:"is_read"`
}

type CreateContactRequest struct {
	UID            *string `json:"-"`
	ContactName    string  `json:"contact_name"`
	ContactPhone   *string `json:"contact_phone"`
	ContactEmail   *string `json:"contact_email"`
	ContactMessage string  `json:"contact_message"`
}

type CreateContactResponse struct {
	Contact *ContactResponse `json:"contact"`
}

type MarkReadContactRequest struct {
	ContactID string `json:"contact_id"`
}

type MarkReadContactResponse struct {
	Contact *ContactResponse `json:"contact"`
}

func ContactUsResponseFromDomain(contact *domain.Contact) ContactResponse {
	return ContactResponse{
		ID:             contact.ID(),
		UID:            contact.UID(),
		ContactName:    contact.ContactName(),
		ContactEmail:   contact.ContactEmail(),
		ContactPhone:   contact.ContactPhone(),
		ContactMessage: contact.ContactMessage(),
		IsRead:         contact.IsRead(),
	}
}

func BuildContactDomainForSubmit(req *CreateContactRequest) *domain.Contact {
	contact := domain.NewContactDomain()
	contact.GenerateID()
	contact.SetUID(req.UID)
	contact.SetContactName(req.ContactName)
	contact.SetContactEmail(req.ContactEmail)
	contact.SetContactPhone(req.ContactPhone)
	contact.SetContactMessage(req.ContactMessage)

	return contact
}
