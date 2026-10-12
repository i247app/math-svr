package ratelimit

import (
	"context"
	"time"
)

type fixedWindowState struct {
	start time.Time // start of the window `count` belongs to
	count int
}

type fixedWindow struct {
	cfg Config
	st  *store[fixedWindowState]
}

func (l *fixedWindow) Allow(_ context.Context, key string) (Decision, error) {
	now := l.cfg.Now()
	d, ok := l.st.with(key, now, func(s *fixedWindowState) Decision {
		start := now.Truncate(l.cfg.Window)
		if !s.start.Equal(start) {
			s.start, s.count = start, 0
		}
		if s.count >= l.cfg.Limit {
			return Decision{Limit: l.cfg.Limit, RetryAfter: start.Add(l.cfg.Window).Sub(now)}
		}
		s.count++
		return Decision{Allowed: true, Limit: l.cfg.Limit, Remaining: l.cfg.Limit - s.count}
	})
	if !ok {
		return untracked(l.cfg), nil
	}
	return d, nil
}
