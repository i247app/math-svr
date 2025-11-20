package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"math-ai.com/math-ai/internal/core/di/repositories"
	domain "math-ai.com/math-ai/internal/core/domain/login"
	"math-ai.com/math-ai/internal/shared/constant/enum"
	"math-ai.com/math-ai/internal/shared/db"
)

type loginRepository struct {
	db db.IDatabase
}

func NewloginRepository(db db.IDatabase) repositories.ILoginRepository {
	return &loginRepository{
		db: db,
	}
}

// StoreLogin stores a user login record in the database.
func (r *loginRepository) StoreLogin(ctx context.Context, tx *sql.Tx, login *domain.Login) error {
	query := `
		INSERT INTO logins (id, uid, hash_pass, status)
		VALUES (?, ?, ?, ?)
	`
	_, err := r.db.Exec(ctx, tx, query,
		login.ID(),
		login.UID(),
		login.HassPass(),
		enum.StatusActive,
	)
	if err != nil {
		return fmt.Errorf("failed to store user login: %v", err)
	}
	return nil
}

// DeleteLogin deletes user logins by user ID.
func (r *loginRepository) DeleteLogin(ctx context.Context, uid string) error {
	query := `
		UPDATE logins
		SET deleted_dt = ?
		WHERE uid = ? AND deleted_dt IS NULL
	`
	_, err := r.db.Exec(ctx, nil, query, time.Now().UTC(), uid)
	if err != nil {
		return fmt.Errorf("failed to delete user logins: %v", err)
	}
	return nil
}

// ForceDeleteLogin permanently deletes user logins by user ID.
func (r *loginRepository) ForceDeleteLogin(ctx context.Context, tx *sql.Tx, uid string) error {
	query := `
		DELETE FROM logins
		WHERE uid = ?
	`
	_, err := r.db.Exec(ctx, tx, query, uid)
	if err != nil {
		return fmt.Errorf("failed to force delete user logins: %v", err)
	}
	return nil
}
