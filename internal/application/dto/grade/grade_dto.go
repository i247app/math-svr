package grade

import (
	domain "math-ai.com/math-ai/internal/domain/grade"
	"math-ai.com/math-ai/internal/shared/pagination"
)

type GradeResponse struct {
	ID           int64   `json:"id"`
	GradeID      int64   `json:"grade_id"`
	Label        string  `json:"label"`
	Description  string  `json:"description"`
	ImageKey     *string `json:"image_key,omitempty"`
	ImageUrl     *string `json:"image_url"` // pre-signed url from image_key
	DisplayOrder int8    `json:"display_order"`
	Note         *string `json:"note,omitempty"`
	CreateDt     string  `json:"create_dt"`
	ModifyDt     string  `json:"modify_dt"`
}

type CreateGradeReq struct {
	Label        string  `json:"label"`
	Description  string  `json:"description"`
	ImageKey     *string `json:"image_key,omitempty"`
	DisplayOrder int8    `json:"display_order"`
	Note         *string `json:"note,omitempty"`
}

type CreateGradeRes struct {
	Grade *GradeResponse `json:"grade"`
}

// UpdateGradeReq uses pointer fields for every patchable column so a
// client can omit fields they don't intend to change.
type UpdateGradeReq struct {
	GradeID      int64   `json:"grade_id"`
	Label        *string `json:"label,omitempty"`
	Description  *string `json:"description,omitempty"`
	ImageKey     *string `json:"image_key,omitempty"`
	DisplayOrder *int8   `json:"display_order,omitempty"`
	Note         *string `json:"note,omitempty"`
}

type UpdateGradeRes struct {
	Grade *GradeResponse `json:"grade"`
}

type DeleteGradeReq struct {
	GradeID int64 `json:"grade_id"`
}

type DeleteGradeRes struct{}

type GetGradeReq struct {
	GradeID int64 `json:"grade_id"`
}

type GetGradeRes struct {
	Grade *GradeResponse `json:"grade"`
}

type ListGradesReq struct {
	GradeIDs []int64 `json:"grade_ids,omitempty"`
	Page     int64   `json:"page"`
	Size     int64   `json:"size"`
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
		GradeID:      g.GradeId(),
		Label:        g.Label(),
		Description:  g.Description(),
		ImageKey:     g.ImageKey(),
		DisplayOrder: g.DisplayOrder(),
		Note:         g.Note(),
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
