package ratelimit

import (
	"context"

	"golang.org/x/time/rate"
)

type tokenBucketState struct {
	lim *rate.Limiter
}

// tokenBucket delegates the bucket arithmetic to golang.org/x/time/rate; only
// the per-key table and the Decision shape are ours.
type tokenBucket struct {
	cfg Config
	st  *store[tokenBucketState]
}

func (l *tokenBucket) Allow(_ context.Context, key string) (Decision, error) {
	now := l.cfg.Now()
	d, ok := l.st.with(key, now, func(s *tokenBucketState) Decision {
		if s.lim == nil {
			s.lim = rate.NewLimiter(rate.Every(interval(l.cfg)), l.cfg.Burst)
		}
		r := s.lim.ReserveN(now, 1)
		if delay := r.DelayFrom(now); delay > 0 {
			r.CancelAt(now)
			return Decision{Limit: l.cfg.Burst, RetryAfter: delay}
		}
		return Decision{Allowed: true, Limit: l.cfg.Burst, Remaining: int(s.lim.TokensAt(now))}
	})
	if !ok {
		return untracked(l.cfg), nil
	}
	return d, nil
}
