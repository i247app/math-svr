package grade

import (
	domain "math-ai.com/math-ai/internal/domain/grade"
	"math-ai.com/math-ai/internal/shared/enum"
	"math-ai.com/math-ai/internal/shared/pagination"
)

// GradeTranslationDTO is the per-language override surface exposed by
// the API. GradeTranslationID is empty on create payloads and populated
// on responses; the service treats (language) as the upsert key.
type GradeTranslationDTO struct {
	GradeTranslationID int64   `json:"grade_translation_id,omitempty"`
	GradeID            int64   `json:"grade_id,omitempty"`
	Language           string  `json:"language"`
	Label              string  `json:"label"`
	Description        string  `json:"description"`
	Note               *string `json:"note,omitempty"`
}

type GradeResponse struct {
	ID           int64                  `json:"id"`
	GradeID      int64                  `json:"grade_id"`
	Label        string                 `json:"label"`
	Description  string                 `json:"description"`
	ImageKey     *string                `json:"image_key,omitempty"`
	ImageUrl     *string                `json:"image_url"` // pre-signed url from image_key
	DisplayOrder int8                   `json:"display_order"`
	Note         *string                `json:"note,omitempty"`
	Translations []*GradeTranslationDTO `json:"translations,omitempty"`
	CreateDt     string                 `json:"create_dt"`
	ModifyDt     string                 `json:"modify_dt"`
}

type CreateGradeReq struct {
	Label        string                 `json:"label"`
	Description  string                 `json:"description"`
	ImageKey     *string                `json:"image_key,omitempty"`
	DisplayOrder int8                   `json:"display_order"`
	Note         *string                `json:"note,omitempty"`
	Translations []*GradeTranslationDTO `json:"translations,omitempty"`
}

type CreateGradeRes struct {
	Grade *GradeResponse `json:"grade"`
}

// UpdateGradeReq uses pointer fields for every patchable column so a
// client can omit fields they don't intend to change. Translations is a
// full upsert payload: rows with a matching (grade_id, language) are
// updated, missing rows are inserted, and rows the client omits are left
// alone (no implicit deletion — translation removal is an explicit call).
type UpdateGradeReq struct {
	GradeID      int64                  `json:"grade_id"`
	Label        *string                `json:"label,omitempty"`
	Description  *string                `json:"description,omitempty"`
	ImageKey     *string                `json:"image_key,omitempty"`
	DisplayOrder *int8                  `json:"display_order,omitempty"`
	Note         *string                `json:"note,omitempty"`
	Translations []*GradeTranslationDTO `json:"translations,omitempty"`
}

type UpdateGradeRes struct {
	Grade *GradeResponse `json:"grade"`
}

type DeleteGradeReq struct {
	GradeID int64 `json:"grade_id"`
}

type DeleteGradeRes struct{}

type GetGradeReq struct {
	GradeID  int64             `json:"grade_id"`
	Language enum.LanguageType `json:"language,omitempty"`
}

type GetGradeRes struct {
	Grade *GradeResponse `json:"grade"`
}

type ListGradesReq struct {
	Language enum.LanguageType `json:"language,omitempty"`
	GradeIDs []int64           `json:"grade_ids,omitempty"`
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
	resp := &GradeResponse{
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
	if ts := g.Translations(); len(ts) > 0 {
		resp.Translations = make([]*GradeTranslationDTO, len(ts))
		for i, t := range ts {
			resp.Translations[i] = TranslationDomainToDTO(t)
		}
	}
	return resp
}

func DomainListToResponse(grades []*domain.Grade) []*GradeResponse {
	result := make([]*GradeResponse, len(grades))
	for i, g := range grades {
		result[i] = DomainToResponse(g)
	}
	return result
}

func TranslationDomainToDTO(t *domain.GradeTranslation) *GradeTranslationDTO {
	if t == nil {
		return nil
	}
	return &GradeTranslationDTO{
		GradeTranslationID: t.GradeTranslationId(),
		GradeID:            t.GradeId(),
		Language:           t.Language(),
		Label:              t.Label(),
		Description:        t.Description(),
		Note:               t.Note(),
	}
}
