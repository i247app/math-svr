package ratelimit

import (
	"context"
	"time"
)

type leakyBucketState struct {
	next time.Time // when the queue is empty again (the last queued request's slot)
}

// leakyBucket is the queue form of the leaky bucket: requests leave at one
// per interval. A request that fits in the queue (at most Burst waiting) is
// accepted with a Delay until its slot; a full queue rejects.
type leakyBucket struct {
	cfg Config
	st  *store[leakyBucketState]
}

func (l *leakyBucket) Allow(_ context.Context, key string) (Decision, error) {
	now := l.cfg.Now()
	step := interval(l.cfg)
	horizon := step * time.Duration(l.cfg.Burst)

	d, ok := l.st.with(key, now, func(s *leakyBucketState) Decision {
		slot := now
		if s.next.After(now) {
			slot = s.next
		}
		wait := slot.Sub(now)
		if wait+step > horizon {
			return Decision{Limit: l.cfg.Burst, RetryAfter: wait + step - horizon}
		}
		s.next = slot.Add(step)
		queued := int((s.next.Sub(now) + step - 1) / step)
		return Decision{Allowed: true, Limit: l.cfg.Burst, Remaining: l.cfg.Burst - queued, Delay: wait}
	})
	if !ok {
		return untracked(l.cfg), nil
	}
	return d, nil
}
