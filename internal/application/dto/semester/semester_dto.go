package semester

import (
	"github.com/google/uuid"
	domain "math-ai.com/math-ai/internal/domain/semester"
	"math-ai.com/math-ai/internal/shared/enum"
	"math-ai.com/math-ai/internal/shared/pagination"
)

type SemesterResponse struct {
	ID           int64     `json:"id"`
	SemesterID   uuid.UUID `json:"semester_id"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	ImageKey     *string   `json:"image_key,omitempty"`
	ImageUrl     *string   `json:"image_url"` // pre-signed url from image_key
	DisplayOrder int8      `json:"display_order"`
	CreateDt     string    `json:"create_dt"`
	ModifyDt     string    `json:"modify_dt"`
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

	return &SemesterResponse{
		ID:           s.Id(),
		SemesterID:   s.SemesterId(),
		Name:         s.Name(),
		Description:  s.Description(),
		ImageKey:     s.ImageKey(),
		DisplayOrder: s.DisplayOrder(),
		CreateDt:     s.CreateDt().String(),
		ModifyDt:     s.ModifyDt().String(),
	}
}

func DomainListToResponse(semesters []*domain.Semester) []*SemesterResponse {
	result := make([]*SemesterResponse, len(semesters))
	for i, s := range semesters {
		result[i] = DomainToResponse(s)
	}
	return result
}
