package program

import (
	domain "math-ai.com/math-ai/internal/domain/program"
	"math-ai.com/math-ai/internal/shared/enum"
	"math-ai.com/math-ai/internal/shared/pagination"
)

// ProgramTranslationDTO is the per-language override surface exposed by
// the API. ProgramTranslationID is empty on create payloads and populated
// on responses; the service treats (language) as the upsert key.
type ProgramTranslationDTO struct {
	ProgramTranslationID int64   `json:"program_translation_id,omitempty"`
	ProgramID            int64   `json:"program_id,omitempty"`
	Language             string  `json:"language"`
	Label                string  `json:"label"`
	Description          string  `json:"description"`
	Note                 *string `json:"note,omitempty"`
}

type ProgramResponse struct {
	ID           int64                    `json:"id"`
	ProgramID    int64                    `json:"program_id"`
	Label        string                   `json:"label"`
	Description  string                   `json:"description"`
	ImageKey     *string                  `json:"image_key,omitempty"`
	ImageUrl     *string                  `json:"image_url"` // pre-signed url from image_key
	DisplayOrder int8                     `json:"display_order"`
	Note         *string                  `json:"note,omitempty"`
	Translations []*ProgramTranslationDTO `json:"translations,omitempty"`
	CreateDt     string                   `json:"create_dt"`
	ModifyDt     string                   `json:"modify_dt"`
}

type CreateProgramReq struct {
	Label        string                   `json:"label"`
	Description  string                   `json:"description"`
	ImageKey     *string                  `json:"image_key,omitempty"`
	DisplayOrder int8                     `json:"display_order"`
	Note         *string                  `json:"note,omitempty"`
	Translations []*ProgramTranslationDTO `json:"translations,omitempty"`
}

type CreateProgramRes struct {
	Program *ProgramResponse `json:"program"`
}

// UpdateProgramReq uses pointer fields for every patchable column so a
// client can omit fields they don't intend to change. Translations is a
// full upsert payload: rows with a matching (program_id, language) are
// updated, missing rows are inserted, and rows the client omits are left
// alone (no implicit deletion — translation removal is an explicit call).
type UpdateProgramReq struct {
	ProgramID    int64                    `json:"program_id"`
	Label        *string                  `json:"label,omitempty"`
	Description  *string                  `json:"description,omitempty"`
	ImageKey     *string                  `json:"image_key,omitempty"`
	DisplayOrder *int8                    `json:"display_order,omitempty"`
	Note         *string                  `json:"note,omitempty"`
	Translations []*ProgramTranslationDTO `json:"translations,omitempty"`
}

type UpdateProgramRes struct {
	Program *ProgramResponse `json:"program"`
}

type DeleteProgramReq struct {
	ProgramID int64 `json:"program_id"`
}

type DeleteProgramRes struct{}

type GetProgramReq struct {
	ProgramID int64             `json:"program_id"`
	Language  enum.LanguageType `json:"language,omitempty"`
}

type GetProgramRes struct {
	Program *ProgramResponse `json:"program"`
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
	resp := &ProgramResponse{
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
	if ts := p.Translations(); len(ts) > 0 {
		resp.Translations = make([]*ProgramTranslationDTO, len(ts))
		for i, t := range ts {
			resp.Translations[i] = TranslationDomainToDTO(t)
		}
	}
	return resp
}

func DomainListToResponse(programs []*domain.Program) []*ProgramResponse {
	result := make([]*ProgramResponse, len(programs))
	for i, p := range programs {
		result[i] = DomainToResponse(p)
	}
	return result
}

func TranslationDomainToDTO(t *domain.ProgramTranslation) *ProgramTranslationDTO {
	if t == nil {
		return nil
	}
	return &ProgramTranslationDTO{
		ProgramTranslationID: t.ProgramTranslationId(),
		ProgramID:            t.ProgramId(),
		Language:             t.Language(),
		Label:                t.Label(),
		Description:          t.Description(),
		Note:                 t.Note(),
	}
}
