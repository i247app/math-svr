package notification

import (
	"encoding/json"

	domain "math-ai.com/math-ai/internal/domain/notification"
	"math-ai.com/math-ai/internal/shared/pagination"
)

// SendNotificationReq creates a notification for a recipient (UserID) and
// pushes it to that user's devices. Used by the admin/test endpoint and as the
// wire shape other modules can map onto when calling the service in-process.
type SendNotificationReq struct {
	UserID     int64           `json:"user_id"`
	Title      string          `json:"title"`
	ShortText  string          `json:"short_text"`
	Category   *string         `json:"category"`
	ActionType *string         `json:"action_type"`
	ActionData json.RawMessage `json:"action_data"`
	Priority   *string         `json:"priority"`
	Note       *string         `json:"note"`

	// CreatorUID is the acting user from the session (audit create_id). Never
	// read from the request body.
	CreatorUID *int64 `json:"-"`
}

// SendNotificationRes returns the persisted notification plus a summary of
// the push fan-out (zeroed when the recipient has no device tokens or the
// adapter is disabled).
type SendNotificationRes struct {
	Notification NotificationResponse `json:"notification"`
	PushSuccess  int                  `json:"push_success"`
	PushFailure  int                  `json:"push_failure"`
}

// ListNotificationsReq lists the caller's notifications. UserID is
// session-injected.
type ListNotificationsReq struct {
	Page       int    `json:"page"`
	Size       int    `json:"size"`
	OnlyUnread bool   `json:"only_unread"`
	UserID     *int64 `json:"-"`
}

type ListNotificationsRes struct {
	Notifications []NotificationResponse `json:"notifications"`
	Pagination    *pagination.Pagination `json:"pagination"`
}

// UnreadCountReq / Res report the caller's unread notification count.
type UnreadCountReq struct {
	UserID *int64 `json:"-"`
}

type UnreadCountRes struct {
	Count int64 `json:"count"`
}

// MarkReadReq marks a single owned notification read.
type MarkReadReq struct {
	NotificationID int64  `json:"notification_id"`
	UserID         *int64 `json:"-"`
}

// MarkAllReadReq marks every unread notification of the caller read.
type MarkAllReadReq struct {
	UserID *int64 `json:"-"`
}

// DeleteNotificationReq soft-deletes an owned notification.
type DeleteNotificationReq struct {
	NotificationID int64  `json:"notification_id"`
	UserID         *int64 `json:"-"`
}

// NotificationResponse is the wire shape for one notification row.
type NotificationResponse struct {
	NotificationID int64           `json:"notification_id"`
	UserID         int64           `json:"user_id"`
	Title          string          `json:"title"`
	ShortText      string          `json:"short_text"`
	Category       *string         `json:"category,omitempty"`
	IsRead         bool            `json:"is_read"`
	ActionType     *string         `json:"action_type,omitempty"`
	ActionData     json.RawMessage `json:"action_data,omitempty"`
	Priority       *string         `json:"priority,omitempty"`
	CreateDt       string          `json:"create_dt"`
	ModifyDt       string          `json:"modify_dt"`
}

func DomainToResponse(n *domain.Notification) NotificationResponse {
	var actionData json.RawMessage
	if n.ActionData() != nil {
		actionData = json.RawMessage(*n.ActionData())
	}
	return NotificationResponse{
		NotificationID: n.NotificationId(),
		UserID:         n.UserId(),
		Title:          n.Title(),
		ShortText:      n.ShortText(),
		Category:       n.Category(),
		IsRead:         n.IsRead(),
		ActionType:     n.ActionType(),
		ActionData:     actionData,
		Priority:       n.Priority(),
		CreateDt:       n.CreateDt().String(),
		ModifyDt:       n.ModifyDt().String(),
	}
}

func DomainListToResponse(ns []*domain.Notification) []NotificationResponse {
	out := make([]NotificationResponse, 0, len(ns))
	for _, n := range ns {
		out = append(out, DomainToResponse(n))
	}
	return out
}
