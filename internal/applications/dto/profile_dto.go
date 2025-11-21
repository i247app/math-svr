package dto

import (
	"time"

	domain "math-ai.com/math-ai/internal/core/domain/profile"
	"math-ai.com/math-ai/internal/shared/constant/enum"
)

type ProfileResponse struct {
	ID         string    `json:"id"`
	UID        string    `json:"uid"`
	Name       string    `json:"name"`
	Email      string    `json:"email"`
	Phone      string    `json:"phone"`
	Grade      string    `json:"grade"`
	Level      string    `json:"level"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	ModifiedAt time.Time `json:"modified_at"`
}

type FetchProfileRequest struct {
	UID string `json:"uid"`
}

type FetchProfileResponse struct {
	Profile *ProfileResponse `json:"result"`
}

type CreateProfileRequest struct {
	UID   string `json:"uid"`
	Grade string `json:"grade"`
	Level string `json:"level"`
}

type CreateProfileResponse struct {
	Profile *ProfileResponse `json:"result"`
}

type UpdateProfileRequest struct {
	UID    string        `json:"uid"`
	Grade  *string       `json:"grade,omitempty"`
	Level  *string       `json:"level,omitempty"`
	Status *enum.EStatus `json:"status,omitempty"`
}

type UpdateProfileResponse struct {
	Profile *ProfileResponse `json:"result"`
}

func BuildProfileDomainForCreate(dto *CreateProfileRequest) *domain.Profile {
	profileDomain := domain.NewProfileDomain()
	profileDomain.GenerateID()
	profileDomain.SetUID(dto.UID)
	profileDomain.SetGrade(dto.Grade)
	profileDomain.SetLevel(dto.Level)
	profileDomain.SetStatus(string(enum.StatusActive))

	return profileDomain
}

func BuildProfileDomainForUpdate(dto *UpdateProfileRequest) *domain.Profile {
	profileDomain := domain.NewProfileDomain()

	profileDomain.SetUID(dto.UID)

	if dto.Grade != nil {
		profileDomain.SetGrade(*dto.Grade)
	}

	if dto.Level != nil {
		profileDomain.SetLevel(*dto.Level)
	}

	if dto.Status != nil {
		profileDomain.SetStatus(string(*dto.Status))
	}

	return profileDomain
}

func ProfileResponseFromDomain(p *domain.Profile) ProfileResponse {
	return ProfileResponse{
		ID:         p.ID(),
		UID:        p.UID(),
		Name:       p.Name(),
		Email:      p.Email(),
		Phone:      p.Phone(),
		Grade:      p.Grade(),
		Level:      p.Level(),
		Status:     p.Status(),
		CreatedAt:  p.CreatedAt(),
		ModifiedAt: p.ModifiedAt(),
	}
}
