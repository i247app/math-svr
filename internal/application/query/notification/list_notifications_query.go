package query

import (
	"context"

	"math-ai.com/math-ai/internal/domain/notification"
	"math-ai.com/math-ai/internal/shared/pagination"
)

// ListNotificationsQuery lists the caller's notifications, most-recent first,
// scoped to UserID for tenant isolation. OnlyUnread restricts to unread rows.
type ListNotificationsQuery struct {
	UserID     int64
	OnlyUnread bool
	Page       int64
	Limit      int64
}

type ListNotificationsQueryHandler struct {
	repo notification.IRepository
}

func NewListNotificationsQueryHandler(repo notification.IRepository) *ListNotificationsQueryHandler {
	return &ListNotificationsQueryHandler{repo: repo}
}

func (h *ListNotificationsQueryHandler) Handle(ctx context.Context, q ListNotificationsQuery) ([]*notification.Notification, *pagination.Pagination, error) {
	return h.repo.ListByUserId(ctx, &notification.ListNotificationsParams{
		UserID:     q.UserID,
		OnlyUnread: q.OnlyUnread,
		Page:       q.Page,
		Limit:      q.Limit,
	})
}
