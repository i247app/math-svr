package ratelimit

import (
	"context"
	"math"
	"time"
)

type slidingWindowCounterState struct {
	start time.Time // start of the current aligned window
	curr  int       // requests in the current window
	prev  int       // requests in the window just before it
}

type slidingWindowCounter struct {
	cfg Config
	st  *store[slidingWindowCounterState]
}

// Allow estimates the count over the last Window as
// prev × (part of the previous window still inside it) + curr.
func (l *slidingWindowCounter) Allow(_ context.Context, key string) (Decision, error) {
	now := l.cfg.Now()
	w := l.cfg.Window
	limit := float64(l.cfg.Limit)

	d, ok := l.st.with(key, now, func(s *slidingWindowCounterState) Decision {
		start := now.Truncate(w)
		switch {
		case s.start.Equal(start):
		case s.start.Add(w).Equal(start):
			s.prev, s.curr, s.start = s.curr, 0, start
		default: // first request, or idle for more than one window
			s.prev, s.curr, s.start = 0, 0, start
		}

		elapsed := now.Sub(start)
		weight := 1 - float64(elapsed)/float64(w)
		est := float64(s.prev)*weight + float64(s.curr)

		if est+1 <= limit {
			s.curr++
			return Decision{Allowed: true, Limit: l.cfg.Limit, Remaining: int(math.Floor(limit - est - 1))}
		}

		// Earliest moment est+1 <= limit: either the previous window's share
		// shrinks enough within this window, or wait for the next one.
		retry := start.Add(w).Sub(now)
		if room := limit - 1 - float64(s.curr); room >= 0 && s.prev > 0 {
			at := time.Duration(float64(w) * (1 - room/float64(s.prev)))
			retry = at - elapsed
		}
		return Decision{Limit: l.cfg.Limit, RetryAfter: max(retry, time.Millisecond)}
	})
	if !ok {
		return untracked(l.cfg), nil
	}
	return d, nil
}
