package dto

import (
	"time"

	domain "math-ai.com/math-ai/internal/core/domain/user_quiz_practices"
	"math-ai.com/math-ai/internal/shared/constant/enum"
)

type UserQuizPracticesResponse struct {
	ID         string    `json:"id"`
	UID        string    `json:"uid"`
	Questions  string    `json:"questions"`
	Answers    string    `json:"answers"`
	AIReview   string    `json:"ai_review"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	ModifiedAt time.Time `json:"modified_at"`
}

type GetUserQuizPracticesRequest struct {
	ID string `json:"id"`
}

type GetUserQuizPracticesByUIDRequest struct {
	UID string `json:"uid"`
}

type CreateUserQuizPracticesRequest struct {
	UID       string `json:"uid"`
	Questions string `json:"questions"`
	Answers   string `json:"answers"`
	AIReview  string `json:"ai_review"`
}

type UpdateUserQuizPracticesRequest struct {
	ID        string        `json:"id"`
	UID       string        `json:"uid"`
	Questions *string       `json:"questions,omitempty"`
	Answers   *string       `json:"answers,omitempty"`
	AIReview  *string       `json:"ai_review,omitempty"`
	Status    *enum.EStatus `json:"status,omitempty"`
}

type DeleteUserQuizPracticesRequest struct {
	ID string `json:"id"`
}

func BuildUserQuizPracticesDomainForCreate(req *CreateUserQuizPracticesRequest) *domain.UserQuizPractices {
	quizDomain := domain.NewUserQuizPracticesDomain()
	quizDomain.GenerateID()
	quizDomain.SetUID(req.UID)
	quizDomain.SetQuestions(req.Questions)
	quizDomain.SetAnswers(req.Answers)
	quizDomain.SetAIReview(req.AIReview)
	quizDomain.SetStatus(string(enum.StatusActive))

	return quizDomain
}

func BuildUserQuizPracticesDomainForUpdate(req *UpdateUserQuizPracticesRequest) *domain.UserQuizPractices {
	quizDomain := domain.NewUserQuizPracticesDomain()
	quizDomain.SetID(req.ID)
	quizDomain.SetUID(req.UID)

	if req.Questions != nil {
		quizDomain.SetQuestions(*req.Questions)
	}

	if req.Answers != nil {
		quizDomain.SetAnswers(*req.Answers)
	}

	if req.AIReview != nil {
		quizDomain.SetAIReview(*req.AIReview)
	}

	if req.Status != nil {
		quizDomain.SetStatus(string(*req.Status))
	}

	return quizDomain
}

func UserQuizPracticesResponseFromDomain(q *domain.UserQuizPractices) UserQuizPracticesResponse {
	return UserQuizPracticesResponse{
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
