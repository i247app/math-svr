package program

import (
	domain "math-ai.com/math-ai/internal/domain/program"
	"math-ai.com/math-ai/internal/shared/pagination"
)

type ProgramResponse struct {
	ID           int64   `json:"id"`
	ProgramID    int64   `json:"program_id"`
	Label        string  `json:"label"`
	Description  string  `json:"description"`
	ImageKey     *string `json:"image_key,omitempty"`
	ImageUrl     *string `json:"image_url"` // pre-signed url from image_key
	DisplayOrder int8    `json:"display_order"`
	Note         *string `json:"note,omitempty"`
	CreateDt     string  `json:"create_dt"`
	ModifyDt     string  `json:"modify_dt"`
}

type CreateProgramReq struct {
	Label        string  `json:"label"`
	Description  string  `json:"description"`
	ImageKey     *string `json:"image_key,omitempty"`
	DisplayOrder int8    `json:"display_order"`
	Note         *string `json:"note,omitempty"`
}

type CreateProgramRes struct {
	Program *ProgramResponse `json:"program"`
}

// UpdateProgramReq uses pointer fields for every patchable column so a
// client can omit fields they don't intend to change.
type UpdateProgramReq struct {
	ProgramID    int64   `json:"program_id"`
	Label        *string `json:"label,omitempty"`
	Description  *string `json:"description,omitempty"`
	ImageKey     *string `json:"image_key,omitempty"`
	DisplayOrder *int8   `json:"display_order,omitempty"`
	Note         *string `json:"note,omitempty"`
}

type UpdateProgramRes struct {
	Program *ProgramResponse `json:"program"`
}

type DeleteProgramReq struct {
	ProgramID int64 `json:"program_id"`
}

type DeleteProgramRes struct{}

type GetProgramReq struct {
	ProgramID int64 `json:"program_id"`
}

type GetProgramRes struct {
	Program *ProgramResponse `json:"program"`
}

type ListProgramsReq struct {
	Page int64 `json:"page"`
	Size int64 `json:"size"`
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
		Note:         p.Note(),
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
