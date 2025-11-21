package dto

import (
	"time"

	domain "math-ai.com/math-ai/internal/core/domain/grade"
	"math-ai.com/math-ai/internal/shared/constant/enum"
	"math-ai.com/math-ai/internal/shared/utils/pagination"
)

type GradeResponse struct {
	ID          string    `json:"id"`
	Label       string    `json:"label"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	ModifiedAt  time.Time `json:"modified_at"`
}

type GetGradeResponse struct {
	Grade *GradeResponse `json:"result"`
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
	Items      []*GradeResponse       `json:"result"`
	Pagination *pagination.Pagination `json:"metadata"`
}

type CreateGradeRequest struct {
	Label       string `json:"label"`
	Description string `json:"description"`
}

type CreateGradeResponse struct {
	Grade *GradeResponse `json:"result"`
}

type UpdateGradeRequest struct {
	ID          string       `json:"id"`
	Label       *string      `json:"label,omitempty"`
	Description *string      `json:"description,omitempty"`
	Status      *enum.EStatus `json:"status,omitempty"`
}

type UpdateGradeResponse struct {
	Grade *GradeResponse `json:"result"`
}

type DeleteGradeRequest struct {
	ID string `json:"id"`
}

func BuildGradeDomainForCreate(dto *CreateGradeRequest) *domain.Grade {
	gradeDomain := domain.NewGradeDomain()
	gradeDomain.GenerateID()
	gradeDomain.SetLabel(dto.Label)
	gradeDomain.SetDescription(dto.Description)
	gradeDomain.SetStatus(string(enum.StatusActive))

	return gradeDomain
}

func BuildGradeDomainForUpdate(dto *UpdateGradeRequest) *domain.Grade {
	gradeDomain := domain.NewGradeDomain()
	gradeDomain.SetID(dto.ID)

	if dto.Label != nil {
		gradeDomain.SetLabel(*dto.Label)
	}

	if dto.Description != nil {
		gradeDomain.SetDescription(*dto.Description)
	}

	if dto.Status != nil {
		gradeDomain.SetStatus(string(*dto.Status))
	}

	return gradeDomain
}

func GradeResponseFromDomain(g *domain.Grade) GradeResponse {
	return GradeResponse{
		ID:          g.ID(),
		Label:       g.Label(),
		Description: g.Description(),
		Status:      g.Status(),
		CreatedAt:   g.CreatedAt(),
		ModifiedAt:  g.ModifiedAt(),
	}
}

func GradeResponseListFromDomain(grades []*domain.Grade) []GradeResponse {
	responses := make([]GradeResponse, len(grades))
	for i, g := range grades {
		responses[i] = GradeResponseFromDomain(g)
	}
	return responses
}
