package presence

import (
	"context"
	"errors"
	"testing"

	"math-ai.com/math-ai/internal/application/transaction"
	domain "math-ai.com/math-ai/internal/domain/presence"
	"math-ai.com/math-ai/internal/domain/shared/mtime"
)

// fakePresenceRepo models the counter the way MySQL does — clamped at zero,
// state derived from the count — so the test exercises the transition logic
// the socket layer and (later) the broadcast path depend on.
type fakePresenceRepo struct {
	count int64
	err   error
}

func (f *fakePresenceRepo) row() *domain.Presence {
	p := domain.NewPresence()
	p.SetConnectionCount(f.count)
	return p
}

func (f *fakePresenceRepo) IncrementConnection(ctx context.Context, userId int64, deviceUuid, platform *string, now mtime.MathTime) (*domain.Presence, error) {
	if f.err != nil {
		return nil, f.err
	}
	f.count++
	return f.row(), nil
}

func (f *fakePresenceRepo) DecrementConnection(ctx context.Context, userId int64, now mtime.MathTime) (*domain.Presence, error) {
	if f.err != nil {
		return nil, f.err
	}
	if f.count > 0 {
		f.count--
	}
	return f.row(), nil
}

func (f *fakePresenceRepo) FindByUserId(ctx context.Context, userId int64) (*domain.Presence, error) {
	return f.row(), nil
}

func (f *fakePresenceRepo) ListByUserIds(ctx context.Context, userIds []int64) (map[int64]*domain.Presence, error) {
	return map[int64]*domain.Presence{}, nil
}

func (f *fakePresenceRepo) ResetAll(ctx context.Context) error { return f.err }

// fakeUow runs fn immediately against the supplied repos — no transaction.
type fakeUow struct{ repo domain.IRepository }

func (u fakeUow) Do(ctx context.Context, fn func(ctx context.Context, repos transaction.Repositories) error) error {
	return fn(ctx, transaction.Repositories{Presence: u.repo})
}

func newTestService(repo domain.IRepository) *Service {
	// nil collaborators = realtime disabled; presence is recorded but not
	// announced, which is exactly the path these tests exercise.
	return NewService(fakeUow{repo: repo}, repo, nil, nil)
}

// The dot must only change on the first connect and the last disconnect.
// Everything in between is a second device, and broadcasting it would tell
// every classmate a user "came online" who was already online.
func TestMarkOnlineOfflineTransitions(t *testing.T) {
	ctx := context.Background()
	repo := &fakePresenceRepo{}
	svc := newTestService(repo)

	steps := []struct {
		name       string
		connect    bool
		wantChange bool
		wantCount  int64
	}{
		{name: "first device connects", connect: true, wantChange: true, wantCount: 1},
		{name: "second device connects", connect: true, wantChange: false, wantCount: 2},
		{name: "second device disconnects", connect: false, wantChange: false, wantCount: 1},
		{name: "last device disconnects", connect: false, wantChange: true, wantCount: 0},
	}

	for _, st := range steps {
		t.Run(st.name, func(t *testing.T) {
			var got bool
			var err error
			if st.connect {
				got, err = svc.MarkOnline(ctx, 7, nil, nil)
			} else {
				got, err = svc.MarkOffline(ctx, 7)
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != st.wantChange {
				t.Errorf("transition = %v, want %v", got, st.wantChange)
			}
			if repo.count != st.wantCount {
				t.Errorf("connection_count = %d, want %d", repo.count, st.wantCount)
			}
		})
	}
}

// A disconnect for a user already at zero must stay at zero and must not
// report a fresh offline transition, or an unclean shutdown would push the
// counter negative and re-broadcast on every stray close.
func TestMarkOfflineClampsAtZero(t *testing.T) {
	ctx := context.Background()
	repo := &fakePresenceRepo{count: 0}
	svc := newTestService(repo)

	if _, err := svc.MarkOffline(ctx, 7); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.count != 0 {
		t.Fatalf("connection_count = %d, want 0", repo.count)
	}
}

func TestMarkOnlineWrapsRepoError(t *testing.T) {
	repo := &fakePresenceRepo{err: errors.New("db down")}
	svc := newTestService(repo)

	if _, err := svc.MarkOnline(context.Background(), 7, nil, nil); err == nil {
		t.Fatal("expected an error when the repository fails")
	}
}
