package validators

import (
	"math-ai.com/math-ai/internal/applications/dto"
	"math-ai.com/math-ai/internal/shared/constant/status"
)

type IUserLatestQuizValidator interface {
	ValidateGetUserLatestQuizRequest(req *dto.GetUserLatestQuizRequest) (status.Code, error)
	ValidateGetUserLatestQuizByUIDRequest(req *dto.GetUserLatestQuizByUIDRequest) (status.Code, error)
	ValidateCreateUserLatestQuizRequest(req *dto.CreateUserLatestQuizRequest) (status.Code, error)
	ValidateUpdateUserLatestQuizRequest(req *dto.UpdateUserLatestQuizRequest) (status.Code, error)
	ValidateDeleteUserLatestQuizRequest(req *dto.DeleteUserLatestQuizRequest) (status.Code, error)
}

type ulqValidator struct{}

func NewUserLatestQuizValidator() *ulqValidator {
	return &ulqValidator{}
}

func (v *ulqValidator) ValidateGetUserLatestQuizRequest(req *dto.GetUserLatestQuizRequest) (status.Code, error) {
	return status.SUCCESS, nil
}

func (v *ulqValidator) ValidateGetUserLatestQuizByUIDRequest(req *dto.GetUserLatestQuizByUIDRequest) (status.Code, error) {
	return status.SUCCESS, nil
}

func (v *ulqValidator) ValidateCreateUserLatestQuizRequest(req *dto.CreateUserLatestQuizRequest) (status.Code, error) {
	return status.SUCCESS, nil
}

func (v *ulqValidator) ValidateUpdateUserLatestQuizRequest(req *dto.UpdateUserLatestQuizRequest) (status.Code, error) {
	return status.SUCCESS, nil
}

func (v *ulqValidator) ValidateDeleteUserLatestQuizRequest(req *dto.DeleteUserLatestQuizRequest) (status.Code, error) {
	return status.SUCCESS, nil
}
