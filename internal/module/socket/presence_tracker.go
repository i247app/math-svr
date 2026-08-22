package socket

import "context"

// PresenceTracker records a user's live-connection count as connections open
// and close. It is an interface here, rather than a direct dependency on
// module/presence, for the same reason Authorizer is: the socket module owns
// transport, and must not import a sibling module.
//
// A nil tracker is valid and disables presence tracking — the realtime channel
// works without it.
type PresenceTracker interface {
	// MarkOnline reports whether this connection took the user from offline
	// to online.
	MarkOnline(ctx context.Context, userId int64, deviceUuid, platform *string) (bool, error)
	// MarkOffline reports whether this disconnect was the user's last.
	MarkOffline(ctx context.Context, userId int64) (bool, error)
}
