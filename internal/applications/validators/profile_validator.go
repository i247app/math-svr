package validators

import (
	"math-ai.com/math-ai/internal/applications/dto"
	"math-ai.com/math-ai/internal/shared/constant/status"
)

type IProfileValidator interface {
	ValidateGetProfileRequest(req *dto.FetchProfileRequest) (status.Code, error)
	ValidateCreateProfileRequest(req *dto.CreateProfileRequest) (status.Code, error)
	ValidateUpdateProfileRequest(req *dto.UpdateProfileRequest) (status.Code, error)
}

type profileValidator struct{}

func NewProfileValidator() *profileValidator {
	return &profileValidator{}
}

func (v *profileValidator) ValidateGetProfileRequest(req *dto.FetchProfileRequest) (status.Code, error) {
	return status.SUCCESS, nil
}

func (v *profileValidator) ValidateCreateProfileRequest(req *dto.CreateProfileRequest) (status.Code, error) {
	return status.SUCCESS, nil
}

func (v *profileValidator) ValidateUpdateProfileRequest(req *dto.UpdateProfileRequest) (status.Code, error) {
	return status.SUCCESS, nil
}
