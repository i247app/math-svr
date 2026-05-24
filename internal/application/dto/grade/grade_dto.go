package grade

import (
	domain "math-ai.com/math-ai/internal/domain/grade"
	"math-ai.com/math-ai/internal/shared/enum"
	"math-ai.com/math-ai/internal/shared/pagination"
)

type GradeResponse struct {
	ID           int64   `json:"id"`
	GradeID      string  `json:"grade_id"`
	Label        string  `json:"label"`
	Description  string  `json:"description"`
	ImageKey     *string `json:"image_key,omitempty"`
	ImageUrl     *string `json:"image_url"` // pre-signed url from image_key
	DisplayOrder int8    `json:"display_order"`
	CreateDt     string  `json:"create_dt"`
	ModifyDt     string  `json:"modify_dt"`
}

type ListGradesReq struct {
	Language enum.LanguageType `json:"language,omitempty"`
	Page     int64             `json:"page"`
	Size     int64             `json:"size"`
}

type ListGradesRes struct {
	Grades     []*GradeResponse       `json:"grades"`
	Pagination *pagination.Pagination `json:"pagination"`
}

func DomainToResponse(g *domain.Grade) *GradeResponse {
	if g == nil {
		return nil
	}

	return &GradeResponse{
		ID:           g.Id(),
		GradeID:      g.GradeId().String(),
		Label:        g.Label(),
		Description:  g.Description(),
		ImageKey:     g.ImageKey(),
		DisplayOrder: g.DisplayOrder(),
		CreateDt:     g.CreateDt().String(),
		ModifyDt:     g.ModifyDt().String(),
	}
}

func DomainListToResponse(grades []*domain.Grade) []*GradeResponse {
	result := make([]*GradeResponse, len(grades))
	for i, g := range grades {
		result[i] = DomainToResponse(g)
	}
	return result
}
