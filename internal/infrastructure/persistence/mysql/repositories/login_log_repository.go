package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"math-ai.com/math-ai/internal/domain/loginlog"
	mtime "math-ai.com/math-ai/internal/domain/shared/time"
	"math-ai.com/math-ai/internal/infrastructure/database"
	"math-ai.com/math-ai/internal/infrastructure/persistence/mysql/models"
	"math-ai.com/math-ai/internal/shared/enum"
	"math-ai.com/math-ai/internal/shared/utils"
)

const (
	loginLogTable = "ma_login_logs"

	loginLogColumns = `l.id, l.login_log_id, l.user_id, l.ip_address, l.device_uuid,
		l.token, l.note, l.login_log_status, l.status,
		l.create_id, l.create_dt, l.modify_id, l.modify_dt`

	loginLogActiveWhere = `l.status IN (?) AND l.deleted_dt IS NULL`
)

func loginLogActiveArgs() []any {
	return []any{enum.StatusActive}
}

type LoginLogRepository struct {
	db database.Executor
}

func NewLoginLogRepository(db database.Executor) loginlog.IRepository {
	return &LoginLogRepository{db: db}
}

func scanLoginLog(s database.RowScanner) (*models.LoginLogModel, error) {
	var m models.LoginLogModel
	if err := s.Scan(&m.Id, &m.LoginLogId, &m.UserId, &m.IpAddress, &m.DeviceUUID,
		&m.Token, &m.Note, &m.LoginLogStatus, &m.Status,
		&m.CreateId, &m.CreateDt, &m.ModifyId, &m.ModifyDt); err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *LoginLogRepository) findOneBy(ctx context.Context, where string, args ...any) (*loginlog.LoginLog, error) {
	fullArgs := append(loginLogActiveArgs(), args...)
	query := `SELECT ` + loginLogColumns + ` FROM ` + loginLogTable + ` l WHERE ` +
		loginLogActiveWhere + ` AND (` + where + `)`

	m, err := scanLoginLog(r.db.QueryRow(ctx, query, fullArgs...))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("login_log repo find (%s): %w", where, err)
	}
	return ModelToDomainLoginLog(m), nil
}

func (r *LoginLogRepository) findBareById(ctx context.Context, id int64) (*loginlog.LoginLog, error) {
	args := append(loginLogActiveArgs(), id)
	query := `SELECT ` + loginLogColumns + ` FROM ` + loginLogTable + ` l WHERE ` +
		loginLogActiveWhere + ` AND l.id = ?`

	m, err := scanLoginLog(r.db.QueryRow(ctx, query, args...))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("login_log repo find bare by id: %w", err)
	}
	return ModelToDomainLoginLog(m), nil
}

func (r *LoginLogRepository) FindByLoginLogId(ctx context.Context, loginLogId string) (*loginlog.LoginLog, error) {
	return r.findOneBy(ctx, "l.login_log_id = ?", loginLogId)
}

// FindActiveByToken returns the login_log only when its business status is
// still ACTIVE. A logout flips login_log_status away from ACTIVE so revoked
// tokens stop resolving here even though the row is preserved.
func (r *LoginLogRepository) FindActiveByToken(ctx context.Context, token string) (*loginlog.LoginLog, error) {
	return r.findOneBy(ctx, "l.token = ? AND l.login_log_status = ?", token, enum.LoginLogStatusTypeActive)
}

func (r *LoginLogRepository) FindActiveByUserDevice(ctx context.Context, userId string, deviceUUID string) (*loginlog.LoginLog, error) {
	return r.findOneBy(ctx,
		"l.user_id = ? AND l.device_uuid = ? AND l.login_log_status = ?",
		userId, deviceUUID, enum.LoginLogStatusTypeActive)
}

func (r *LoginLogRepository) ListByUserId(ctx context.Context, userId string) ([]*loginlog.LoginLog, error) {
	args := append(loginLogActiveArgs(), userId)
	query := `SELECT ` + loginLogColumns + ` FROM ` + loginLogTable + ` l WHERE ` +
		loginLogActiveWhere + ` AND l.user_id = ? ORDER BY l.id DESC`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("login_log repo list by user id: %w", err)
	}
	defer rows.Close()

	var logs []*loginlog.LoginLog
	for rows.Next() {
		m, err := scanLoginLog(rows)
		if err != nil {
			return nil, fmt.Errorf("login_log repo scan row: %w", err)
		}
		logs = append(logs, ModelToDomainLoginLog(m))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("login_log repo rows iteration: %w", err)
	}
	return logs, nil
}

