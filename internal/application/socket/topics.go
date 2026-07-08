package socket

import "fmt"

// Topic-name builders are the single source of truth for the topic strings
// shared between the socket module (which auto-subscribes / authorizes) and
// producers (which publish). Keeping them here prevents the subscriber and the
// publisher from drifting on the format.

// UserTopic is a user's personal address, auto-subscribed on connect.
func UserTopic(userID int64) string { return fmt.Sprintf("user:%d", userID) }

// NotificationsTopic carries in-app notification events for a user.
func NotificationsTopic(userID int64) string { return fmt.Sprintf("notifications:%d", userID) }
