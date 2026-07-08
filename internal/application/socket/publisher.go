// Package socket defines the application-layer port that producers use to push
// realtime events, decoupling them from the concrete transport. The in-memory
// implementation is infrastructure/socket.Hub (via HubPublisher); a Redis-backed
// implementation can replace it later without touching any producer.
package socket

import "context"

// Publisher is the write side of the realtime channel. Producers (notification,
// classroom, …) depend on this interface, obtained from the app Resource and
// always nil-guarded — the socket runtime is a deploy-profile concern like the
// other adapters.
//
// ctx carries the request deadline/cancellation and (for a future networked
// implementation) bounds the publish call; the in-memory Hub ignores it and
// never returns an error.
type Publisher interface {
	// Publish fans event+data out to every connection subscribed to topic.
	Publish(ctx context.Context, topic, event string, data any) error

	// BroadcastUser fans event+data out to every connection owned by userID,
	// independent of topic subscription (direct, user-addressed push).
	BroadcastUser(ctx context.Context, userID int64, event string, data any) error
}
