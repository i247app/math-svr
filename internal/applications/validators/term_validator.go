package validators

import (
	"math-ai.com/math-ai/internal/applications/dto"
	"math-ai.com/math-ai/internal/shared/constant/status"
	err_svc "math-ai.com/math-ai/internal/shared/error"
)

type ITermValidator interface {
	ValidateCreateTermRequest(req *dto.CreateTermRequest) (status.Code, error)
	ValidateUpdateTermRequest(req *dto.UpdateTermRequest) (status.Code, error)
}

type termValidator struct{}

func NewTermValidator() *termValidator {
	return &termValidator{}
}

func (v *termValidator) ValidateCreateTermRequest(req *dto.CreateTermRequest) (status.Code, error) {
	if req.Name == "" {
		return status.TERM_MISSING_NAME, err_svc.ErrTermMissingName
	}

	if req.Description == nil {
		return status.TERM_MISSING_DESCRIPTION, err_svc.ErrTermMissingDescription
	}

	return status.SUCCESS, nil
}

func (v *termValidator) ValidateUpdateTermRequest(req *dto.UpdateTermRequest) (status.Code, error) {
	if req.ID == "" {
		return status.TERM_NOT_FOUND, err_svc.ErrTermNotFound
	}
	return status.SUCCESS, nil
}
