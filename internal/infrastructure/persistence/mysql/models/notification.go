package models

// NotificationModel mirrors ma_notifications. Pure data; the repository maps
// it to the domain Notification via ModelToDomainNotification.
type NotificationModel struct {
	Audit

	NotificationId     int64
	Uid                int64
	Title              string
	ShortText          string
	Category           *string
	IsRead             bool
	ActionType         *string
	ActionData         *string
	Priority           *string
	NotificationStatus *string
}
