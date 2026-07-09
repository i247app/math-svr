package enum

// NotificationStatusType is the business lifecycle of a notification row
// (the `notification_status` column), mirroring the dual-status pattern used
// by classroom/conversation. ACTIVE is the live state; ARCHIVED is hidden
// from default listings but recoverable; DELETED is the soft-delete tombstone
// hidden by *ActiveWhere filters.
type NotificationStatusType string

const (
	NotificationStatusTypeActive   NotificationStatusType = "ACTIVE"
	NotificationStatusTypeArchived NotificationStatusType = "ARCHIVED"
	NotificationStatusTypeDeleted  NotificationStatusType = "DELETED"
)

func (s NotificationStatusType) String() string {
	return string(s)
}

func (s NotificationStatusType) IsValid() bool {
	switch s {
	case NotificationStatusTypeActive, NotificationStatusTypeArchived, NotificationStatusTypeDeleted:
		return true
	default:
		return false
	}
}

// NotificationPriorityType is the delivery priority of a notification (the
// `priority` column). NORMAL is the default.
type NotificationPriorityType string

const (
	NotificationPriorityTypeLow    NotificationPriorityType = "LOW"
	NotificationPriorityTypeNormal NotificationPriorityType = "NORMAL"
	NotificationPriorityTypeHigh   NotificationPriorityType = "HIGH"
)

func (p NotificationPriorityType) String() string {
	return string(p)
}

func (p NotificationPriorityType) IsValid() bool {
	switch p {
	case NotificationPriorityTypeLow, NotificationPriorityTypeNormal, NotificationPriorityTypeHigh:
		return true
	default:
		return false
	}
}

type NotificationCategoryType string

const (
	NotificationCategoryTypeInfo    NotificationCategoryType = "INFO"
	NotificationCategoryTypeWarning NotificationCategoryType = "WARNING"
	NotificationCategoryTypeError   NotificationCategoryType = "ERROR"
	NotificationCategoryTypeSuccess NotificationCategoryType = "SUCCESS"
)

func (c NotificationCategoryType) String() string {
	return string(c)
}

func (c NotificationCategoryType) IsValid() bool {
	switch c {
	case NotificationCategoryTypeInfo, NotificationCategoryTypeWarning, NotificationCategoryTypeError, NotificationCategoryTypeSuccess:
		return true
	default:
		return false
	}
}
