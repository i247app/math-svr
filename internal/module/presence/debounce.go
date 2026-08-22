package presence

import (
	"sync"
	"time"
)

// defaultOfflineDelay is how long a disconnect must persist before classmates
// are told. Mobile networks drop constantly — a lift, a tunnel, the app going
// to background — and each of those would otherwise flip the dot for everyone
// who shares a classroom. Ten seconds is long enough to swallow a reconnect
// and short enough that a real departure still looks immediate.
const defaultOfflineDelay = 10 * time.Second

// offlineDebouncer delays offline broadcasts and cancels them when the user
// comes back first.
//
// It debounces only the BROADCAST. The database is written immediately on
// every connect and disconnect, so a fresh read is always the truth; this only
// controls what other people are told, and when.
//
// State is per-process, matching the Hub it shadows. With more than one
// instance each would debounce its own connections, which is correct: an
// instance only knows about the sockets it holds.
type offlineDebouncer struct {
	mu     sync.Mutex
	timers map[int64]*time.Timer
	delay  time.Duration
}

func newOfflineDebouncer(delay time.Duration) *offlineDebouncer {
	if delay <= 0 {
		delay = defaultOfflineDelay
	}
	return &offlineDebouncer{timers: make(map[int64]*time.Timer), delay: delay}
}

// schedule arranges for fn to run after the delay unless cancelled first.
// A second call for the same user replaces the pending timer.
func (d *offlineDebouncer) schedule(userId int64, fn func()) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if existing, ok := d.timers[userId]; ok {
		existing.Stop()
	}
	d.timers[userId] = time.AfterFunc(d.delay, func() {
		d.mu.Lock()
		delete(d.timers, userId)
		d.mu.Unlock()
		fn()
	})
}

// cancel stops a pending offline broadcast and reports whether one was
// waiting.
//
// The return value is what keeps the two sides consistent: a pending timer
// means nobody was ever told this user went offline, so there is nothing to
// correct and the matching online broadcast must be suppressed. Without it,
// every brief network blip would emit a spurious "came online" to the class.
func (d *offlineDebouncer) cancel(userId int64) (wasPending bool) {
	d.mu.Lock()
	defer d.mu.Unlock()

	timer, ok := d.timers[userId]
	if !ok {
		return false
	}
	timer.Stop()
	delete(d.timers, userId)
	return true
}

// stopAll cancels every pending timer. Called on shutdown so no callback fires
// into a half-torn-down process.
func (d *offlineDebouncer) stopAll() {
	d.mu.Lock()
	defer d.mu.Unlock()

	for id, timer := range d.timers {
		timer.Stop()
		delete(d.timers, id)
	}
}
