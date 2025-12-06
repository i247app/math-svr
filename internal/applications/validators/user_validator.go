package validators

import (
	"math-ai.com/math-ai/internal/applications/dto"
	"math-ai.com/math-ai/internal/shared/constant/status"
	err_svc "math-ai.com/math-ai/internal/shared/error"
	"math-ai.com/math-ai/internal/shared/utils/validate"
)

type IUserValidator interface {
	ValidateCreateUserRequest(req *dto.CreateUserRequest) (status.Code, error)
	ValidateUpdateUserRequest(req *dto.UpdateUserRequest) (status.Code, error)
	ValidateDeleteUserRequest(req *dto.DeleteUserRequest) (status.Code, error)
}

type userValidator struct{}

func NewUserValidator() *userValidator {
	return &userValidator{}
}

func (v *userValidator) ValidateCreateUserRequest(req *dto.CreateUserRequest) (status.Code, error) {
	if req.Name == "" {
		return status.USER_MISSING_NAME, err_svc.ErrMissingName
	}

	if req.Phone == "" {
		return status.USER_MISSING_PHONE, err_svc.ErrMissingPhone
	} else if !validate.IsValidPhoneNumber(req.Phone) {
		return status.USER_INVALID_PHONE, err_svc.ErrInvalidPhone
	}

	if req.Email == "" {
		return status.USER_MISSING_EMAIL, err_svc.ErrMissingEmail
	} else if !validate.IsValidEmail(req.Email) {
		return status.USER_INVALID_EMAIL, err_svc.ErrInvalidEmail
	}

	if req.Password == "" {
		return status.USER_MISSING_PASSWORD, err_svc.ErrMissingPassword
	}

	if req.Role != "" && !validate.IsValidRole(req.Role) {
		return status.USER_INVALID_ROLE, err_svc.ErrInvalidRole
	}

	return status.SUCCESS, nil
}

func (v *userValidator) ValidateUpdateUserRequest(req *dto.UpdateUserRequest) (status.Code, error) {
	if req.UID == "" {
		return status.USER_MISSING_ID, err_svc.ErrMissingUID
	}

	return status.SUCCESS, nil
}

func (v *userValidator) ValidateDeleteUserRequest(req *dto.DeleteUserRequest) (status.Code, error) {
	if req.UID == "" {
		return status.USER_MISSING_ID, err_svc.ErrMissingUID
	}

	return status.SUCCESS, nil
}
