package semester

import (
	domain "math-ai.com/math-ai/internal/domain/semester"
	"math-ai.com/math-ai/internal/shared/enum"
	"math-ai.com/math-ai/internal/shared/pagination"
)

// SemesterTranslationDTO is the per-language override surface exposed by
// the API. SemesterTranslationID is empty on create payloads and populated
// on responses; the service treats (language) as the upsert key.
type SemesterTranslationDTO struct {
	SemesterTranslationID string  `json:"semester_translation_id,omitempty"`
	SemesterID            string  `json:"semester_id,omitempty"`
	Language              string  `json:"language"`
	Name                  string  `json:"name"`
	Description           string  `json:"description"`
	Note                  *string `json:"note,omitempty"`
}

type SemesterResponse struct {
	ID           int64                     `json:"id"`
	SemesterID   string                    `json:"semester_id"`
	Name         string                    `json:"name"`
	Description  string                    `json:"description"`
	ImageKey     *string                   `json:"image_key,omitempty"`
	ImageUrl     *string                   `json:"image_url"` // pre-signed url from image_key
	DisplayOrder int8                      `json:"display_order"`
	Note         *string                   `json:"note,omitempty"`
	Translations []*SemesterTranslationDTO `json:"translations,omitempty"`
	CreateDt     string                    `json:"create_dt"`
	ModifyDt     string                    `json:"modify_dt"`
}

type CreateSemesterReq struct {
	Name         string                    `json:"name"`
	Description  string                    `json:"description"`
	ImageKey     *string                   `json:"image_key,omitempty"`
	DisplayOrder int8                      `json:"display_order"`
	Note         *string                   `json:"note,omitempty"`
	Translations []*SemesterTranslationDTO `json:"translations,omitempty"`
}

type CreateSemesterRes struct {
	Semester *SemesterResponse `json:"semester"`
}

// UpdateSemesterReq uses pointer fields for every patchable column so a
// client can omit fields they don't intend to change. Translations is a
// full upsert payload: rows with a matching (semester_id, language) are
// updated, missing rows are inserted, and rows the client omits are left
// alone (no implicit deletion — translation removal is an explicit call).
type UpdateSemesterReq struct {
	SemesterID   string                    `json:"semester_id"`
	Name         *string                   `json:"name,omitempty"`
	Description  *string                   `json:"description,omitempty"`
	ImageKey     *string                   `json:"image_key,omitempty"`
	DisplayOrder *int8                     `json:"display_order,omitempty"`
	Note         *string                   `json:"note,omitempty"`
	Translations []*SemesterTranslationDTO `json:"translations,omitempty"`
}

type UpdateSemesterRes struct {
	Semester *SemesterResponse `json:"semester"`
}

type DeleteSemesterReq struct {
	SemesterID string `json:"semester_id"`
}

type DeleteSemesterRes struct{}

type GetSemesterReq struct {
	SemesterID string            `json:"semester_id"`
	Language   enum.LanguageType `json:"language,omitempty"`
}

type GetSemesterRes struct {
	Semester *SemesterResponse `json:"semester"`
}

type ListSemestersReq struct {
	Language enum.LanguageType `json:"language,omitempty"`
	Page     int64             `json:"page"`
	Size     int64             `json:"size"`
}

type ListSemestersRes struct {
	Semesters  []*SemesterResponse    `json:"semesters"`
	Pagination *pagination.Pagination `json:"pagination"`
}

func DomainToResponse(s *domain.Semester) *SemesterResponse {
	if s == nil {
		return nil
	}
	resp := &SemesterResponse{
		ID:           s.Id(),
		SemesterID:   s.SemesterId(),
		Name:         s.Name(),
		Description:  s.Description(),
		ImageKey:     s.ImageKey(),
		DisplayOrder: s.DisplayOrder(),
		Note:         s.Note(),
		CreateDt:     s.CreateDt().String(),
		ModifyDt:     s.ModifyDt().String(),
	}
	if ts := s.Translations(); len(ts) > 0 {
		resp.Translations = make([]*SemesterTranslationDTO, len(ts))
		for i, t := range ts {
			resp.Translations[i] = TranslationDomainToDTO(t)
		}
	}
	return resp
}

func DomainListToResponse(semesters []*domain.Semester) []*SemesterResponse {
	result := make([]*SemesterResponse, len(semesters))
	for i, s := range semesters {
		result[i] = DomainToResponse(s)
	}
	return result
}

func TranslationDomainToDTO(t *domain.SemesterTranslation) *SemesterTranslationDTO {
	if t == nil {
		return nil
	}
	return &SemesterTranslationDTO{
		SemesterTranslationID: t.SemesterTranslationId(),
		SemesterID:            t.SemesterId(),
		Language:              t.Language(),
		Name:                  t.Name(),
		Description:           t.Description(),
		Note:                  t.Note(),
	}
}
