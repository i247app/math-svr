package notification

type NotificationProvider interface {
	Name() NotificationProviderName
}
