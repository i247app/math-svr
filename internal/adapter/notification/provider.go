package notification

import "context"

// NotificationProvider is the contract every push vendor implements. Today
// the only implementation is Firebase Cloud Messaging.
type NotificationProvider interface {
	Name() NotificationProviderName
	Send(ctx context.Context, msg PushMessage) (*SendResult, error)
}
