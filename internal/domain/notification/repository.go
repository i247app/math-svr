package notification

import (
	"context"

	"math-ai.com/math-ai/internal/shared/pagination"
)

// ListNotificationsParams narrows the per-user notification listing. Results
// are scoped to UserID (ownership isolation) and ordered most-recent-first.
// OnlyUnread restricts the list to is_read = false rows.
type ListNotificationsParams struct {
	UserID     int64
	OnlyUnread bool
	Page       int64
	Limit      int64
}

// IRepository owns ma_notifications persistence. All reads exclude
// soft-deleted/inactive rows.
type IRepository interface {
	FindByNotificationId(ctx context.Context, notificationId int64) (*Notification, error)
	ListByUserId(ctx context.Context, params *ListNotificationsParams) ([]*Notification, *pagination.Pagination, error)
	CountUnreadByUserId(ctx context.Context, userId int64) (int64, error)
	Create(ctx context.Context, n *Notification) (*Notification, error)
	MarkReadByNotificationId(ctx context.Context, notificationId int64) error
	MarkAllReadByUserId(ctx context.Context, userId int64) error
	SoftDeleteByNotificationId(ctx context.Context, notificationId int64) error
}
