package validators

import (
	"math-ai.com/math-ai/internal/applications/dto"
	"math-ai.com/math-ai/internal/shared/constant/status"
	err_svc "math-ai.com/math-ai/internal/shared/error"
)

type IGradeValidator interface {
	ValidateCreateGradeRequest(req *dto.CreateGradeRequest) (status.Code, error)
	ValidateUpdateGradeRequest(req *dto.UpdateGradeRequest) (status.Code, error)
}

type gradeValidator struct{}

func NewGradeValidator() *gradeValidator {
	return &gradeValidator{}
}

func (v *gradeValidator) ValidateCreateGradeRequest(req *dto.CreateGradeRequest) (status.Code, error) {
	if req.Label == "" {
		return status.GRADE_MISSING_LABEL, err_svc.ErrGradeMissingLabel
	}

	if req.Description == "" {
		return status.GRADE_MISSING_DESCRIPTION, err_svc.ErrGradeMissingDescripton
	}

	if req.DisplayOrder == 0 {
		return status.GRADE_MISSING_DISPLAY_ORDER, err_svc.ErrGradeMissingDisplayOrder
	}

	return status.SUCCESS, nil
}

func (v *gradeValidator) ValidateUpdateGradeRequest(req *dto.UpdateGradeRequest) (status.Code, error) {
	if req.ID == "" {
		return status.GRADE_NOT_FOUND, err_svc.ErrGradeNotFound
	}
	return status.SUCCESS, nil
}
