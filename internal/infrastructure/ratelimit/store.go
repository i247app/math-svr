package ratelimit

import (
	"hash/fnv"
	"sync"
	"time"
)

const shardCount = 64

// store is the per-key state table every algorithm shares: sharded maps so
// unrelated keys do not contend on one mutex. Idle keys are swept inline,
// at most once per idle period per shard, so there is no janitor goroutine
// to start or stop.
type store[S any] struct {
	shards      [shardCount]shard[S]
	maxPerShard int
	idle        time.Duration
}

type shard[S any] struct {
	mu        sync.Mutex
	m         map[string]*entry[S]
	lastSweep time.Time
}

type entry[S any] struct {
	state S
	seen  time.Time
}

func newStore[S any](maxKeys int, idle time.Duration) *store[S] {
	st := &store[S]{maxPerShard: max(1, maxKeys/shardCount), idle: idle}
	for i := range st.shards {
		st.shards[i].m = make(map[string]*entry[S])
	}
	return st
}

// with runs fn on key's state under the shard lock. ok is false when the
// key is new and the shard is full even after a sweep — the caller then
// lets the request through untracked.
func (st *store[S]) with(key string, now time.Time, fn func(s *S) Decision) (d Decision, ok bool) {
	h := fnv.New32a()
	_, _ = h.Write([]byte(key))
	sh := &st.shards[h.Sum32()%shardCount]

	sh.mu.Lock()
	defer sh.mu.Unlock()

	if now.Sub(sh.lastSweep) >= st.idle {
		sh.sweep(now, st.idle)
	}

	e, found := sh.m[key]
	if !found {
		if len(sh.m) >= st.maxPerShard {
			sh.sweep(now, st.idle)
			if len(sh.m) >= st.maxPerShard {
				return Decision{}, false
			}
		}
		e = &entry[S]{}
		sh.m[key] = e
	}
	e.seen = now
	return fn(&e.state), true
}

func (sh *shard[S]) sweep(now time.Time, idle time.Duration) {
	for k, e := range sh.m {
		if now.Sub(e.seen) >= idle {
			delete(sh.m, k)
		}
	}
	sh.lastSweep = now
}
