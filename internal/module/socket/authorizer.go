package socket

import appsocket "math-ai.com/math-ai/internal/application/socket"

// Authorizer decides whether a user may subscribe to a topic. It is injected
// into the Service so topic policy can grow (classroom membership, AI-thread
// ownership, …) without touching the transport layer.
type Authorizer interface {
	CanSubscribe(userID int64, topic string) bool
}

// DefaultAuthorizer permits only a user's own personal topics. Resource topics
// (e.g. classroom:{id}, ai:conversation:{id}) are denied until a membership-
// aware authorizer is wired in a later phase — deny-by-default is the safe base.
type DefaultAuthorizer struct{}

func (DefaultAuthorizer) CanSubscribe(userID int64, topic string) bool {
	switch topic {
	case personalTopic(userID), notificationsTopic(userID):
		return true
	default:
		return false
	}
}

// personalTopic / notificationsTopic delegate to the shared builders in
// application/socket so the subscriber (here) and publishers (producers) never
// disagree on the topic string.
func personalTopic(userID int64) string      { return appsocket.UserTopic(userID) }
func notificationsTopic(userID int64) string { return appsocket.NotificationsTopic(userID) }
