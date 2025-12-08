package validators

import (
	"math-ai.com/math-ai/internal/applications/dto"
	"math-ai.com/math-ai/internal/shared/constant/status"
	err_svc "math-ai.com/math-ai/internal/shared/error"
)

type IUserLatestQuizValidator interface {
	ValidateGetUserLatestQuizRequest(req *dto.GetUserQuizPracticesRequest) (status.Code, error)
	ValidateGetUserLatestQuizByUIDRequest(req *dto.GetUserQuizPracticesByUIDRequest) (status.Code, error)
	ValidateCreateUserLatestQuizRequest(req *dto.CreateUserQuizPracticesRequest) (status.Code, error)
	ValidateUpdateUserLatestQuizRequest(req *dto.UpdateUserQuizPracticesRequest) (status.Code, error)
	ValidateDeleteUserLatestQuizRequest(req *dto.DeleteUserQuizPracticesRequest) (status.Code, error)
}

type ulqValidator struct{}

func NewUserQuizPracticesValidator() *ulqValidator {
	return &ulqValidator{}
}

func (v *ulqValidator) ValidateGetUserLatestQuizRequest(req *dto.GetUserQuizPracticesRequest) (status.Code, error) {
	return status.SUCCESS, nil
}

func (v *ulqValidator) ValidateGetUserLatestQuizByUIDRequest(req *dto.GetUserQuizPracticesByUIDRequest) (status.Code, error) {
	if req.UID == "" {
		return status.USER_MISSING_ID, err_svc.ErrMissingUID
	}

	return status.SUCCESS, nil
}

func (v *ulqValidator) ValidateCreateUserLatestQuizRequest(req *dto.CreateUserQuizPracticesRequest) (status.Code, error) {
	if req.UID == "" {
		return status.USER_MISSING_ID, err_svc.ErrMissingUID
	}

	return status.SUCCESS, nil
}

func (v *ulqValidator) ValidateUpdateUserLatestQuizRequest(req *dto.UpdateUserQuizPracticesRequest) (status.Code, error) {
	if req.UID == "" {
		return status.USER_MISSING_ID, err_svc.ErrMissingUID
	}

	return status.SUCCESS, nil
}

func (v *ulqValidator) ValidateDeleteUserLatestQuizRequest(req *dto.DeleteUserQuizPracticesRequest) (status.Code, error) {
	return status.SUCCESS, nil
}
