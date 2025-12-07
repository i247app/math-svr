package validators

import (
	"math-ai.com/math-ai/internal/applications/dto"
	"math-ai.com/math-ai/internal/shared/constant/status"
	err_svc "math-ai.com/math-ai/internal/shared/error"
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
	if req.UID == "" {
		return status.USER_MISSING_ID, err_svc.ErrMissingUID
	}

	return status.SUCCESS, nil
}

func (v *profileValidator) ValidateCreateProfileRequest(req *dto.CreateProfileRequest) (status.Code, error) {
	if req.UID == "" {
		return status.USER_MISSING_ID, err_svc.ErrMissingUID
	}

	if req.GradeID == "" {
		return status.PROFILE_MISSING_GRADE, err_svc.ErrProfileMissingGrade
	}

	if req.SemesterID == "" {
		return status.PROFILE_MISSING_SEMESTER, err_svc.ErrProfileMissingSemester
	}

	return status.SUCCESS, nil
}

func (v *profileValidator) ValidateUpdateProfileRequest(req *dto.UpdateProfileRequest) (status.Code, error) {
	if req.UID == "" {
		return status.FAIL, err_svc.ErrMissingUID
	}

	return status.SUCCESS, nil
}
