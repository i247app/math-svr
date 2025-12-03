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
	if req.Label == "" {
		return status.LEVEL_MISSING_LABEL, nil
	}

	if req.Description == "" {
		return status.LEVEL_MISSING_DESCRIPTION, nil
	}

	if req.DisplayOrder == 0 {
		return status.LEVEL_MISSING_DISPLAY_ORDER, nil
	}

	return status.SUCCESS, nil
}

func (v *levelValidator) ValidateUpdateLevelRequest(req *dto.UpdateLevelRequest) (status.Code, error) {
	if req.ID == "" {
		return status.LEVEL_NOT_FOUND, nil
	}

	return status.SUCCESS, nil
}
