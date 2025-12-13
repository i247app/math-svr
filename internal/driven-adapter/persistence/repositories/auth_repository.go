package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	di "math-ai.com/math-ai/internal/core/di/repositories"
	domain "math-ai.com/math-ai/internal/core/domain/login"
	"math-ai.com/math-ai/internal/driven-adapter/persistence/models"
	"math-ai.com/math-ai/internal/shared/constant/enum"
	"math-ai.com/math-ai/internal/shared/db"
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
func (r *authRepository) DeleteLogin(ctx context.Context, tx *sql.Tx, uid string) error {
	query := `
		UPDATE logins
		SET deleted_dt = ?,
			modify_dt = ?
		WHERE uid = ? AND deleted_dt IS NULL
	`
	_, err := r.db.Exec(ctx, tx, query, time.Now().UTC(), time.Now().UTC(), uid)
	if err != nil {
		return fmt.Errorf("failed to delete user logins: %v", err)
	}
	return nil
}

// ForceDeleteLogin permanently deletes user logins by user ID.
func (r *authRepository) ForceDeleteLogin(ctx context.Context, tx *sql.Tx, uid string) error {
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

func (r *authRepository) GetLoginLogByUIDAndDeviceUUID(ctx context.Context, uid string, deviceUUID string) (*domain.LoginLog, error) {
	query := `
		SELECT id, uid, ip_address, device_uuid, token
		FROM login_logs
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

func (r *authRepository) StoreLoginLog(ctx context.Context, loginLog *domain.LoginLog) error {
	query := `
		INSERT INTO login_logs (id, uid, ip_address, device_uuid, token, status)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.Exec(ctx, nil, query,
		loginLog.ID(),
		loginLog.UID(),
		loginLog.IPAddress(),
		loginLog.DeviceUUID(),
		loginLog.Token(),
		loginLog.Status(),
	)

	return err
}

func (r *authRepository) UpdateLoginLog(ctx context.Context, loginLog *domain.LoginLog) error {
	var queryBuilder strings.Builder
	args := []interface{}{}

	queryBuilder.WriteString("UPDATE login_logs SET ")
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
