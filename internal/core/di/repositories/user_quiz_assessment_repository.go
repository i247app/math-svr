package di

import (
	"context"
	"database/sql"

	domain "math-ai.com/math-ai/internal/core/domain/user_quiz_assessment"
	"math-ai.com/math-ai/internal/shared/utils/pagination"
)

type ListUserQuizAssessmentsParams struct {
	UID      string
	Page     int64
	Limit    int64
	TakeAll  bool
	OrderBy  string
	OrderDesc bool
}

type IUserQuizAssessmentRepository interface {
	FindByID(ctx context.Context, id string) (*domain.UserQuizAssessment, error)
	ListByUID(ctx context.Context, params ListUserQuizAssessmentsParams) ([]*domain.UserQuizAssessment, *pagination.Pagination, error)
	Create(ctx context.Context, tx *sql.Tx, quiz *domain.UserQuizAssessment) (int64, error)
	Update(ctx context.Context, quiz *domain.UserQuizAssessment) (int64, error)
	Delete(ctx context.Context, id string) (int64, error)
	ForceDelete(ctx context.Context, id string) (int64, error)
}
