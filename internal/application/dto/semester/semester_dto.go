package semester

import (
	domain "math-ai.com/math-ai/internal/domain/semester"
	"math-ai.com/math-ai/internal/shared/pagination"
)

type SemesterResponse struct {
	ID           int64   `json:"id"`
	SemesterID   int64   `json:"semester_id"`
	Name         string  `json:"name"`
	Description  string  `json:"description"`
	ImageKey     *string `json:"image_key,omitempty"`
	ImageUrl     *string `json:"image_url"` // pre-signed url from image_key
	DisplayOrder int8    `json:"display_order"`
	Note         *string `json:"note,omitempty"`
	CreateDt     string  `json:"create_dt"`
	ModifyDt     string  `json:"modify_dt"`
}

type CreateSemesterReq struct {
	Name         string  `json:"name"`
	Description  string  `json:"description"`
	ImageKey     *string `json:"image_key,omitempty"`
	DisplayOrder int8    `json:"display_order"`
	Note         *string `json:"note,omitempty"`
}

type CreateSemesterRes struct {
	Semester *SemesterResponse `json:"semester"`
}

// UpdateSemesterReq uses pointer fields for every patchable column so a
// client can omit fields they don't intend to change.
type UpdateSemesterReq struct {
	SemesterID   int64   `json:"semester_id"`
	Name         *string `json:"name,omitempty"`
	Description  *string `json:"description,omitempty"`
	ImageKey     *string `json:"image_key,omitempty"`
	DisplayOrder *int8   `json:"display_order,omitempty"`
	Note         *string `json:"note,omitempty"`
}

type UpdateSemesterRes struct {
	Semester *SemesterResponse `json:"semester"`
}

type DeleteSemesterReq struct {
	SemesterID int64 `json:"semester_id"`
}

type DeleteSemesterRes struct{}

type GetSemesterReq struct {
	SemesterID int64 `json:"semester_id"`
}

type GetSemesterRes struct {
	Semester *SemesterResponse `json:"semester"`
}

type ListSemestersReq struct {
	Page int64 `json:"page"`
	Size int64 `json:"size"`
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
		Note:         s.Note(),
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
