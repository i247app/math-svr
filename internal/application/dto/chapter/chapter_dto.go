package chapter

import (
	domain "math-ai.com/math-ai/internal/domain/chapter"
	"math-ai.com/math-ai/internal/shared/enum"
	"math-ai.com/math-ai/internal/shared/pagination"
)

// ChapterTranslationDTO is the per-language override surface exposed by
// the API. ChapterTranslationID is empty on create payloads and populated
// on responses; the service treats (language) as the upsert key.
type ChapterTranslationDTO struct {
	ChapterTranslationID string  `json:"chapter_translation_id,omitempty"`
	ChapterID            string  `json:"chapter_id,omitempty"`
	Language             string  `json:"language"`
	Label                string  `json:"label"`
	Description          string  `json:"description"`
	Note                 *string `json:"note,omitempty"`
}

type ChapterResponse struct {
	ID           int64                    `json:"id"`
	ChapterID    string                   `json:"chapter_id"`
	ProgramID    string                   `json:"program_id"`
	GradeID      string                   `json:"grade_id"`
	SemesterID   string                   `json:"semester_id"`
	Label        string                   `json:"label"`
	Description  string                   `json:"description"`
	DisplayOrder int8                     `json:"display_order"`
	Note         *string                  `json:"note,omitempty"`
	Translations []*ChapterTranslationDTO `json:"translations,omitempty"`
	CreateDt     string                   `json:"create_dt"`
	ModifyDt     string                   `json:"modify_dt"`
}

type CreateChapterReq struct {
	ProgramID    string                   `json:"program_id"`
	GradeID      string                   `json:"grade_id"`
	SemesterID   string                   `json:"semester_id"`
	Label        string                   `json:"label"`
	Description  string                   `json:"description"`
	DisplayOrder int8                     `json:"display_order"`
	Note         *string                  `json:"note,omitempty"`
	Translations []*ChapterTranslationDTO `json:"translations,omitempty"`
}

type CreateChapterRes struct {
	Chapter *ChapterResponse `json:"chapter"`
}

// UpdateChapterReq uses pointer fields for every patchable column so a
// client can omit fields they don't intend to change. Translations is a
// full upsert payload: rows with a matching (chapter_id, language) are
// updated, missing rows are inserted, and rows the client omits are left
// alone (no implicit deletion — translation removal is an explicit call).
type UpdateChapterReq struct {
	ChapterID    string                   `json:"chapter_id"`
	ProgramID    *string                  `json:"program_id,omitempty"`
	GradeID      *string                  `json:"grade_id,omitempty"`
	SemesterID   *string                  `json:"semester_id,omitempty"`
	Label        *string                  `json:"label,omitempty"`
	Description  *string                  `json:"description,omitempty"`
	DisplayOrder *int8                    `json:"display_order,omitempty"`
	Note         *string                  `json:"note,omitempty"`
	Translations []*ChapterTranslationDTO `json:"translations,omitempty"`
}

type UpdateChapterRes struct {
	Chapter *ChapterResponse `json:"chapter"`
}

type DeleteChapterReq struct {
	ChapterID string `json:"chapter_id"`
}

type DeleteChapterRes struct{}

type GetChapterReq struct {
	ChapterID string            `json:"chapter_id"`
	Language  enum.LanguageType `json:"language,omitempty"`
}

type GetChapterRes struct {
	Chapter *ChapterResponse `json:"chapter"`
}

type ListChaptersReq struct {
	ProgramID  *string           `json:"program_id,omitempty"`
	GradeID    *string           `json:"grade_id,omitempty"`
	SemesterID *string           `json:"semester_id,omitempty"`
	Language   enum.LanguageType `json:"language,omitempty"`
	Page       int64             `json:"page"`
	Size       int64             `json:"size"`
}

type ListChaptersRes struct {
	Chapters   []*ChapterResponse     `json:"chapters"`
	Pagination *pagination.Pagination `json:"pagination"`
}

func DomainToResponse(c *domain.Chapter) *ChapterResponse {
	if c == nil {
		return nil
	}
	resp := &ChapterResponse{
		ID:           c.Id(),
		ChapterID:    c.ChapterId(),
		ProgramID:    c.ProgramId(),
		GradeID:      c.GradeId(),
		SemesterID:   c.SemesterId(),
		Label:        c.Label(),
		Description:  c.Description(),
		DisplayOrder: c.DisplayOrder(),
		Note:         c.Note(),
		CreateDt:     c.CreateDt().String(),
		ModifyDt:     c.ModifyDt().String(),
	}
	if ts := c.Translations(); len(ts) > 0 {
		resp.Translations = make([]*ChapterTranslationDTO, len(ts))
		for i, t := range ts {
			resp.Translations[i] = TranslationDomainToDTO(t)
		}
	}
	return resp
}

func DomainListToResponse(chapters []*domain.Chapter) []*ChapterResponse {
	result := make([]*ChapterResponse, len(chapters))
	for i, c := range chapters {
		result[i] = DomainToResponse(c)
	}
	return result
}

func TranslationDomainToDTO(t *domain.ChapterTranslation) *ChapterTranslationDTO {
	if t == nil {
		return nil
	}
	return &ChapterTranslationDTO{
		ChapterTranslationID: t.ChapterTranslationId(),
		ChapterID:            t.ChapterId(),
		Language:             t.Language(),
		Label:                t.Label(),
		Description:          t.Description(),
		Note:                 t.Note(),
	}
}
