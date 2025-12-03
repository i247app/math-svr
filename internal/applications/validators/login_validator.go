package validators

import (
	"math-ai.com/math-ai/internal/applications/dto"
	"math-ai.com/math-ai/internal/shared/constant/status"
	err_svc "math-ai.com/math-ai/internal/shared/error"
)

type ILoginValidator interface {
	ValidateLoginRequest(req *dto.LoginRequest) (status.Code, error)
}

type loginValidator struct{}

func NewLoginValidator() *loginValidator {
	return &loginValidator{}
}

func (v *loginValidator) ValidateLoginRequest(req *dto.LoginRequest) (status.Code, error) {
	if req.LoginName == "" || req.RawPassword == "" {
		return status.LOGIN_MISSING_PARAMETERS, err_svc.ErrMissingParameters
	}

	return status.SUCCESS, nil
}
