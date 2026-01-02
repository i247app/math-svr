package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	di "math-ai.com/math-ai/internal/core/di/repositories"
	domain "math-ai.com/math-ai/internal/core/domain/login"
	"math-ai.com/math-ai/internal/driven-adapter/persistence/models"
	"math-ai.com/math-ai/internal/shared/constant/enum"
	"math-ai.com/math-ai/internal/shared/db"
	mathtime "math-ai.com/math-ai/internal/shared/utils/time"
)

type authRepository struct {
	db db.IDatabase
}

func NewAuthRepository(db db.IDatabase) di.IAuthRepository {
	return &authRepository{
		db: db,
	}
}

// StoreLogin stores a user login record in the database.
func (r *authRepository) StoreLogin(ctx context.Context, tx *sql.Tx, login *domain.Login) error {
	query := `
		INSERT INTO ma_logins (id, uid, hash_pass, status, create_dt, modify_dt)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.Exec(ctx, tx, query,
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

// DeleteLogin deletes user logins by user ID.
func (r *authRepository) DeleteLoginByUID(ctx context.Context, tx *sql.Tx, uid string) error {
	query := `
		UPDATE ma_logins
		SET deleted_dt = ?,
			modify_dt = ?
		WHERE uid = ? AND deleted_dt IS NULL
	`
	_, err := r.db.Exec(ctx, tx, query, mathtime.Now(), mathtime.Now(), uid)
	if err != nil {
		return fmt.Errorf("failed to delete user logins: %v", err)
	}
	return nil
}

// ForceDeleteLogin permanently deletes user logins by user ID.
func (r *authRepository) ForceDeleteLoginByUID(ctx context.Context, tx *sql.Tx, uid string) error {
	query := `
		DELETE FROM ma_logins
		WHERE uid = ?
	`
	_, err := r.db.Exec(ctx, tx, query, uid)
	if err != nil {
		return fmt.Errorf("failed to force delete user logins: %v", err)
	}
	return nil
}

// GetLoginByUID retrieves a user login by user ID.
func (r *authRepository) GetLoginLogByUIDAndDeviceUUID(ctx context.Context, uid string, deviceUUID string) (*domain.LoginLog, error) {
	query := `
		SELECT id, uid, ip_address, device_uuid, token
		FROM ma_login_logs
		WHERE uid = ? AND device_uuid = ?
	`

	var ll models.LoginLogModel
	result := r.db.QueryRow(ctx, nil, query, uid, deviceUUID)
	err := result.Scan(&ll.ID, &ll.UID, &ll.IPaddress, &ll.DeviceUUID, &ll.Token)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	loginLog := domain.BuildLoginLogFromModel(&ll)

	return loginLog, nil
}

// StoreLoginLog stores a user login log record in the database.
func (r *authRepository) StoreLoginLog(ctx context.Context, loginLog *domain.LoginLog) error {
	query := `
		INSERT INTO ma_login_logs (id, uid, ip_address, device_uuid, token, status, create_dt, modify_dt)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.Exec(ctx, nil, query,
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
func (r *authRepository) UpdateLoginLog(ctx context.Context, loginLog *domain.LoginLog) error {
	var queryBuilder strings.Builder
	args := []interface{}{}

	queryBuilder.WriteString("UPDATE ma_login_logs SET ")
	updates := []string{}

	if loginLog.Token() != "" {
		updates = append(updates, "token = ?")
		args = append(args, loginLog.Token())
	}

	if loginLog.Status() != "" {
		updates = append(updates, "status = ?")
		args = append(args, loginLog.Status())
	}

	queryBuilder.WriteString(strings.Join(updates, ", "))
	queryBuilder.WriteString(" WHERE uid = ? AND device_uuid = ?")
	args = append(args, loginLog.UID(), loginLog.DeviceUUID())

	query := queryBuilder.String()

	_, err := r.db.Exec(ctx, nil, query, args...)
	return err
}

// DeleteLoginLogByUID marks login logs as deleted for a given user ID.
func (r *authRepository) DeleteLoginLogByUID(ctx context.Context, uid string) error {
	query := `
		UPDATE ma_login_logs
		SET deleted_dt = ?,
			modify_dt = ?
		WHERE uid = ? AND deleted_dt IS NULL
	`
	_, err := r.db.Exec(ctx, nil, query, mathtime.Now(), mathtime.Now(), uid)
	if err != nil {
		return fmt.Errorf("failed to delete login logs: %v", err)
	}
	return nil
}

// ForceDeleteLoginLogByUID permanently deletes login logs for a given user ID.
func (r *authRepository) ForceDeleteLoginLogByUID(ctx context.Context, uid string) error {
	query := `
		DELETE FROM ma_login_logs
		WHERE uid = ?
	`
	_, err := r.db.Exec(ctx, nil, query, uid)
	if err != nil {
		return fmt.Errorf("failed to force delete login logs: %v", err)
	}
	return nil
}
