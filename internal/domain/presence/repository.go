package presence

import (
	"context"

	"math-ai.com/math-ai/internal/domain/shared/mtime"
)

// IRepository owns ma_user_presence.
//
// The two mutating methods are deliberately not a generic Update: presence is
// written concurrently from every connect and disconnect across every device,
// so read-modify-write from Go would lose increments. Both are expressed as a
// single atomic statement and return the resulting row, letting the caller see
// whether the write crossed the online/offline boundary without a second read.
type IRepository interface {
	FindByUserId(ctx context.Context, userId int64) (*Presence, error)

	// ListByUserIds batches the lookup for a member list. Users with no row
	// are simply absent from the map — callers must treat a missing key as
	// OFFLINE, not as an error, because a row only appears after the user's
	// first ever connection.
	ListByUserIds(ctx context.Context, userIds []int64) (map[int64]*Presence, error)

	// IncrementConnection upserts the row, adds one connection and marks the
	// user ONLINE.
	IncrementConnection(ctx context.Context, userId int64, deviceUuid, platform *string, now mtime.MathTime) (*Presence, error)

	// DecrementConnection removes one connection, clamped at zero, and flips
	// the row to OFFLINE when the last one goes away.
	DecrementConnection(ctx context.Context, userId int64, now mtime.MathTime) (*Presence, error)

	// ResetAll zeroes every counter and marks everyone OFFLINE. Called once at
	// boot: the Hub's registry is process memory, so after a restart or crash
	// every non-zero counter in the table describes connections that no longer
	// exist. Without this, those users stay green forever.
	ResetAll(ctx context.Context) error
}
