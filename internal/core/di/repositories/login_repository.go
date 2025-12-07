package di

import (
	"context"
	"database/sql"

	domain "math-ai.com/math-ai/internal/core/domain/login"
)

type ILoginRepository interface {
	// logins
	StoreLogin(ctx context.Context, tx *sql.Tx, login *domain.Login) error
	DeleteLogin(ctx context.Context, tx *sql.Tx, uid string) error
	ForceDeleteLogin(ctx context.Context, tx *sql.Tx, uid string) error

	// login logs
	GetLoginLogByUIDAndDeviceUUID(ctx context.Context, uid string, deviceUUID string) (*domain.LoginLog, error)
	StoreLoginLog(ctx context.Context, loginLog *domain.LoginLog) error
	UpdateLoginLog(ctx context.Context, loginLog *domain.LoginLog) error
}
