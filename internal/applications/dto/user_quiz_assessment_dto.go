package dto

import (
	"time"

	domain "math-ai.com/math-ai/internal/core/domain/user_quiz_assessment"
	"math-ai.com/math-ai/internal/shared/constant/enum"
	"math-ai.com/math-ai/internal/shared/utils/pagination"
)

// UserQuizAssessmentResponse represents a single quiz assessment
type UserQuizAssessmentResponse struct {
	ID            string    `json:"id"`
	UID           string    `json:"uid"`
	Questions     string    `json:"questions"`
	Answers       string    `json:"answers"`
	AIReview      string    `json:"ai_review"`
	AIDetectGrade string    `json:"ai_detect_grade"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	ModifiedAt    time.Time `json:"modified_at"`
}

// UserQuizAssessmentsHistoryResponse contains paginated list of assessments
type UserQuizAssessmentsHistoryResponse struct {
	Items      []UserQuizAssessmentResponse `json:"items"`
	Pagination *pagination.Pagination       `json:"metadata"`
}

// Request DTOs for generate quiz
type GenerateQuizAssessmentRequest struct {
	UID string `json:"uid"`
	ChatBoxRequestCommon
}

type GenerateQuizAssessmentResponse struct {
	Result *ChatBoxResponse[[]Question] `json:"result"`
}

// Request DTOs for submit quiz
type SubmitQuizAssessmentRequest struct {
	UID     string `json:"uid"`
	Answers []struct {
		QuestionNumber int64  `json:"question_number"`
		Answer         string `json:"answer"`
	} `json:"answers"`
	ChatBoxRequestCommon
}

type QuizAssessmentAnswer struct {
	TotalQuestions  int64  `json:"total_questions"`
	CorrectNumber   int64  `json:"correct_number"`
	ScorePercentage int    `json:"score_percentage"`
	AIReview        string `json:"ai_review"`
	AIDetectGrade   string `json:"ai_detect_grade"`
}

type SubmitQuizAssessmentResponse struct {
	Result *ChatBoxResponse[QuizAssessmentAnswer] `json:"result"`
}

// Request DTOs for reinforce quiz
type ReinforceQuizAssessmentRequest struct {
	UserQuizAssessmentID string `json:"user_quiz_assessment_id"`
	UID                  string `json:"uid"`
	ChatBoxRequestCommon
}

type ReinforceQuizAssessmentResponse struct {
	Result *ChatBoxResponse[[]Question] `json:"result"`
}

// Request DTOs for submit reinforce quiz
type SubmitReinforceQuizAssessmentRequest struct {
	UserQuizAssessmentID string `json:"user_quiz_assessment_id"`
	UID                  string `json:"uid"`
	Answers              []struct {
		QuestionNumber int64  `json:"question_number"`
		Answer         string `json:"answer"`
	} `json:"answers"`
	ChatBoxRequestCommon
}

type SubmitReinforceQuizAssessmentResponse struct {
	Result *ChatBoxResponse[QuizAssessmentAnswer] `json:"result"`
}

// Request DTOs for history
type GetUserQuizAssessmentsHistoryRequest struct {
	UID       string `json:"uid"`
	Page      int64  `json:"page"`
	Limit     int64  `json:"limit"`
	OrderBy   string `json:"order_by" form:"order_by"`
	OrderDesc bool   `json:"order_desc" form:"order_desc"`
	TakeAll   bool   `json:"take_all" form:"take_all"`
}

// Builder functions
func BuildUserQuizAssessmentDomainForCreate(uid, questions, answers, aiReview, aiDetectGrade string) *domain.UserQuizAssessment {
	assessmentDomain := domain.NewUserQuizAssessmentDomain()
	assessmentDomain.GenerateID()
	assessmentDomain.SetUID(uid)
	assessmentDomain.SetQuestions(questions)
	assessmentDomain.SetAnswers(answers)
	assessmentDomain.SetAIReview(aiReview)
	assessmentDomain.SetAIDetectGrade(aiDetectGrade)
	assessmentDomain.SetStatus(string(enum.StatusActive))

	return assessmentDomain
}

func BuildUserQuizAssessmentDomainForUpdate(id, questions, answers, aiReview, aiDetectGrade string) *domain.UserQuizAssessment {
	assessmentDomain := domain.NewUserQuizAssessmentDomain()
	assessmentDomain.SetID(id)

	if questions != "" {
		assessmentDomain.SetQuestions(questions)
	}

	if answers != "" {
		assessmentDomain.SetAnswers(answers)
	}

	if aiReview != "" {
		assessmentDomain.SetAIReview(aiReview)
	}

	if aiDetectGrade != "" {
		assessmentDomain.SetAIDetectGrade(aiDetectGrade)
	}

	return assessmentDomain
}

func UserQuizAssessmentResponseFromDomain(q *domain.UserQuizAssessment) UserQuizAssessmentResponse {
	return UserQuizAssessmentResponse{
		ID:            q.ID(),
		UID:           q.UID(),
		Questions:     q.Questions(),
		Answers:       q.Answers(),
		AIReview:      q.AIReview(),
		AIDetectGrade: q.AIDetectGrade(),
		Status:        q.Status(),
		CreatedAt:     q.CreatedAt(),
		ModifiedAt:    q.ModifiedAt(),
	}
}
