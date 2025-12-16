package dto

import (
	"fmt"
	"strings"

	// domain "math-ai.com/math-ai/internal/core/domain/contact"
	"math-ai.com/math-ai/internal/shared/utils/pagination"
	"math-ai.com/math-ai/internal/shared/utils/validate"
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

func (r *CreateContactRequest) Validate() error {
	// Trim whitespace from all fields
	r.ContactName = strings.TrimSpace(r.ContactName)
	r.ContactEmail = strings.TrimSpace(r.ContactEmail)
	r.ContactMessage = strings.TrimSpace(r.ContactMessage)
	r.ContactPhone = strings.TrimSpace(r.ContactPhone)

	// Validate contact name
	if r.ContactName == "" {
		return fmt.Errorf("contact name is required")
	}
	if len(r.ContactName) > 200 {
		return fmt.Errorf("contact name must be less than 200 characters")
	}

	// Validate contact email
	if r.ContactEmail == "" {
		return fmt.Errorf("contact email is required")
	}
	if len(r.ContactEmail) > 200 {
		return fmt.Errorf("contact email must be less than 200 characters")
	}
	if !validate.IsValidEmail(r.ContactEmail) {
		return fmt.Errorf("contact email is invalid")
	}

	// Validate contact message
	if r.ContactMessage == "" {
		return fmt.Errorf("contact message is required")
	}
	if len(r.ContactMessage) > 200 {
		return fmt.Errorf("contact message must be less than 200 characters")
	}

	// Validate contact phone
	if !validate.IsValidPhoneNumber(r.ContactPhone) {
		return fmt.Errorf("contact phone is invalid")
	}

	return nil
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
