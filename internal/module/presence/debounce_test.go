package presence

import (
	"sync"
	"testing"
	"time"
)

// A disconnect followed quickly by a reconnect — a tunnel, a backgrounded app
// — must never reach classmates. This is the whole reason the debouncer
// exists.
func TestDebounceCancelledBeforeFiring(t *testing.T) {
	d := newOfflineDebouncer(50 * time.Millisecond)

	var mu sync.Mutex
	fired := false
	d.schedule(7, func() {
		mu.Lock()
		fired = true
		mu.Unlock()
	})

	if wasPending := d.cancel(7); !wasPending {
		t.Fatal("cancel should report a pending timer")
	}

	time.Sleep(120 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if fired {
		t.Error("cancelled timer must not fire")
	}
}

// A real departure still has to be announced, just later.
func TestDebounceFiresAfterDelay(t *testing.T) {
	d := newOfflineDebouncer(30 * time.Millisecond)

	done := make(chan struct{})
	d.schedule(7, func() { close(done) })

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("timer did not fire within the timeout")
	}

	// Once fired, the entry is gone: a later cancel has nothing to report, so
	// a subsequent reconnect DOES announce the user as online again.
	if wasPending := d.cancel(7); wasPending {
		t.Error("a fired timer should no longer be pending")
	}
}

// cancel on a user with nothing scheduled means classmates already believe the
// user is offline, so the caller must go ahead and announce them online.
func TestDebounceCancelWithoutPendingTimer(t *testing.T) {
	d := newOfflineDebouncer(time.Second)
	if wasPending := d.cancel(999); wasPending {
		t.Error("cancel with no pending timer should report false")
	}
}

// Rescheduling replaces the pending timer rather than stacking a second one,
// so a flapping connection cannot queue up a burst of offline broadcasts.
func TestDebounceRescheduleReplacesTimer(t *testing.T) {
	d := newOfflineDebouncer(40 * time.Millisecond)

	var mu sync.Mutex
	count := 0
	bump := func() {
		mu.Lock()
		count++
		mu.Unlock()
	}

	d.schedule(7, bump)
	d.schedule(7, bump)
	d.schedule(7, bump)

	time.Sleep(150 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if count != 1 {
		t.Errorf("callback fired %d times, want exactly 1", count)
	}
}

func TestDebounceStopAll(t *testing.T) {
	d := newOfflineDebouncer(40 * time.Millisecond)

	var mu sync.Mutex
	fired := 0
	for _, uid := range []int64{1, 2, 3} {
		d.schedule(uid, func() {
			mu.Lock()
			fired++
			mu.Unlock()
		})
	}

	d.stopAll()
	time.Sleep(120 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if fired != 0 {
		t.Errorf("%d timers fired after stopAll, want 0", fired)
	}
}
