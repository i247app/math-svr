package validators

import (
	"math-ai.com/math-ai/internal/applications/dto"
	"math-ai.com/math-ai/internal/shared/constant/status"
)

type IDeviceValidator interface {
	ValidateCreateDeviceRequest(req *dto.CreateDeviceRequest) (status.Code, error)
	ValidateUpdateDeviceRequest(req *dto.UpdateDeviceRequest) (status.Code, error)
}

type deviceValidator struct{}

func NewDeviceValidator() *deviceValidator {
	return &deviceValidator{}
}

func (v *deviceValidator) ValidateCreateDeviceRequest(req *dto.CreateDeviceRequest) (status.Code, error) {
	return status.SUCCESS, nil
}

func (v *deviceValidator) ValidateUpdateDeviceRequest(req *dto.UpdateDeviceRequest) (status.Code, error) {
	return status.SUCCESS, nil
}
