package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"math-ai.com/math-ai/internal/domain/notification"
	"math-ai.com/math-ai/internal/domain/shared/mtime"
	"math-ai.com/math-ai/internal/infrastructure/database"
	"math-ai.com/math-ai/internal/infrastructure/persistence/mysql/models"
	"math-ai.com/math-ai/internal/shared/enum"
	"math-ai.com/math-ai/internal/shared/pagination"
)

const (
	notificationTable = "ma_notifications"

	notificationColumns = `n.id, n.notification_id, n.user_id, n.title, n.short_text,
		n.category, n.is_read, n.action_type, n.action_data, n.priority, n.note,
		n.notification_status, n.status,
		n.create_id, n.create_dt, n.modify_id, n.modify_dt`

	// Reads exclude both system-inactive rows and the DELETED business
	// status used by the soft-delete path. ARCHIVED rows remain visible
	// (recoverable) — callers filter them out at the query level if needed.
	notificationActiveWhere = `n.status = ? AND n.deleted_dt IS NULL
		AND (n.notification_status IS NULL OR n.notification_status != ?)`
)

func notificationActiveArgs() []any {
	return []any{enum.StatusActive, enum.NotificationStatusTypeDeleted}
}

type NotificationRepository struct {
	db database.Executor
}

func NewNotificationRepository(db database.Executor) notification.IRepository {
	return &NotificationRepository{db: db}
}

func scanNotification(s database.RowScanner) (*models.NotificationModel, error) {
	var m models.NotificationModel
	if err := s.Scan(&m.Id, &m.NotificationId, &m.UserId, &m.Title, &m.ShortText,
		&m.Category, &m.IsRead, &m.ActionType, &m.ActionData, &m.Priority, &m.Note,
		&m.NotificationStatus, &m.Status,
		&m.CreateId, &m.CreateDt, &m.ModifyId, &m.ModifyDt); err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *NotificationRepository) findOneBy(ctx context.Context, where string, args ...any) (*notification.Notification, error) {
	fullArgs := append(notificationActiveArgs(), args...)
	query := `SELECT ` + notificationColumns + ` FROM ` + notificationTable + ` n WHERE ` +
		notificationActiveWhere + ` AND (` + where + `)`

	m, err := scanNotification(r.db.QueryRow(ctx, query, fullArgs...))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("notification repo find (%s): %w", where, err)
	}
	return ModelToDomainNotification(m), nil
}

func (r *NotificationRepository) findBareById(ctx context.Context, id int64) (*notification.Notification, error) {
	args := append(notificationActiveArgs(), id)
	query := `SELECT ` + notificationColumns + ` FROM ` + notificationTable + ` n WHERE ` +
		notificationActiveWhere + ` AND n.id = ?`

	m, err := scanNotification(r.db.QueryRow(ctx, query, args...))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("notification repo find bare by id: %w", err)
	}
	return ModelToDomainNotification(m), nil
}

func (r *NotificationRepository) FindByNotificationId(ctx context.Context, notificationId int64) (*notification.Notification, error) {
	return r.findOneBy(ctx, "n.notification_id = ?", notificationId)
}

func (r *NotificationRepository) ListByUserId(ctx context.Context, params *notification.ListNotificationsParams) ([]*notification.Notification, *pagination.Pagination, error) {
	if params == nil {
		return nil, nil, fmt.Errorf("notification repo list: params is required")
	}

	filter := ` AND n.user_id = ?`
	filterArgs := []any{params.UserID}
	if params.OnlyUnread {
		filter += ` AND n.is_read = ?`
		filterArgs = append(filterArgs, false)
	}

	countArgs := append(notificationActiveArgs(), filterArgs...)
	countQuery := `SELECT COUNT(*) FROM ` + notificationTable + ` n WHERE ` +
		notificationActiveWhere + filter

	var total int64
	if err := r.db.QueryRow(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, nil, fmt.Errorf("notification repo count: %w", err)
	}

	pg := pagination.NewPagination(params.Page, params.Limit, total)

	listArgs := append(notificationActiveArgs(), filterArgs...)
	listArgs = append(listArgs, pg.Size, pg.Skip)
	query := `SELECT ` + notificationColumns + ` FROM ` + notificationTable + ` n WHERE ` +
		notificationActiveWhere + filter +
		` ORDER BY n.id DESC LIMIT ? OFFSET ?`

	rows, err := r.db.Query(ctx, query, listArgs...)
	if err != nil {
		return nil, nil, fmt.Errorf("notification repo list: %w", err)
	}
	defer rows.Close()

	var notifications []*notification.Notification
	for rows.Next() {
		m, err := scanNotification(rows)
		if err != nil {
			return nil, nil, fmt.Errorf("notification repo scan row: %w", err)
		}
		notifications = append(notifications, ModelToDomainNotification(m))
	}
	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("notification repo rows iteration: %w", err)
	}
	return notifications, pg, nil
}

