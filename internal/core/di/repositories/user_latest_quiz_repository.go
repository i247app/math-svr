package di

import (
	"context"
	"database/sql"

	domain "math-ai.com/math-ai/internal/core/domain/user_latest_quiz"
)

type IUserLatestQuizRepository interface {
	FindByID(ctx context.Context, id string) (*domain.UserLatestQuiz, error)
	FindByUID(ctx context.Context, uid string) (*domain.UserLatestQuiz, error)
	List(ctx context.Context, limit, offset int) ([]*domain.UserLatestQuiz, error)
	Create(ctx context.Context, tx *sql.Tx, quiz *domain.UserLatestQuiz) (int64, error)
	Update(ctx context.Context, quiz *domain.UserLatestQuiz) (int64, error)
	Delete(ctx context.Context, id string) (int64, error)
	ForceDelete(ctx context.Context, id string) (int64, error)
}
