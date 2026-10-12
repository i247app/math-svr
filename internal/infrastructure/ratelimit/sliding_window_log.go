package ratelimit

import (
	"context"
	"time"
)

type slidingWindowLogState struct {
	times []time.Time // accepted requests inside the window, oldest first
}

type slidingWindowLog struct {
	cfg Config
	st  *store[slidingWindowLogState]
}

func (l *slidingWindowLog) Allow(_ context.Context, key string) (Decision, error) {
	now := l.cfg.Now()
	d, ok := l.st.with(key, now, func(s *slidingWindowLogState) Decision {
		cutoff := now.Add(-l.cfg.Window)
		i := 0
		for i < len(s.times) && !s.times[i].After(cutoff) {
			i++
		}
		s.times = s.times[i:]

		if len(s.times) >= l.cfg.Limit {
			return Decision{Limit: l.cfg.Limit, RetryAfter: s.times[0].Add(l.cfg.Window).Sub(now)}
		}
		s.times = append(s.times, now)
		return Decision{Allowed: true, Limit: l.cfg.Limit, Remaining: l.cfg.Limit - len(s.times)}
	})
	if !ok {
		return untracked(l.cfg), nil
	}
	return d, nil
}
