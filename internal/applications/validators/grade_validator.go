package validators

import (
	"math-ai.com/math-ai/internal/applications/dto"
	"math-ai.com/math-ai/internal/shared/constant/status"
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
	return status.SUCCESS, nil
}

func (v *gradeValidator) ValidateUpdateGradeRequest(req *dto.UpdateGradeRequest) (status.Code, error) {
	return status.SUCCESS, nil
}
