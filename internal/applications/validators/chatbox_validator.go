package validators

import (
	"math-ai.com/math-ai/internal/applications/dto"
	"math-ai.com/math-ai/internal/shared/constant/status"
)

type IChatboxValidator interface {
	ValidateGenerateQuizRequest(req *dto.GenerateQuizRequest) (status.Code, error)
	ValidateSubmitAnswerRequest(req *dto.SubmitQuizRequest) (status.Code, error)
	ValidateGenerateQuizPracticeRequest(req *dto.GenerateQuizPracticeRequest) (status.Code, error)
}

type chatboxValidator struct{}

func NewChatboxValidator() *chatboxValidator {
	return &chatboxValidator{}
}

func (v *chatboxValidator) ValidateGenerateQuizRequest(req *dto.GenerateQuizRequest) (status.Code, error) {
	if req.UID == "" {
		return status.USER_MISSING_ID, nil
	}

	return status.SUCCESS, nil
}

func (v *chatboxValidator) ValidateSubmitAnswerRequest(req *dto.SubmitQuizRequest) (status.Code, error) {
	if req.UID == "" {
		return status.USER_MISSING_ID, nil
	}

	if req.Answers == nil {
		return status.USER_MISSING_QUIZ_ANSWERS, nil
	}

	return status.SUCCESS, nil
}

func (v *chatboxValidator) ValidateGenerateQuizPracticeRequest(req *dto.GenerateQuizPracticeRequest) (status.Code, error) {
	if req.UID == "" {
		return status.USER_MISSING_ID, nil
	}

	return status.SUCCESS, nil
}
