package di

import (
	"context"
	"database/sql"

	domain "math-ai.com/math-ai/internal/core/domain/user_quiz_practices"
)

type IUserQuizPracticesRepository interface {
	FindByID(ctx context.Context, id string) (*domain.UserQuizPractices, error)
	FindByUID(ctx context.Context, uid string) (*domain.UserQuizPractices, error)
	Create(ctx context.Context, tx *sql.Tx, quiz *domain.UserQuizPractices) (int64, error)
	Update(ctx context.Context, tx *sql.Tx, quiz *domain.UserQuizPractices) (int64, error)
	Delete(ctx context.Context, tx *sql.Tx, id string) (int64, error)
	DeleteByUID(ctx context.Context, tx *sql.Tx, uid string) error
	ForceDelete(ctx context.Context, tx *sql.Tx, id string) (int64, error)
	ForceDeleteByUID(ctx context.Context, tx *sql.Tx, uid string) error
}
