package dto

import (
	// domain "math-ai.com/math-ai/internal/core/domain/contact"
	"math-ai.com/math-ai/internal/shared/utils/pagination"
)

type ContactResponse struct {
	ID             string `json:"id"`
	UID            string `json:"uid"`
	ContactName    string `json:"contact_name"`
	ContactEmail   string `json:"contact_email"`
	ContactPhone   string `json:"contact_phone"`
	ContactMessage string `json:"contact_message"`
}

type CreateContactRequest struct {
	ContactName    string `json:"contact_name"`
	ContactPhone   string `json:"contact_phone"`
	ContactEmail   string `json:"contact_email"`
	ContactMessage string `json:"contact_message"`
}
type ListContactsRequest struct {
	Page    int64 `json:"page" form:"page"`
	Size    int64 `json:"size" form:"size"`
	TakeAll bool  `json:"take_all" form:"take_all"`
}

type ListContactsParams struct {
	Limit  int64
	Offset int64
}

type GetContactsResponse struct {
	Items      []*ContactResponse     `json:"items"`
	Pagination *pagination.Pagination `json:"metadata"`
}
