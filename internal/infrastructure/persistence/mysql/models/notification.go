package models

import "time"

// NotificationModel mirrors ma_notifications. Pure data; the repository maps
// it to the domain Notification via ModelToDomainNotification.
type NotificationModel struct {
	Id                 int64
	NotificationId     int64
	UserId             int64
	Title              string
	ShortText          string
	Category           *string
	IsRead             bool
	ActionType         *string
	ActionData         *string
	Priority           *string
	Note               *string
	NotificationStatus *string
	Status             string
	CreateId           *int64
	CreateDt           time.Time
	ModifyId           *int64
	ModifyDt           time.Time
}
