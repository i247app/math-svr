package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"math-ai.com/math-ai/internal/domain/presence"
	"math-ai.com/math-ai/internal/domain/shared/mtime"
	"math-ai.com/math-ai/internal/infrastructure/database"
	"math-ai.com/math-ai/internal/infrastructure/persistence/mysql/models"
	"math-ai.com/math-ai/internal/shared/enum"
)

const (
	presenceTable = "ma_user_presence"

	presenceColumns = `p.id, p.user_id, p.presence_state, p.connection_count,
		p.last_online_dt, p.last_seen_dt, p.last_device_uuid, p.last_platform,
		p.note, p.status, p.create_id, p.create_dt, p.modify_id, p.modify_dt`

	// Presence has no business-status column — a user is never "soft-deleted"
	// from presence, the row simply goes OFFLINE. Only the system status and
	// the soft-delete stamp are filtered.
	presenceActiveWhere = `p.status = ? AND p.deleted_dt IS NULL`
)

func presenceActiveArgs() []any {
	return []any{enum.StatusActive}
}

type PresenceRepository struct {
	db database.Executor
}

func NewPresenceRepository(db database.Executor) presence.IRepository {
	return &PresenceRepository{db: db}
}

func scanPresence(s database.RowScanner) (*models.PresenceModel, error) {
	var m models.PresenceModel
	if err := s.Scan(&m.Id, &m.UserId, &m.PresenceState, &m.ConnectionCount,
		&m.LastOnlineDt, &m.LastSeenDt, &m.LastDeviceUuid, &m.LastPlatform,
		&m.Note, &m.Status, &m.CreateId, &m.CreateDt, &m.ModifyId,
		&m.ModifyDt); err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *PresenceRepository) FindByUserId(ctx context.Context, userId int64) (*presence.Presence, error) {
	args := append(presenceActiveArgs(), userId)
	query := `SELECT ` + presenceColumns + ` FROM ` + presenceTable + ` p WHERE ` +
		presenceActiveWhere + ` AND p.user_id = ?`

	m, err := scanPresence(r.db.QueryRow(ctx, query, args...))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("presence repo find by user id: %w", err)
	}
	return ModelToDomainPresence(m), nil
}

func (r *PresenceRepository) ListByUserIds(ctx context.Context, userIds []int64) (map[int64]*presence.Presence, error) {
	out := make(map[int64]*presence.Presence, len(userIds))
	if len(userIds) == 0 {
		return out, nil
	}

	placeholders := strings.Repeat("?,", len(userIds)-1) + "?"
	args := presenceActiveArgs()
	for _, id := range userIds {
		args = append(args, id)
	}

	query := `SELECT ` + presenceColumns + ` FROM ` + presenceTable + ` p WHERE ` +
		presenceActiveWhere + ` AND p.user_id IN (` + placeholders + `)`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("presence repo list by user ids: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		m, err := scanPresence(rows)
		if err != nil {
			return nil, fmt.Errorf("presence repo scan: %w", err)
		}
		out[m.UserId] = ModelToDomainPresence(m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("presence repo list rows: %w", err)
	}
	return out, nil
}

// IncrementConnection is an upsert because the row only comes into existence on
// the user's first ever connection. The UNIQUE key on user_id is what makes
// ON DUPLICATE KEY fire; two devices connecting at the same instant therefore
// serialise on the row lock instead of both inserting.
//
// The update clause repeats the placeholders rather than using VALUES(col),
// which MySQL 8.0.20 deprecated.
func (r *PresenceRepository) IncrementConnection(ctx context.Context, userId int64, deviceUuid, platform *string, now mtime.MathTime) (*presence.Presence, error) {
	query := `INSERT INTO ` + presenceTable + `
		  (user_id, presence_state, connection_count, last_online_dt, last_seen_dt,
		   last_device_uuid, last_platform, status)
		VALUES (?, ?, 1, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
		  connection_count = connection_count + 1,
		  presence_state   = ?,
		  last_online_dt   = ?,
		  last_seen_dt     = ?,
		  last_device_uuid = ?,
		  last_platform    = ?`

	online := string(enum.PresenceStateOnline)
	if _, err := r.db.Exec(ctx, query,
		userId, online, now.Time, now.Time, deviceUuid, platform, enum.StatusActive,
		online, now.Time, now.Time, deviceUuid, platform,
	); err != nil {
		return nil, fmt.Errorf("presence repo increment connection: %w", err)
	}

	return r.FindByUserId(ctx, userId)
}

// DecrementConnection clamps at zero so a double-disconnect (or a counter that
// drifted during an unclean shutdown) can never push the column negative.
//
// The assignment order is load-bearing: MySQL evaluates SET clauses left to
// right and later clauses see the ALREADY-UPDATED value of an earlier column.
// presence_state must therefore be decided BEFORE connection_count is
// decremented, or the CASE would read the new count and never fire. Do not
// reorder these two lines.
func (r *PresenceRepository) DecrementConnection(ctx context.Context, userId int64, now mtime.MathTime) (*presence.Presence, error) {
	query := `UPDATE ` + presenceTable + ` SET
		  presence_state   = CASE WHEN connection_count <= 1 THEN ? ELSE presence_state END,
		  connection_count = GREATEST(0, connection_count - 1),
		  last_seen_dt     = ?
		WHERE user_id = ?`

	if _, err := r.db.Exec(ctx, query, string(enum.PresenceStateOffline), now.Time, userId); err != nil {
		return nil, fmt.Errorf("presence repo decrement connection: %w", err)
	}

	return r.FindByUserId(ctx, userId)
}

// ResetAll deliberately leaves last_seen_dt untouched. For a process that died
// uncleanly we do not know when those connections actually ended, so the last
// value we did observe is more honest than stamping the boot time.
func (r *PresenceRepository) ResetAll(ctx context.Context) error {
	query := `UPDATE ` + presenceTable + `
		SET connection_count = 0, presence_state = ?
		WHERE connection_count <> 0 OR presence_state <> ?`

	offline := string(enum.PresenceStateOffline)
	if _, err := r.db.Exec(ctx, query, offline, offline); err != nil {
		return fmt.Errorf("presence repo reset all: %w", err)
	}
	return nil
}

func ModelToDomainPresence(m *models.PresenceModel) *presence.Presence {
	p := presence.NewPresence()
	p.SetId(m.Id)
	p.SetUserId(m.UserId)
	p.SetPresenceState(m.PresenceState)
	p.SetConnectionCount(m.ConnectionCount)
	if m.LastOnlineDt != nil {
		p.SetLastOnlineDt(mtime.MathTime{Time: *m.LastOnlineDt})
	}
	if m.LastSeenDt != nil {
		p.SetLastSeenDt(mtime.MathTime{Time: *m.LastSeenDt})
	}
	p.SetLastDeviceUuid(m.LastDeviceUuid)
	p.SetLastPlatform(m.LastPlatform)
	p.SetNote(m.Note)
	p.SetStatus(m.Status)
	p.SetCreateId(m.CreateId)
	p.SetCreateDt(mtime.MathTime{Time: m.CreateDt})
	p.SetModifyId(m.ModifyId)
	p.SetModifyDt(mtime.MathTime{Time: m.ModifyDt})
	return p
}
