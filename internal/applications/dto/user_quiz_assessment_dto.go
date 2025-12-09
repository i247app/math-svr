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
	UID   string `json:"uid"`
	Grade string `json:"grade,omitempty"`
	ChatBoxRequestCommon
}

type GenerateQuizAssessmentResponse struct {
	Result *ChatBoxResponse[[]Question] `json:"result"`
}

// Request DTOs for submit quiz
type SubmitQuizAssessmentRequest struct {
	UserQuizAssessmentID string `json:"user_quiz_assessment_id"`
	UID                  string `json:"uid"`
	Answers              []struct {
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

type CreateUserQuizAssessmentRequest struct {
	UID           string `json:"uid"`
	Questions     string `json:"questions"`
	Answers       string `json:"answers"`
	AIReview      string `json:"ai_review"`
	AIDetectGrade string `json:"ai_detect_grade"`
}

type UpdateUserQuizAssessmentRequest struct {
	ID            string        `json:"id"`
	UID           string        `json:"uid"`
	Questions     *string       `json:"questions,omitempty"`
	Answers       *string       `json:"answers,omitempty"`
	AIReview      *string       `json:"ai_review,omitempty"`
	AIDetectGrade *string       `json:"ai_detect_grade,omitempty"`
	Status        *enum.EStatus `json:"status,omitempty"`
}

type DeleteUserQuizAssessmentRequest struct {
	ID string `json:"id"`
}

// Builder functions
func BuildUserQuizAssessmentDomainForCreate(req *CreateUserQuizAssessmentRequest) *domain.UserQuizAssessment {
	assessmentDomain := domain.NewUserQuizAssessmentDomain()
	assessmentDomain.GenerateID()
	assessmentDomain.SetUID(req.UID)
	assessmentDomain.SetQuestions(req.Questions)
	assessmentDomain.SetAnswers(req.Answers)
	assessmentDomain.SetAIReview(req.AIReview)
	assessmentDomain.SetAIDetectGrade(req.AIDetectGrade)
	assessmentDomain.SetStatus(string(enum.StatusActive))

	return assessmentDomain
}

func BuildUserQuizAssessmentDomainForUpdate(req *UpdateUserQuizAssessmentRequest) *domain.UserQuizAssessment {
	assessmentDomain := domain.NewUserQuizAssessmentDomain()
	assessmentDomain.SetID(req.ID)

	if req.Questions != nil {
		assessmentDomain.SetQuestions(*req.Questions)
	}

	if req.Answers != nil {
		assessmentDomain.SetAnswers(*req.Answers)
	}

	if req.AIReview != nil {
		assessmentDomain.SetAIReview(*req.AIReview)
	}

	if req.AIDetectGrade != nil {
		assessmentDomain.SetAIDetectGrade(*req.AIDetectGrade)
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
