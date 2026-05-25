package program

import (
	domain "math-ai.com/math-ai/internal/domain/program"
	"math-ai.com/math-ai/internal/shared/enum"
	"math-ai.com/math-ai/internal/shared/pagination"
)

type ProgramResponse struct {
	ID           int64   `json:"id"`
	ProgramID    string  `json:"program_id"`
	Label        string  `json:"label"`
	Description  string  `json:"description"`
	ImageKey     *string `json:"image_key,omitempty"`
	ImageUrl     *string `json:"image_url"` // pre-signed url from image_key
	DisplayOrder int8    `json:"display_order"`
	CreateDt     string  `json:"create_dt"`
	ModifyDt     string  `json:"modify_dt"`
}

type ListProgramsReq struct {
	Language enum.LanguageType `json:"language,omitempty"`
	Page     int64             `json:"page"`
	Size     int64             `json:"size"`
}

type ListProgramsRes struct {
	Programs   []*ProgramResponse     `json:"programs"`
	Pagination *pagination.Pagination `json:"pagination"`
}

func DomainToResponse(p *domain.Program) *ProgramResponse {
	if p == nil {
		return nil
	}

	return &ProgramResponse{
		ID:           p.Id(),
		ProgramID:    p.ProgramId(),
		Label:        p.Label(),
		Description:  p.Description(),
		ImageKey:     p.ImageKey(),
		DisplayOrder: p.DisplayOrder(),
		CreateDt:     p.CreateDt().String(),
		ModifyDt:     p.ModifyDt().String(),
	}
}

func DomainListToResponse(programs []*domain.Program) []*ProgramResponse {
	result := make([]*ProgramResponse, len(programs))
	for i, p := range programs {
		result[i] = DomainToResponse(p)
	}
	return result
}
