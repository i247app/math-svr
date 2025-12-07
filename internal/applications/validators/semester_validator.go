package validators

import (
	"math-ai.com/math-ai/internal/applications/dto"
	"math-ai.com/math-ai/internal/shared/constant/status"
	err_svc "math-ai.com/math-ai/internal/shared/error"
)

type ISemesterValidator interface {
	ValidateCreateSemesterRequest(req *dto.CreateSemesterRequest) (status.Code, error)
	ValidateUpdateSemesterRequest(req *dto.UpdateSemesterRequest) (status.Code, error)
}

type semesterValidator struct{}

func NewSemesterValidator() *semesterValidator {
	return &semesterValidator{}
}

func (v *semesterValidator) ValidateCreateSemesterRequest(req *dto.CreateSemesterRequest) (status.Code, error) {
	if req.Name == "" {
		return status.SEMESTER_MISSING_NAME, err_svc.ErrSemesterMissingName
	}

	if req.Description == nil {
		return status.SEMESTER_MISSING_DESCRIPTION, err_svc.ErrSemesterMissingDescription
	}

	return status.SUCCESS, nil
}

func (v *semesterValidator) ValidateUpdateSemesterRequest(req *dto.UpdateSemesterRequest) (status.Code, error) {
	if req.ID == "" {
		return status.SEMESTER_NOT_FOUND, err_svc.ErrSemesterNotFound
	}
	return status.SUCCESS, nil
}
