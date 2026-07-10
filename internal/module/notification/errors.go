package notification

import "errors"

// Module-scoped sentinel errors. See conventions.md §Errors.
var (
	ErrUserNotFound           = errors.New("user not found")
	ErrUidRequired            = errors.New("uid is required")
	ErrTitleRequired          = errors.New("title is required")
	ErrShortTextRequired      = errors.New("short_text is required")
	ErrInvalidPriority        = errors.New("priority must be one of LOW, NORMAL, HIGH")
	ErrInvalidActionData      = errors.New("action_data must be valid JSON")
	ErrNotificationIDInvalid  = errors.New("notification_id must be positive")
	ErrUidNotFoundFromSession = errors.New("uid not found from session")
)
