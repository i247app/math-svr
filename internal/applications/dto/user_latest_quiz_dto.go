package dto

import (
	"time"

	domain "math-ai.com/math-ai/internal/core/domain/user_latest_quiz"
	"math-ai.com/math-ai/internal/shared/constant/enum"
)

type UserLatestQuizResponse struct {
	ID         string    `json:"id"`
	UID        string    `json:"uid"`
	Questions  string    `json:"questions"`
	Answers    string    `json:"answers"`
	AIReview   string    `json:"ai_review"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	ModifiedAt time.Time `json:"modified_at"`
}

type ListUserLatestQuizzesRequest struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

type GetUserLatestQuizRequest struct {
	ID string `json:"id"`
}

type GetUserLatestQuizByUIDRequest struct {
	UID string `json:"uid"`
}

type CreateUserLatestQuizRequest struct {
	UID       string `json:"uid"`
	Questions string `json:"questions"`
	Answers   string `json:"answers"`
	AIReview  string `json:"ai_review"`
}

type UpdateUserLatestQuizRequest struct {
	ID        string        `json:"id"`
	Questions *string       `json:"questions,omitempty"`
	Answers   *string       `json:"answers,omitempty"`
	AIReview  *string       `json:"ai_review,omitempty"`
	Status    *enum.EStatus `json:"status,omitempty"`
}

type DeleteUserLatestQuizRequest struct {
	ID string `json:"id"`
}

type ForceDeleteUserLatestQuizRequest struct {
	ID string `json:"id"`
}

func BuildUserLatestQuizDomainForCreate(dto *CreateUserLatestQuizRequest) *domain.UserLatestQuiz {
	quizDomain := domain.NewUserLatestQuizDomain()
	quizDomain.GenerateID()
	quizDomain.SetUID(dto.UID)
	quizDomain.SetQuestions(dto.Questions)
	quizDomain.SetAnswers(dto.Answers)
	quizDomain.SetAIReview(dto.AIReview)
	quizDomain.SetStatus(string(enum.StatusActive))

	return quizDomain
}

func BuildUserLatestQuizDomainForUpdate(dto *UpdateUserLatestQuizRequest) *domain.UserLatestQuiz {
	quizDomain := domain.NewUserLatestQuizDomain()
	quizDomain.SetID(dto.ID)

	if dto.Questions != nil {
		quizDomain.SetQuestions(*dto.Questions)
	}

	if dto.Answers != nil {
		quizDomain.SetAnswers(*dto.Answers)
	}

	if dto.AIReview != nil {
		quizDomain.SetAIReview(*dto.AIReview)
	}

	if dto.Status != nil {
		quizDomain.SetStatus(string(*dto.Status))
	}

	return quizDomain
}

func UserLatestQuizResponseFromDomain(q *domain.UserLatestQuiz) UserLatestQuizResponse {
	return UserLatestQuizResponse{
		ID:         q.ID(),
		UID:        q.UID(),
		Questions:  q.Questions(),
		Answers:    q.Answers(),
		AIReview:   q.AIReview(),
		Status:     q.Status(),
		CreatedAt:  q.CreatedAt(),
		ModifiedAt: q.ModifiedAt(),
	}
}
