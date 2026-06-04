package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"math-ai.com/math-ai/internal/domain/otp"
	mtime "math-ai.com/math-ai/internal/domain/shared/time"
	"math-ai.com/math-ai/internal/infrastructure/database"
	"math-ai.com/math-ai/internal/infrastructure/persistence/mysql/models"
	"math-ai.com/math-ai/internal/shared/enum"
)

const (
	otpTable = "ma_otps"

	otpColumns = `o.id, o.otp_id, o.otp_type, o.user_id, o.identifier,
		o.device_uuid, o.device_name, o.otp_code, o.otp_create_dt, o.otp_expire_dt,
		o.attempt_count, o.note, o.otp_status, o.status,
		o.create_id, o.create_dt, o.modify_id, o.modify_dt`

	otpActiveWhere = `o.status IN (?) AND o.deleted_dt IS NULL`
)

func otpActiveArgs() []any {
	return []any{enum.StatusActive}
}

type OtpRepository struct {
	db database.Executor
}

func NewOtpRepository(db database.Executor) otp.IRepository {
	return &OtpRepository{db: db}
}

func scanOtp(s database.RowScanner) (*models.OtpModel, error) {
	var m models.OtpModel
	if err := s.Scan(&m.Id, &m.OtpId, &m.OtpType, &m.UserId, &m.Identifier,
		&m.DeviceUUID, &m.DeviceName, &m.OtpCode, &m.OtpCreateDt, &m.OtpExpireDt,
		&m.AttemptCount, &m.Note, &m.OtpStatus, &m.Status,
		&m.CreateId, &m.CreateDt, &m.ModifyId, &m.ModifyDt); err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *OtpRepository) findOneBy(ctx context.Context, where string, args ...any) (*otp.Otp, error) {
	fullArgs := append(otpActiveArgs(), args...)
	query := `SELECT ` + otpColumns + ` FROM ` + otpTable + ` o WHERE ` +
		otpActiveWhere + ` AND (` + where + `)`

	m, err := scanOtp(r.db.QueryRow(ctx, query, fullArgs...))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("otp repo find (%s): %w", where, err)
	}
	return ModelToDomainOtp(m), nil
}

func (r *OtpRepository) findBareById(ctx context.Context, id int64) (*otp.Otp, error) {
	args := append(otpActiveArgs(), id)
	query := `SELECT ` + otpColumns + ` FROM ` + otpTable + ` o WHERE ` +
		otpActiveWhere + ` AND o.id = ?`

	m, err := scanOtp(r.db.QueryRow(ctx, query, args...))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("otp repo find bare by id: %w", err)
	}
	return ModelToDomainOtp(m), nil
}

func (r *OtpRepository) FindByOtpId(ctx context.Context, otpId int64) (*otp.Otp, error) {
	return r.findOneBy(ctx, "o.otp_id = ?", otpId)
}

// FindLatestPending returns the freshest still-PENDING OTP for the given
// (type, identifier). Order by id DESC since the same row was just inserted
// in the same transaction in most call paths.
func (r *OtpRepository) FindLatestPending(ctx context.Context, otpType enum.OtpType, identifier string) (*otp.Otp, error) {
	args := append(otpActiveArgs(), otpType, identifier, enum.OtpStatusTypePending)
	query := `SELECT ` + otpColumns + ` FROM ` + otpTable + ` o WHERE ` +
		otpActiveWhere + ` AND o.otp_type = ? AND o.identifier = ? AND o.otp_status = ?
		ORDER BY o.id DESC LIMIT 1`

	m, err := scanOtp(r.db.QueryRow(ctx, query, args...))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("otp repo find latest pending: %w", err)
	}
	return ModelToDomainOtp(m), nil
}

func (r *OtpRepository) CountSentSince(ctx context.Context, otpType enum.OtpType, identifier string, since time.Time) (int, error) {
	args := append(otpActiveArgs(), otpType, identifier, since)
	query := `SELECT COUNT(*) FROM ` + otpTable + ` o WHERE ` +
		otpActiveWhere + ` AND o.otp_type = ? AND o.identifier = ? AND o.create_dt >= ?`

	var count int
	if err := r.db.QueryRow(ctx, query, args...).Scan(&count); err != nil {
		return 0, fmt.Errorf("otp repo count sent since: %w", err)
	}
	return count, nil
}

