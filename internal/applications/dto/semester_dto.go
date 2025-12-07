package dto

import (
	"io"
	"time"

	domain "math-ai.com/math-ai/internal/core/domain/semester"
	"math-ai.com/math-ai/internal/shared/constant/enum"
	"math-ai.com/math-ai/internal/shared/utils/pagination"
)

type SemesterResponse struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Description  *string   `json:"description"`
	ImageUrl     *string   `json:"image_url"`
	Status       string    `json:"status"`
	DisplayOrder int8      `json:"display_order"`
	CreatedAt    time.Time `json:"created_at"`
	ModifiedAt   time.Time `json:"modified_at"`
}

type GetSemesterResponse struct {
	Semester *SemesterResponse `json:"semester"`
}

type ListSemesterRequest struct {
	Search    string `json:"search,omitempty" form:"search"`
	Page      int64  `json:"page" form:"page"`
	Limit     int64  `json:"size" form:"size"`
	OrderBy   string `json:"order_by" form:"order_by"`
	OrderDesc bool   `json:"order_desc" form:"order_desc"`
	TakeAll   bool   `json:"take_all" form:"take_all"`
}

type ListSemesterResponse struct {
	Items      []*SemesterResponse    `json:"items"`
	Pagination *pagination.Pagination `json:"metadata"`
}

type CreateSemesterRequest struct {
	Name         string  `json:"name"`
	Description  *string `json:"description"`
	ImageKey     *string `json:"image_key,omitempty"`
	DisplayOrder int8    `json:"display_order"`

	// image upload fields (for multipart form)
	IconFile        io.Reader `json:"icon_file"`         // File reader
	IconFilename    string    `json:"icon_file_name"`    // Original filename
	IconContentType string    `json:"icon_content_type"` // MIME type
}

type CreateSemesterResponse struct {
	Semester *SemesterResponse `json:"semester"`
}

type UpdateSemesterRequest struct {
	ID           string        `json:"id"`
	Name         *string       `json:"name,omitempty"`
	Description  *string       `json:"description,omitempty"`
	ImageKey     *string       `json:"image_key,omitempty"`
	Status       *enum.EStatus `json:"status,omitempty"`
	DisplayOrder *int8         `json:"display_order,omitempty"`
}

type UpdateSemesterResponse struct {
	Semester *SemesterResponse `json:"semester"`
}

type DeleteSemesterRequest struct {
	ID string `json:"id"`
}

func BuildSemesterDomainForCreate(req *CreateSemesterRequest) *domain.Semester {
	semesterDomain := domain.NewSemesterDomain()
	semesterDomain.GenerateID()
	semesterDomain.SetName(req.Name)
	semesterDomain.SetDescription(req.Description)
	semesterDomain.SetImageKey(req.ImageKey)
	semesterDomain.SetStatus(string(enum.StatusActive))
	semesterDomain.SetDisplayOrder(req.DisplayOrder)

	return semesterDomain
}

func BuildSemesterDomainForUpdate(req *UpdateSemesterRequest) *domain.Semester {
	semesterDomain := domain.NewSemesterDomain()
	semesterDomain.SetID(req.ID)

	if req.Name != nil {
		semesterDomain.SetName(*req.Name)
	}

	if req.Description != nil {
		semesterDomain.SetDescription(req.Description)
	}

	if req.ImageKey != nil {
		semesterDomain.SetImageKey(req.ImageKey)
	}

	if req.Status != nil {
		semesterDomain.SetStatus(string(*req.Status))
	}

	if req.DisplayOrder != nil {
		semesterDomain.SetDisplayOrder(*req.DisplayOrder)
	}

	return semesterDomain
}

func SemesterResponseFromDomain(s *domain.Semester) SemesterResponse {
	return SemesterResponse{
		ID:           s.ID(),
		Name:         s.Name(),
		Description:  s.Description(),
		Status:       s.Status(),
		DisplayOrder: s.DisplayOrder(),
		CreatedAt:    s.CreatedAt(),
		ModifiedAt:   s.ModifiedAt(),
	}
}

func SemesterResponseListFromDomain(semesters []*domain.Semester) []SemesterResponse {
	responses := make([]SemesterResponse, len(semesters))
	for i, s := range semesters {
		responses[i] = SemesterResponseFromDomain(s)
	}
	return responses
}
