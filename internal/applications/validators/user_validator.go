package validators

import (
	"math-ai.com/math-ai/internal/applications/dto"
	"math-ai.com/math-ai/internal/shared/constant/status"
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
	return status.SUCCESS, nil
}

func (v *userValidator) ValidateUpdateUserRequest(req *dto.UpdateUserRequest) (status.Code, error) {
	return status.SUCCESS, nil
}

func (v *userValidator) ValidateDeleteUserRequest(req *dto.DeleteUserRequest) (status.Code, error) {
	return status.SUCCESS, nil
}
