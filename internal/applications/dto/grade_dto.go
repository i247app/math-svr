package dto

import (
	"io"
	"time"

	domain "math-ai.com/math-ai/internal/core/domain/grade"
	"math-ai.com/math-ai/internal/shared/constant/enum"
	"math-ai.com/math-ai/internal/shared/utils/pagination"
)

type GradeResponse struct {
	ID           string    `json:"id"`
	Label        string    `json:"label"`
	Description  string    `json:"description"`
	IconURL      *string   `json:"image_key"`
	Status       string    `json:"status"`
	DisplayOrder int8      `json:"display_order"`
	CreatedAt    time.Time `json:"created_at"`
	ModifiedAt   time.Time `json:"modified_at"`
}

type GetGradeResponse struct {
	Grade *GradeResponse `json:"grade"`
}

type ListGradeRequest struct {
	Search    string `json:"search,omitempty" form:"search"`
	Page      int64  `json:"page" form:"page"`
	Limit     int64  `json:"size" form:"size"`
	OrderBy   string `json:"order_by" form:"order_by"`
	OrderDesc bool   `json:"order_desc" form:"order_desc"`
	TakeAll   bool   `json:"take_all" form:"take_all"`
}

type ListGradeResponse struct {
	Items      []*GradeResponse       `json:"items"`
	Pagination *pagination.Pagination `json:"metadata"`
}

type CreateGradeRequest struct {
	Label        string  `json:"label"`
	Description  string  `json:"description"`
	ImageKey     *string `json:"image_key,omitempty"`
	DisplayOrder int8    `json:"display_order"`

	// image upload fields (for multipart form)
	IconFile        io.Reader `json:"icon_file"`         // File reader
	IconFilename    string    `json:"icon_file_name"`    // Original filename
	IconContentType string    `json:"icon_content_type"` // MIME type
}

type CreateGradeResponse struct {
	Grade *GradeResponse `json:"grade"`
}

type UpdateGradeRequest struct {
	ID           string        `json:"id"`
	Label        *string       `json:"label,omitempty"`
	Description  *string       `json:"description,omitempty"`
	ImageKey     *string       `json:"image_key,omitempty"`
	Status       *enum.EStatus `json:"status,omitempty"`
	DisplayOrder *int8         `json:"display_order,omitempty"`
}

type UpdateGradeResponse struct {
	Grade *GradeResponse `json:"grade"`
}

type DeleteGradeRequest struct {
	ID string `json:"id"`
}

func BuildGradeDomainForCreate(req *CreateGradeRequest) *domain.Grade {
	gradeDomain := domain.NewGradeDomain()
	gradeDomain.GenerateID()
	gradeDomain.SetLabel(req.Label)
	gradeDomain.SetDescription(req.Description)
	gradeDomain.SetImageKey(req.ImageKey)
	gradeDomain.SetStatus(string(enum.StatusActive))
	gradeDomain.SetDisplayOrder(req.DisplayOrder)

	return gradeDomain
}

func BuildGradeDomainForUpdate(req *UpdateGradeRequest) *domain.Grade {
	gradeDomain := domain.NewGradeDomain()
	gradeDomain.SetID(req.ID)

	if req.Label != nil {
		gradeDomain.SetLabel(*req.Label)
	}

	if req.Description != nil {
		gradeDomain.SetDescription(*req.Description)
	}

	if req.ImageKey != nil {
		gradeDomain.SetImageKey(req.ImageKey)
	}

	if req.Status != nil {
		gradeDomain.SetStatus(string(*req.Status))
	}

	if req.DisplayOrder != nil {
		gradeDomain.SetDisplayOrder(*req.DisplayOrder)
	}

	return gradeDomain
}

func GradeResponseFromDomain(g *domain.Grade) GradeResponse {
	return GradeResponse{
		ID:           g.ID(),
		Label:        g.Label(),
		Description:  g.Description(),
		IconURL:      g.ImageKey(),
		Status:       g.Status(),
		DisplayOrder: g.DisplayOrder(),
		CreatedAt:    g.CreatedAt(),
		ModifiedAt:   g.ModifiedAt(),
	}
}

func GradeResponseListFromDomain(grades []*domain.Grade) []GradeResponse {
	responses := make([]GradeResponse, len(grades))
	for i, g := range grades {
		responses[i] = GradeResponseFromDomain(g)
	}
	return responses
}
