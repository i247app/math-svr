package repositories

import (
	"context"
	"database/sql"

	domain "math-ai.com/math-ai/internal/core/domain/login"
)

type ILoginRepository interface {
	// logins
	StoreLogin(ctx context.Context, tx *sql.Tx, login *domain.Login) error
	DeleteLogin(ctx context.Context, uid string) error
	ForceDeleteLogin(ctx context.Context, tx *sql.Tx, uid string) error
}