func (r *LoginLogRepository) Create(ctx context.Context, l *loginlog.LoginLog) (*loginlog.LoginLog, error) {
	query := `
		INSERT INTO ` + loginLogTable + `
			(login_log_id, user_id, ip_address, device_uuid, token, note, login_log_status)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	result, err := r.db.Exec(ctx, query,
		l.LoginLogId(), l.UserId(), l.IpAddress(), l.DeviceUUID(),
		l.Token(), l.Note(), l.LoginLogStatus())
	if err != nil {
		return nil, fmt.Errorf("login_log repo create: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("login_log repo last insert id: %w", err)
	}
	return r.findBareById(ctx, id)
}

func (r *LoginLogRepository) MarkStatusByLoginLogId(ctx context.Context, loginLogId string, st enum.LoginLogStatusType) error {
	query := `
		UPDATE ` + loginLogTable + `
		SET login_log_status = ?,
			modify_dt        = ?
		WHERE login_log_id = ?
	`
	if _, err := r.db.Exec(ctx, query, st, mtime.Now().Time, loginLogId); err != nil {
		return fmt.Errorf("login_log repo mark status: %w", err)
	}
	return nil
}

// MarkStatusByUserDevice flips every still-ACTIVE row for the (user, device)
// pair. Used at login to enforce the "one active session per device" rule —
// the new login_log row is inserted right after.
func (r *LoginLogRepository) MarkStatusByUserDevice(ctx context.Context, userId string, deviceUUID string, st enum.LoginLogStatusType) error {
	query := `
		UPDATE ` + loginLogTable + `
		SET login_log_status = ?,
			modify_dt        = ?
		WHERE user_id = ?
		  AND device_uuid = ?
		  AND login_log_status = ?
	`
	if _, err := r.db.Exec(ctx, query, st, mtime.Now().Time, userId, deviceUUID, enum.LoginLogStatusTypeActive); err != nil {
		return fmt.Errorf("login_log repo mark status by user device: %w", err)
	}
	return nil
}

func (r *LoginLogRepository) SoftDeleteByLoginLogId(ctx context.Context, loginLogId string) error {
	query := `
		UPDATE ` + loginLogTable + `
		SET login_log_status = ?,
			status           = ?,
			deleted_dt       = ?
		WHERE login_log_id = ?
	`
	if _, err := r.db.Exec(ctx, query,
		enum.LoginLogStatusTypeDeleted, enum.StatusInactive, mtime.Now().Time, loginLogId); err != nil {
		return fmt.Errorf("login_log repo soft delete: %w", err)
	}
	return nil
}

func ModelToDomainLoginLog(m *models.LoginLogModel) *loginlog.LoginLog {
	loginLogId, err := utils.StringToUUID(m.LoginLogId)
	if err != nil {
		return nil
	}

	userId, err := utils.StringToUUID(m.UserId)
	if err != nil {
		return nil
	}

	createId, err := utils.PtrStringToUUID(m.CreateId)
	if err != nil {
		return nil
	}

	modifyId, err := utils.PtrStringToUUID(m.ModifyId)
	if err != nil {
		return nil
	}

	l := loginlog.NewLoginLog()
	l.SetId(m.Id)
	l.SetLoginLogId(loginLogId)
	l.SetUserId(userId)
	l.SetIpAddress(m.IpAddress)
	l.SetDeviceUUID(m.DeviceUUID)
	l.SetToken(m.Token)
	l.SetNote(m.Note)
	l.SetLoginLogStatus(m.LoginLogStatus)
	l.SetStatus(m.Status)
	l.SetCreateId(&createId)
	l.SetCreateDt(mtime.MathTime{Time: m.CreateDt})
	l.SetModifyId(&modifyId)
	l.SetModifyDt(mtime.MathTime{Time: m.ModifyDt})
	return l
}
