package query

import (
	"context"

	"math-ai.com/math-ai/internal/domain/notification"
)

// UnreadCountQuery returns the caller's unread notification count.
type UnreadCountQuery struct {
	UID int64
}

type UnreadCountQueryHandler struct {
	repo notification.IRepository
}

func NewUnreadCountQueryHandler(repo notification.IRepository) *UnreadCountQueryHandler {
	return &UnreadCountQueryHandler{repo: repo}
}

func (h *UnreadCountQueryHandler) Handle(ctx context.Context, q UnreadCountQuery) (int64, error) {
	return h.repo.CountUnreadByUid(ctx, q.UID)
}
