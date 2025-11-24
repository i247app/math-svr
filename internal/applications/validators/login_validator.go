package validators

import (
	"math-ai.com/math-ai/internal/applications/dto"
	"math-ai.com/math-ai/internal/shared/constant/status"
)

type ILoginValidator interface {
	ValidateLoginRequest(req *dto.LoginRequest) (status.Code, error)
}

type loginValidator struct{}

func NewLoginValidator() *loginValidator {
	return &loginValidator{}
}

func (v *loginValidator) ValidateLoginRequest(req *dto.LoginRequest) (status.Code, error) {
	return status.SUCCESS, nil
}
