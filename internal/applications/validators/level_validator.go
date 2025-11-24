package validators

import (
	"math-ai.com/math-ai/internal/applications/dto"
	"math-ai.com/math-ai/internal/shared/constant/status"
)

type ILevelValidator interface {
	ValidateCreateLevelRequest(req *dto.CreateLevelRequest) (status.Code, error)
	ValidateUpdateLevelRequest(req *dto.UpdateLevelRequest) (status.Code, error)
}

type levelValidator struct{}

func NewLevelValidator() *levelValidator {
	return &levelValidator{}
}

func (v *levelValidator) ValidateCreateLevelRequest(req *dto.CreateLevelRequest) (status.Code, error) {
	return status.SUCCESS, nil
}

func (v *levelValidator) ValidateUpdateLevelRequest(req *dto.UpdateLevelRequest) (status.Code, error) {
	return status.SUCCESS, nil
}
