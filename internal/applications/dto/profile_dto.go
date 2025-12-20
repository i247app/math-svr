package dto

import (
	domain "math-ai.com/math-ai/internal/core/domain/profile"
	"math-ai.com/math-ai/internal/shared/constant/enum"
	"math-ai.com/math-ai/internal/shared/utils/time"
)

type ProfileResponse struct {
	ID               string        `json:"id"`
	UID              string        `json:"uid"`
	Name             string        `json:"name"`
	Email            string        `json:"email"`
	Phone            string        `json:"phone"`
	Age              *int          `json:"age"`
	Grade            string        `json:"grade"`
	Semester         string        `json:"semester"`
	AvatarPreviewURL *string       `json:"avatar_url"`
	Status           string        `json:"status"`
	CreatedAt        time.MathTime `json:"created_at"`
	ModifiedAt       time.MathTime `json:"modified_at"`
}

type FetchProfileRequest struct {
	UID string `json:"uid"`
}

type FetchProfileResponse struct {
	Profile *ProfileResponse `json:"profile"`
}

type CreateProfileRequest struct {
	UID        string `json:"uid"`
	GradeID    string `json:"grade_id"`
	SemesterID string `json:"semester_id"`
}

type CreateProfileResponse struct {
	Profile *ProfileResponse `json:"profile"`
}

type UpdateProfileRequest struct {
	UID        string        `json:"uid"`
	GradeID    *string       `json:"grade_id,omitempty"`
	SemesterID *string       `json:"semester_id,omitempty"`
	Status     *enum.EStatus `json:"status,omitempty"`
}

type UpdateProfileResponse struct {
	Profile *ProfileResponse `json:"profile"`
}

func BuildProfileDomainForCreate(req *CreateProfileRequest) *domain.Profile {
	profileDomain := domain.NewProfileDomain()
	profileDomain.GenerateID()
	profileDomain.SetUID(req.UID)
	profileDomain.SetGradeID(req.GradeID)
	profileDomain.SetSemesterID(req.SemesterID)
	profileDomain.SetStatus(string(enum.StatusActive))

	return profileDomain
}

func BuildProfileDomainForUpdate(req *UpdateProfileRequest) *domain.Profile {
	profileDomain := domain.NewProfileDomain()

	profileDomain.SetUID(req.UID)

	if req.GradeID != nil {
		profileDomain.SetGradeID(*req.GradeID)
	}

	if req.SemesterID != nil {
		profileDomain.SetSemesterID(*req.SemesterID)
	}

	if req.Status != nil {
		profileDomain.SetStatus(string(*req.Status))
	}

	return profileDomain
}

func ProfileResponseFromDomain(p *domain.Profile) ProfileResponse {
	var age int

	if p.Dob() != nil {
		age = time.Now().Year() - p.Dob().Year()
		if time.Now().YearDay() < p.Dob().YearDay() {
			age--
		}
	}

	return ProfileResponse{
		ID:         p.ID(),
		UID:        p.UID(),
		Name:       p.Name(),
		Email:      p.Email(),
		Phone:      p.Phone(),
		Age:        &age,
		Grade:      p.Grade(),
		Semester:   p.Semester(),
		Status:     p.Status(),
		CreatedAt:  p.CreatedAt(),
		ModifiedAt: p.ModifiedAt(),
	}
}
