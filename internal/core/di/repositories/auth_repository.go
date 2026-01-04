package di

import (
	"context"
	"database/sql"

	domain "math-ai.com/math-ai/internal/core/domain/login"
)

type IAuthRepository interface {
	// logins
	StoreLogin(ctx context.Context, tx *sql.Tx, login *domain.Login) error
	DeleteLoginByUID(ctx context.Context, tx *sql.Tx, uid string) error
	ForceDeleteLoginByUID(ctx context.Context, tx *sql.Tx, uid string) error

	// login logs
	GetLoginLogByUIDAndDeviceUUID(ctx context.Context, uid string, deviceUUID string) (*domain.LoginLog, error)
	StoreLoginLog(ctx context.Context, tx *sql.Tx, loginLog *domain.LoginLog) error
	UpdateLoginLog(ctx context.Context, tx *sql.Tx, loginLog *domain.LoginLog) error
	DeleteLoginLogByUID(ctx context.Context, tx *sql.Tx, uid string) error
	ForceDeleteLoginLogByUID(ctx context.Context, tx *sql.Tx, uid string) error
}
