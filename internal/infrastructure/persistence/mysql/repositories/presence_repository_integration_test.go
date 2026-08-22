package repositories

import (
	"context"
	"sync"
	"testing"

	"math-ai.com/math-ai/internal/domain/shared/mtime"
	"math-ai.com/math-ai/internal/shared/enum"
)

// TestIntegrationPresenceCounterLifecycle exercises the online/offline dot
// against the real engine.
//
// The upsert and the CASE-before-decrement are both MySQL semantics that a
// fake cannot check: the first row must be created by ON DUPLICATE KEY UPDATE,
// and the state flip depends on MySQL evaluating SET clauses left to right.
func TestIntegrationPresenceCounterLifecycle(t *testing.T) {
	db := openTestDB(t)

	userID := uniqueTestID(20)
	cleanupPresenceRows(t, db, userID)

	repo := NewPresenceRepository(db)
	ctx := context.Background()

	// First ever connection: no row exists yet, so this must insert one.
	p, err := repo.IncrementConnection(ctx, userID, nil, nil, mtime.Now())
	if err != nil {
		t.Fatalf("first connect: %v", err)
	}
	if p == nil {
		t.Fatal("first connect did not create a presence row")
	}
	if p.ConnectionCount() != 1 || !p.IsOnline() {
		t.Fatalf("after first connect: count=%d state=%s, want 1/ONLINE", p.ConnectionCount(), p.PresenceState())
	}

	// Second device on the same account.
	p, err = repo.IncrementConnection(ctx, userID, nil, nil, mtime.Now())
	if err != nil {
		t.Fatalf("second connect: %v", err)
	}
	if p.ConnectionCount() != 2 {
		t.Fatalf("after second connect: count=%d, want 2", p.ConnectionCount())
	}

	// Closing one device must NOT take the user offline — they are still
	// reachable on the other one.
	p, err = repo.DecrementConnection(ctx, userID, mtime.Now())
	if err != nil {
		t.Fatalf("first disconnect: %v", err)
	}
	if p.ConnectionCount() != 1 {
		t.Fatalf("after one disconnect: count=%d, want 1", p.ConnectionCount())
	}
	if !p.IsOnline() || p.PresenceState() != string(enum.PresenceStateOnline) {
		t.Errorf("user went offline while a second device was still connected (state=%s)", p.PresenceState())
	}

	// Last device closes.
	p, err = repo.DecrementConnection(ctx, userID, mtime.Now())
	if err != nil {
		t.Fatalf("last disconnect: %v", err)
	}
	if p.ConnectionCount() != 0 {
		t.Fatalf("after last disconnect: count=%d, want 0", p.ConnectionCount())
	}
	if p.IsOnline() || p.PresenceState() != string(enum.PresenceStateOffline) {
		t.Errorf("state = %s, want OFFLINE once the last connection went away", p.PresenceState())
	}
	if p.LastSeenDt().Time.IsZero() {
		t.Error("last_seen_dt was not stamped — 'hoạt động N phút trước' would have nothing to show")
	}

	// A stray extra disconnect (unclean shutdown, double close) must clamp.
	p, err = repo.DecrementConnection(ctx, userID, mtime.Now())
	if err != nil {
		t.Fatalf("extra disconnect: %v", err)
	}
	if p.ConnectionCount() != 0 {
		t.Errorf("count = %d after an extra disconnect, want it clamped at 0", p.ConnectionCount())
	}
}

// Connections open concurrently from several devices. The counter is a
// read-modify-write in SQL, so if the statement were not atomic some
// increments would be lost and the user would appear offline while connected.
func TestIntegrationPresenceConcurrentConnects(t *testing.T) {
	db := openTestDB(t)

	userID := uniqueTestID(21)
	cleanupPresenceRows(t, db, userID)

	repo := NewPresenceRepository(db)
	const devices = 15

	var wg sync.WaitGroup
	errCh := make(chan error, devices)
	for i := 0; i < devices; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := repo.IncrementConnection(context.Background(), userID, nil, nil, mtime.Now()); err != nil {
				errCh <- err
			}
		}()
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		t.Fatalf("concurrent connect: %v", err)
	}

	p, err := repo.FindByUserId(context.Background(), userID)
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	if p == nil || p.ConnectionCount() != devices {
		got := int64(-1)
		if p != nil {
			got = p.ConnectionCount()
		}
		t.Fatalf("connection_count = %d after %d concurrent connects, want %d — increments were lost", got, devices, devices)
	}
}

// ResetAll is what runs at boot. Counters left over from a process that died
// describe sockets that no longer exist; without this everyone who was online
// at the crash stays green forever.
func TestIntegrationPresenceResetAll(t *testing.T) {
	db := openTestDB(t)

	userID := uniqueTestID(22)
	cleanupPresenceRows(t, db, userID)

	repo := NewPresenceRepository(db)
	ctx := context.Background()

	if _, err := repo.IncrementConnection(ctx, userID, nil, nil, mtime.Now()); err != nil {
		t.Fatalf("connect: %v", err)
	}

	if err := repo.ResetAll(ctx); err != nil {
		t.Fatalf("reset all: %v", err)
	}

	p, err := repo.FindByUserId(ctx, userID)
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	if p == nil {
		t.Fatal("presence row disappeared — ResetAll must clear counters, not delete rows")
	}
	if p.ConnectionCount() != 0 || p.IsOnline() {
		t.Errorf("after reset: count=%d state=%s, want 0/OFFLINE", p.ConnectionCount(), p.PresenceState())
	}
}