func (r *NotificationRepository) CountUnreadByUserId(ctx context.Context, userId int64) (int64, error) {
	args := append(notificationActiveArgs(), userId, false)
	query := `SELECT COUNT(*) FROM ` + notificationTable + ` n WHERE ` +
		notificationActiveWhere + ` AND n.user_id = ? AND n.is_read = ?`

	var total int64
	if err := r.db.QueryRow(ctx, query, args...).Scan(&total); err != nil {
		return 0, fmt.Errorf("notification repo count unread: %w", err)
	}
	return total, nil
}

func (r *NotificationRepository) Create(ctx context.Context, n *notification.Notification) (*notification.Notification, error) {
	query := `
		INSERT INTO ` + notificationTable + `
			(notification_id, user_id, title, short_text, category, is_read, action_type,
			 action_data, priority, note, notification_status, create_id, create_dt, modify_dt)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	now := mtime.Now().Time
	result, err := r.db.Exec(ctx, query,
		n.NotificationId(), n.UserId(), n.Title(), n.ShortText(), n.Category(),
		n.IsRead(), n.ActionType(), n.ActionData(), n.Priority(), n.Note(),
		n.NotificationStatus(), n.CreateId(), now, now)
	if err != nil {
		return nil, fmt.Errorf("notification repo create: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("notification repo last insert id: %w", err)
	}
	return r.findBareById(ctx, id)
}

func (r *NotificationRepository) MarkReadByNotificationId(ctx context.Context, notificationId int64) error {
	query := `
		UPDATE ` + notificationTable + `
		SET is_read   = TRUE,
			modify_dt = ?
		WHERE notification_id = ?
	`
	if _, err := r.db.Exec(ctx, query, mtime.Now().Time, notificationId); err != nil {
		return fmt.Errorf("notification repo mark read: %w", err)
	}
	return nil
}

func (r *NotificationRepository) MarkAllReadByUserId(ctx context.Context, userId int64) error {
	query := `
		UPDATE ` + notificationTable + `
		SET is_read   = TRUE,
			modify_dt = ?
		WHERE uid = ? AND is_read = FALSE
	`
	if _, err := r.db.Exec(ctx, query, mtime.Now().Time, userId); err != nil {
		return fmt.Errorf("notification repo mark all read: %w", err)
	}
	return nil
}

func (r *NotificationRepository) SoftDeleteByNotificationId(ctx context.Context, notificationId int64) error {
	query := `
		UPDATE ` + notificationTable + `
		SET notification_status = ?,
			status              = ?,
			deleted_dt          = ?,
			modify_dt           = ?
		WHERE notification_id = ?
	`
	now := mtime.Now().Time
	if _, err := r.db.Exec(ctx, query,
		enum.NotificationStatusTypeDeleted, enum.StatusInactive, now, now, notificationId); err != nil {
		return fmt.Errorf("notification repo soft delete: %w", err)
	}
	return nil
}

func ModelToDomainNotification(m *models.NotificationModel) *notification.Notification {
	n := notification.NewNotification()
	n.SetId(m.Id)
	n.SetNotificationId(m.NotificationId)
	n.SetUserId(m.UserId)
	n.SetTitle(m.Title)
	n.SetShortText(m.ShortText)
	n.SetCategory(m.Category)
	n.SetIsRead(m.IsRead)
	n.SetActionType(m.ActionType)
	n.SetActionData(m.ActionData)
	n.SetPriority(m.Priority)
	n.SetNote(m.Note)
	n.SetNotificationStatus(m.NotificationStatus)
	n.SetStatus(m.Status)
	n.SetCreateId(m.CreateId)
	n.SetCreateDt(mtime.MathTime{Time: m.CreateDt})
	n.SetModifyId(m.ModifyId)
	n.SetModifyDt(mtime.MathTime{Time: m.ModifyDt})
	return n
}
