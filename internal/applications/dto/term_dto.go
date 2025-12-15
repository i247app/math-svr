package dto

import (
	"io"
	"time"

	domain "math-ai.com/math-ai/internal/core/domain/term"
	"math-ai.com/math-ai/internal/shared/constant/enum"
	"math-ai.com/math-ai/internal/shared/utils/pagination"
)

type TermResponse struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Description  *string   `json:"description"`
	ImageUrl     *string   `json:"image_url"`
	Status       string    `json:"status"`
	DisplayOrder int8      `json:"display_order"`
	CreatedAt    time.Time `json:"created_at"`
	ModifiedAt   time.Time `json:"modified_at"`
}

type GetTermResponse struct {
	Term *TermResponse `json:"term"`
}

type ListTermRequest struct {
	Search    string `json:"search,omitempty" form:"search"`
	Page      int64  `json:"page" form:"page"`
	Limit     int64  `json:"size" form:"size"`
	OrderBy   string `json:"order_by" form:"order_by"`
	OrderDesc bool   `json:"order_desc" form:"order_desc"`
	TakeAll   bool   `json:"take_all" form:"take_all"`
}

type ListTermResponse struct {
	Items      []*TermResponse        `json:"items"`
	Pagination *pagination.Pagination `json:"metadata"`
}

type CreateTermRequest struct {
	Name         string  `json:"name"`
	Description  *string `json:"description"`
	ImageKey     *string `json:"image_key,omitempty"`
	DisplayOrder int8    `json:"display_order"`

	// image upload fields (for multipart form)
	ImageFile        io.Reader `json:"icon_file"`         // File reader
	ImageFilename    string    `json:"icon_file_name"`    // Original filename
	ImageContentType string    `json:"icon_content_type"` // MIME type
}

type CreateTermResponse struct {
	Term *TermResponse `json:"term"`
}

type UpdateTermRequest struct {
	ID           string        `json:"id"`
	Name         *string       `json:"name,omitempty"`
	Description  *string       `json:"description,omitempty"`
	Status       *enum.EStatus `json:"status,omitempty"`
	DisplayOrder *int8         `json:"display_order,omitempty"`

	// image upload fields (for multipart form)
	ImageFile        io.Reader `json:"icon_file"`         // File reader
	ImageFilename    string    `json:"icon_file_name"`    // Original filename
	ImageContentType string    `json:"icon_content_type"` // MIME type
}

type UpdateTermResponse struct {
	Term *TermResponse `json:"term"`
}

type DeleteTermRequest struct {
	ID string `json:"id"`
}

func BuildTermDomainForCreate(req *CreateTermRequest) *domain.Term {
	termDomain := domain.NewTermDomain()
	termDomain.GenerateID()
	termDomain.SetName(req.Name)
	termDomain.SetDescription(req.Description)
	termDomain.SetImageKey(req.ImageKey)
	termDomain.SetStatus(string(enum.StatusActive))
	termDomain.SetDisplayOrder(req.DisplayOrder)

	return termDomain
}

func BuildTermDomainForUpdate(req *UpdateTermRequest) *domain.Term {
	termDomain := domain.NewTermDomain()
	termDomain.SetID(req.ID)

	if req.Name != nil {
		termDomain.SetName(*req.Name)
	}

	if req.Description != nil {
		termDomain.SetDescription(req.Description)
	}

	if req.Status != nil {
		termDomain.SetStatus(string(*req.Status))
	}

	if req.DisplayOrder != nil {
		termDomain.SetDisplayOrder(*req.DisplayOrder)
	}

	return termDomain
}

func TermResponseFromDomain(s *domain.Term) TermResponse {
	return TermResponse{
		ID:           s.ID(),
		Name:         s.Name(),
		Description:  s.Description(),
		Status:       s.Status(),
		DisplayOrder: s.DisplayOrder(),
		CreatedAt:    s.CreatedAt(),
		ModifiedAt:   s.ModifiedAt(),
	}
}

func TermResponseListFromDomain(terms []*domain.Term) []TermResponse {
	responses := make([]TermResponse, len(terms))
	for i, s := range terms {
		responses[i] = TermResponseFromDomain(s)
	}
	return responses
}