func (r *OtpRepository) Create(ctx context.Context, o *otp.Otp) (*otp.Otp, error) {
	query := `
		INSERT INTO ` + otpTable + `
			(otp_id, otp_type, user_id, identifier, device_uuid, device_name,
			 otp_code, otp_create_dt, otp_expire_dt, attempt_count, note, otp_status, create_dt, modify_dt)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	var createDtArg, expireDtArg any
	if o.OtpCreateDt().IsValid() {
		createDtArg = o.OtpCreateDt()
	}
	if o.OtpExpireDt().IsValid() {
		expireDtArg = o.OtpExpireDt()
	}

	result, err := r.db.Exec(ctx, query,
		o.OtpId(), o.OtpType(), o.UserId(), o.Identifier(), o.DeviceUUID(), o.DeviceName(),
		o.OtpCode(), createDtArg, expireDtArg, o.AttemptCount(), o.Note(), o.OtpStatus(), mtime.Now().Time, mtime.Now().Time)
	if err != nil {
		return nil, fmt.Errorf("otp repo create: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("otp repo last insert id: %w", err)
	}
	return r.findBareById(ctx, id)
}

func (r *OtpRepository) MarkStatusByOtpId(ctx context.Context, otpId int64, st enum.OtpStatusType) error {
	query := `
		UPDATE ` + otpTable + `
		SET otp_status = ?,
			modify_dt  = ?
		WHERE otp_id = ?
	`
	if _, err := r.db.Exec(ctx, query, st, mtime.Now().Time, otpId); err != nil {
		return fmt.Errorf("otp repo mark status: %w", err)
	}
	return nil
}

// RevokePendingByTypeIdentifier mass-revokes prior PENDING rows so a freshly
// inserted OTP is the only PENDING row for (type, identifier). Filters on
// otp_status = PENDING so already-VERIFIED rows stay as-is.
func (r *OtpRepository) RevokePendingByTypeIdentifier(ctx context.Context, otpType enum.OtpType, identifier string) error {
	query := `
		UPDATE ` + otpTable + `
		SET otp_status = ?,
			modify_dt  = ?
		WHERE otp_type = ?
		  AND identifier = ?
		  AND otp_status = ?
	`
	if _, err := r.db.Exec(ctx, query,
		enum.OtpStatusTypeRevoked, mtime.Now().Time,
		otpType, identifier, enum.OtpStatusTypePending,
	); err != nil {
		return fmt.Errorf("otp repo revoke pending: %w", err)
	}
	return nil
}

// IncrementAttemptCount bumps attempt_count atomically and returns the new
// value. The UPDATE..LAST_INSERT_ID trick keeps the read+write a single round
// trip and tolerates concurrent verify attempts: the second one will see the
// already-incremented value through SELECT LAST_INSERT_ID().
func (r *OtpRepository) IncrementAttemptCount(ctx context.Context, otpId int64) (int, error) {
	update := `
		UPDATE ` + otpTable + `
		SET attempt_count = LAST_INSERT_ID(attempt_count + 1),
			modify_dt     = ?
		WHERE otp_id = ?
	`
	if _, err := r.db.Exec(ctx, update, mtime.Now().Time, otpId); err != nil {
		return 0, fmt.Errorf("otp repo increment attempt count: %w", err)
	}

	var n int
	if err := r.db.QueryRow(ctx, `SELECT LAST_INSERT_ID()`).Scan(&n); err != nil {
		return 0, fmt.Errorf("otp repo read attempt count: %w", err)
	}
	return n, nil
}

func ModelToDomainOtp(m *models.OtpModel) *otp.Otp {
	o := otp.NewOtp()
	o.SetId(m.Id)
	o.SetOtpId(m.OtpId)
	o.SetOtpType(m.OtpType)
	o.SetUserId(m.UserId)
	o.SetIdentifier(m.Identifier)
	o.SetDeviceUUID(m.DeviceUUID)
	o.SetDeviceName(m.DeviceName)
	o.SetOtpCode(m.OtpCode)
	if m.OtpCreateDt != nil {
		o.SetOtpCreateDt(mtime.MathTime{Time: *m.OtpCreateDt})
	}
	if m.OtpExpireDt != nil {
		o.SetOtpExpireDt(mtime.MathTime{Time: *m.OtpExpireDt})
	}
	o.SetAttemptCount(m.AttemptCount)
	o.SetNote(m.Note)
	o.SetOtpStatus(m.OtpStatus)
	o.SetStatus(m.Status)
	o.SetCreateId(m.CreateId)
	o.SetCreateDt(mtime.MathTime{Time: m.CreateDt})
	o.SetModifyId(m.ModifyId)
	o.SetModifyDt(mtime.MathTime{Time: m.ModifyDt})
	return o
}
