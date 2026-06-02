package school

import (
	domain "math-ai.com/math-ai/internal/domain/school"
	"math-ai.com/math-ai/internal/shared/pagination"
)

// SchoolResponse is the school surface exposed to the API. ImageUrl is
// the short-lived presigned URL derived from ImageKey at the service
// edge; clients should prefer ImageUrl for display and ImageKey only
// when they need to re-sign later.
type SchoolResponse struct {
	ID           int64   `json:"id"`
	SchoolID     int64   `json:"school_id"`
	Name         string  `json:"name"`
	Description  *string `json:"description,omitempty"`
	ImageKey     *string `json:"image_key,omitempty"`
	ImageUrl     *string `json:"image_url"`
	District     *string `json:"district,omitempty"`
	Province     *string `json:"province,omitempty"`
	Note         *string `json:"note,omitempty"`
	SchoolStatus *string `json:"school_status,omitempty"`
	CreateDt     string  `json:"create_dt"`
	ModifyDt     string  `json:"modify_dt"`
}

type CreateSchoolReq struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	ImageKey    *string `json:"image_key,omitempty"`
	District    *string `json:"district,omitempty"`
	Province    *string `json:"province,omitempty"`
	Note        *string `json:"note,omitempty"`
}

type CreateSchoolRes struct {
	School *SchoolResponse `json:"school"`
}

// UpdateSchoolReq uses pointer fields for every patchable column so a
// client can omit fields they don't intend to change. A nil pointer
// means "leave alone"; a non-nil value (including an empty string for
// optional fields) overwrites.
type UpdateSchoolReq struct {
	SchoolID    int64   `json:"school_id"`
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	ImageKey    *string `json:"image_key,omitempty"`
	District    *string `json:"district,omitempty"`
	Province    *string `json:"province,omitempty"`
	Note        *string `json:"note,omitempty"`
}

type UpdateSchoolRes struct {
	School *SchoolResponse `json:"school"`
}

type DeleteSchoolReq struct {
	SchoolID int64 `json:"school_id"`
}

type DeleteSchoolRes struct{}

type GetSchoolReq struct {
	SchoolID int64 `json:"school_id"`
}

type GetSchoolRes struct {
	School *SchoolResponse `json:"school"`
}

type ListSchoolsReq struct {
	Search    *string `json:"search,omitempty"`
	District  *string `json:"district,omitempty"`
	Province  *string `json:"province,omitempty"`
	SchoolIDs []int64 `json:"school_ids,omitempty"`
	Page      int64   `json:"page"`
	Size      int64   `json:"size"`
}

type ListSchoolsRes struct {
	Schools    []*SchoolResponse      `json:"schools"`
	Pagination *pagination.Pagination `json:"pagination"`
}

func DomainToResponse(s *domain.School) *SchoolResponse {
	if s == nil {
		return nil
	}
	return &SchoolResponse{
		ID:           s.Id(),
		SchoolID:     s.SchoolId(),
		Name:         s.Name(),
		Description:  s.Description(),
		ImageKey:     s.ImageKey(),
		District:     s.District(),
		Province:     s.Province(),
		Note:         s.Note(),
		SchoolStatus: s.SchoolStatus(),
		CreateDt:     s.CreateDt().String(),
		ModifyDt:     s.ModifyDt().String(),
	}
}

func DomainListToResponse(schools []*domain.School) []*SchoolResponse {
	result := make([]*SchoolResponse, len(schools))
	for i, s := range schools {
		result[i] = DomainToResponse(s)
	}
	return result
}
