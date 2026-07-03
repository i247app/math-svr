package quiz

import (
	"context"

	"math-ai.com/math-ai/internal/shared/pagination"
)

// GradingUpdate bundles the bot's grading output for a single quiz so
// the repository write site has one named argument instead of six. The
// counts are pointers because future EXAM types may opt out of full
// scoring; today PRACTICE and ASSESSMENT both populate all three.
type GradingUpdate struct {
	AIReview        string
	AssessmentGrade *string
	TotalQuestions  *int
	CorrectNumber   *int
	ScorePercentage *int
}

// ListQuizzesFilter narrows the listing query. Either or both of
// ProfileID / UserID may be set; when both are set the repo AND's them
// so the result is the intersection. At least one must be non-nil —
// the caller is responsible for that check.
type ListQuizzesFilter struct {
	ProfileID *int64
	UserID    *int64
	Purpose   *string
}

// IRepository owns all quiz persistence. UpdateAnswersAndGrading is
// split from a generic Update so the grade-after-submit path can write
// the answers + AI grading fields in one shot without forcing COALESCE
// on the JSON columns.
type IRepository interface {
	FindByQuizId(ctx context.Context, quizId int64) (*Quiz, error)
	ListQuizzes(ctx context.Context, filter ListQuizzesFilter, page, limit int64) ([]*Quiz, *pagination.Pagination, error)
	Create(ctx context.Context, q *Quiz) (*Quiz, error)
	UpdateAnswersAndGrading(ctx context.Context, quizId int64, answers string, grading GradingUpdate, quizStatus string) error
	SoftDelete(ctx context.Context, quizId int64) error
	SoftDeleteByUserId(ctx context.Context, userId int64) error
	ForceDeleteByUserId(ctx context.Context, userId int64) error
}
