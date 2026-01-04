package repositories

import (
	"context"
	"database/sql"
	"fmt"

	di "math-ai.com/math-ai/internal/core/di/repositories"
	domain "math-ai.com/math-ai/internal/core/domain/login"
	"math-ai.com/math-ai/internal/driven-adapter/persistence/models"
	"math-ai.com/math-ai/internal/driven-adapter/persistence/queries"
	"math-ai.com/math-ai/internal/shared/constant/enum"
	"math-ai.com/math-ai/internal/shared/db"
	mathtime "math-ai.com/math-ai/internal/shared/utils/time"
)

const (
	loginTableName    = "ma_logins"
	loginLogTableName = "ma_login_logs"
)

type authRepository struct {
	BaseRepository // Embed BaseRepository for common operations
}

func NewAuthRepository(database db.IDatabase) di.IAuthRepository {
	return &authRepository{
		BaseRepository: NewBaseRepository(database),
	}
}

// scanLoginLog is a reusable helper method to scan login log data from a row
func (r *authRepository) scanLoginLog(scanner Scanner) (*domain.LoginLog, error) {
	var ll models.LoginLogModel
	err := scanner.Scan(&ll.ID, &ll.UID, &ll.IPaddress, &ll.DeviceUUID, &ll.Token)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("scan error: %v", err)
	}

	return domain.BuildLoginLogFromModel(&ll), nil
}

// StoreLogin stores a user login record in the database.
// Now uses query constants from queries package
func (r *authRepository) StoreLogin(ctx context.Context, tx *sql.Tx, login *domain.Login) error {
	_, err := r.db.Exec(ctx, tx, queries.LoginInsert,
		login.ID(),
		login.UID(),
		login.HassPass(),
		enum.StatusActive,
		mathtime.Now(),
		mathtime.Now(),
	)
	if err != nil {
		return fmt.Errorf("failed to store user login: %v", err)
	}
	return nil
}

// DeleteLoginByUID deletes user logins by user ID.
// Now uses query constants from queries package
func (r *authRepository) DeleteLoginByUID(ctx context.Context, tx *sql.Tx, uid string) error {
	_, err := r.db.Exec(ctx, tx, queries.LoginSoftDeleteByUID, mathtime.Now(), mathtime.Now(), uid)
	if err != nil {
		return fmt.Errorf("failed to delete user logins: %v", err)
	}
	return nil
}

// ForceDeleteLoginByUID permanently deletes user logins by user ID.
// Now uses query constants from queries package
func (r *authRepository) ForceDeleteLoginByUID(ctx context.Context, tx *sql.Tx, uid string) error {
	_, err := r.db.Exec(ctx, tx, queries.LoginForceDeleteByUID, uid)
	if err != nil {
		return fmt.Errorf("failed to force delete user logins: %v", err)
	}
	return nil
}

// GetLoginLogByUIDAndDeviceUUID retrieves a user login log by user ID and device UUID.
// Now uses query constants from queries package
func (r *authRepository) GetLoginLogByUIDAndDeviceUUID(ctx context.Context, uid string, deviceUUID string) (*domain.LoginLog, error) {
	row := r.db.QueryRow(ctx, nil, queries.LoginLogFindByUIDAndDeviceUUID, uid, deviceUUID)
	return r.scanLoginLog(row)
}

// StoreLoginLog stores a user login log record in the database.
// Now uses query constants from queries package
func (r *authRepository) StoreLoginLog(ctx context.Context, tx *sql.Tx, loginLog *domain.LoginLog) error {
	_, err := r.db.Exec(ctx, tx, queries.LoginLogInsert,
		loginLog.ID(),
		loginLog.UID(),
		loginLog.IPAddress(),
		loginLog.DeviceUUID(),
		loginLog.Token(),
		loginLog.Status(),
		mathtime.Now(),
		mathtime.Now(),
	)

	return err
}

// UpdateLoginLog updates a user login log record in the database.
// Now uses query constants from queries package
func (r *authRepository) UpdateLoginLog(ctx context.Context, tx *sql.Tx, loginLog *domain.LoginLog) error {
	_, err := r.db.Exec(ctx, tx, queries.LoginLogUpdate,
		PrepareForUpdate(loginLog.Token()),
		PrepareForUpdate(loginLog.Status()),
		mathtime.Now(),
		loginLog.UID(),
		loginLog.DeviceUUID(),
	)
	return err
}

// DeleteLoginLogByUID marks login logs as deleted for a given user ID.
// Now uses query constants from queries package
func (r *authRepository) DeleteLoginLogByUID(ctx context.Context, tx *sql.Tx, uid string) error {
	_, err := r.db.Exec(ctx, tx, queries.LoginLogSoftDeleteByUID, mathtime.Now(), mathtime.Now(), uid)
	if err != nil {
		return fmt.Errorf("failed to delete login logs: %v", err)
	}
	return nil
}

// ForceDeleteLoginLogByUID permanently deletes login logs for a given user ID.
// Now uses query constants from queries package
func (r *authRepository) ForceDeleteLoginLogByUID(ctx context.Context, tx *sql.Tx, uid string) error {
	_, err := r.db.Exec(ctx, tx, queries.LoginLogForceDeleteByUID, uid)
	if err != nil {
		return fmt.Errorf("failed to force delete login logs: %v", err)
	}
	return nil
}
