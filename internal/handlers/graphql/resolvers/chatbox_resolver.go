package resolvers

import (
	"github.com/graphql-go/graphql"
	"math-ai.com/math-ai/internal/applications/dto"
	di "math-ai.com/math-ai/internal/core/di/services"
	"math-ai.com/math-ai/internal/shared/constant/status"
)

type ChatBoxResolver struct {
	userQuizPracticesSvc di.IUserQuizPracticesService
}

func NewChatBoxResolver(userQuizPracticesSvc di.IUserQuizPracticesService) *ChatBoxResolver {
	return &ChatBoxResolver{
		userQuizPracticesSvc: userQuizPracticesSvc,
	}
}

// GenerateQuiz resolves quiz generation
func (r *ChatBoxResolver) GenerateQuiz(params graphql.ResolveParams) (interface{}, error) {
	input, ok := params.Args["input"].(map[string]interface{})
	if !ok {
		return nil, nil
	}

	req := &dto.GenerateQuizRequest{
		UID: input["uid"].(string),
	}

	if typeOfTask, ok := input["type_of_task"].(string); ok {
		req.TypeOfQuiz = typeOfTask
	}
	if typeOfPurpose, ok := input["type_of_purpose"].(string); ok {
		// req.TypeOfPurpose = enum.ETypeQuizPurpuse(typeOfPurpose)
		_ = typeOfPurpose // Handle type conversion as needed
	}
	if message, ok := input["message"].(string); ok {
		req.Message = message
	}

	statusCode, response, err := r.userQuizPracticesSvc.GenerateQuiz(params.Context, req)
	if err != nil || statusCode != status.SUCCESS {
		return nil, err
	}

	return response, nil
}

// SubmitQuiz resolves quiz answer submission
func (r *ChatBoxResolver) SubmitQuiz(params graphql.ResolveParams) (interface{}, error) {
	input, ok := params.Args["input"].(map[string]interface{})
	if !ok {
		return nil, nil
	}

	req := &dto.SubmitQuizRequest{
		UID: input["uid"].(string),
	}

	// Parse answers
	if answersRaw, ok := input["answers"].([]interface{}); ok {
		answers := make([]struct {
			QuestionNumber int64  `json:"question_number"`
			Answer         string `json:"answer"`
		}, len(answersRaw))

		for i, answerRaw := range answersRaw {
			answerMap := answerRaw.(map[string]interface{})
			answers[i].QuestionNumber = int64(answerMap["question_number"].(int))
			answers[i].Answer = answerMap["answer"].(string)
		}
		req.Answers = answers
	}

	statusCode, response, err := r.userQuizPracticesSvc.SubmitQuiz(params.Context, req)
	if err != nil || statusCode != status.SUCCESS {
		return nil, err
	}

	return response, nil
}

// GenerateQuizPractice resolves practice quiz generation
func (r *ChatBoxResolver) GenerateQuizPractice(params graphql.ResolveParams) (interface{}, error) {
	input, ok := params.Args["input"].(map[string]interface{})
	if !ok {
		return nil, nil
	}

	req := &dto.GenerateQuizPracticeRequest{
		UID: input["uid"].(string),
	}

	if message, ok := input["message"].(string); ok {
		req.Message = message
	}

	statusCode, response, err := r.userQuizPracticesSvc.GenerateQuizPractice(params.Context, req)
	if err != nil || statusCode != status.SUCCESS {
		return nil, err
	}

	return response, nil
}
