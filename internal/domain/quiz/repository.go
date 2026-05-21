package quiz

import (
	"context"

	"github.com/google/uuid"
	"math-ai.com/math-ai/internal/shared/pagination"
)

// GradingUpdate bundles the bot's grading output for a single quiz so
// the repository write site has one named argument instead of six. The
// counts are pointers because future EXAM types may opt out of full
// scoring; today PRACTICE and ASSESSMENT both populate all three.
type GradingUpdate struct {
	AIReview        string
	AIDetectGrade   *string
	TotalQuestions  *int
	CorrectNumber   *int
	ScorePercentage *int
}

// IRepository owns all quiz persistence. UpdateAnswersAndGrading is
// split from a generic Update so the grade-after-submit path can write
// the answers + AI grading fields in one shot without forcing COALESCE
// on the JSON columns.
type IRepository interface {
	FindByQuizId(ctx context.Context, quizId uuid.UUID) (*Quiz, error)
	ListByProfileId(ctx context.Context, profileId uuid.UUID, page, limit int64) ([]*Quiz, *pagination.Pagination, error)
	Create(ctx context.Context, q *Quiz) (*Quiz, error)
	UpdateAnswersAndGrading(ctx context.Context, quizId uuid.UUID, answers string,
		grading GradingUpdate, quizStatus string) error
	SoftDelete(ctx context.Context, quizId uuid.UUID) error
	SoftDeleteByUserId(ctx context.Context, userId uuid.UUID) error
	ForceDeleteByUserId(ctx context.Context, userId uuid.UUID) error
}
