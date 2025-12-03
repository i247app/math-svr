package validators

import (
	"math-ai.com/math-ai/internal/applications/dto"
	"math-ai.com/math-ai/internal/shared/constant/status"
	err_svc "math-ai.com/math-ai/internal/shared/error"
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
	if req.UID == "" {
		return status.USER_MISSING_ID, err_svc.ErrMissingUID
	}

	return status.SUCCESS, nil
}

func (v *ulqValidator) ValidateCreateUserLatestQuizRequest(req *dto.CreateUserLatestQuizRequest) (status.Code, error) {
	if req.UID == "" {
		return status.USER_MISSING_ID, err_svc.ErrMissingUID
	}

	return status.SUCCESS, nil
}

func (v *ulqValidator) ValidateUpdateUserLatestQuizRequest(req *dto.UpdateUserLatestQuizRequest) (status.Code, error) {
	if req.UID == "" {
		return status.USER_MISSING_ID, err_svc.ErrMissingUID
	}

	return status.SUCCESS, nil
}

func (v *ulqValidator) ValidateDeleteUserLatestQuizRequest(req *dto.DeleteUserLatestQuizRequest) (status.Code, error) {
	return status.SUCCESS, nil
}
