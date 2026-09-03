package job

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

// fakeCron is a CronJob whose declared schedule is far enough in the
// future that it never fires on its own — so any observed run is proof
// the schedule override took effect, not a coincidence of timing.
type fakeCron struct {
	name  string
	fired chan struct{}

	mu       sync.Mutex
	declared Schedule
}

func newFakeCron(name string, declared Schedule) *fakeCron {
	return &fakeCron{name: name, declared: declared, fired: make(chan struct{}, 64)}
}

func (f *fakeCron) Name() string { return f.name }

func (f *fakeCron) Schedule() Schedule {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.declared
}

func (f *fakeCron) Timeout() time.Duration { return time.Second }

func (f *fakeCron) Run(context.Context) error {
	select {
	case f.fired <- struct{}{}:
	default: // never block the runtime on a full test buffer
	}
	return nil
}

// startRuntime boots a Runtime with j registered and guarantees it is
// stopped when the test ends.
func startRuntime(t *testing.T, j CronJob) *Runtime {
	t.Helper()
	reg := NewRegistry()
	reg.RegisterCron(j)
	rt := NewRuntime(Config{TaskWorkers: 1}, reg)
	rt.Start(context.Background())
	t.Cleanup(func() { rt.Stop(context.Background()) })
	return rt
}

// awaitFire waits for the job to run, failing the test on timeout.
func awaitFire(t *testing.T, f *fakeCron, within time.Duration, what string) {
	t.Helper()
	select {
	case <-f.fired:
	case <-time.After(within):
		t.Fatalf("job did not fire within %s: %s", within, what)
	}
}

// TestUpdateScheduleWakesSleepingLoop is the core behavioural claim of
// the /jobs/schedule/update endpoint: changing the schedule takes effect
// immediately, not after the pending sleep expires.
//
// The job is declared "daily at 03:00 UTC", so its scheduler is parked
// on a timer up to 24h long. If UpdateSchedule only mutated state
// without closing the wake channel, this test would hang until timeout.
func TestUpdateScheduleWakesSleepingLoop(t *testing.T) {
	j := newFakeCron("test.wake", DailyAt(3, 0, time.UTC))
	rt := startRuntime(t, j)

	// Nothing should fire on the declared daily schedule.
	select {
	case <-j.fired:
		t.Fatal("job fired on its declared daily schedule; test setup is wrong")
	case <-time.After(100 * time.Millisecond):
	}

	if err := rt.UpdateSchedule("test.wake", EveryDuration(20*time.Millisecond)); err != nil {
		t.Fatalf("UpdateSchedule() = %v, want nil", err)
	}
	awaitFire(t, j, 2*time.Second, "after switching to every 20ms")

	info, ok := rt.JobByName("test.wake")
	if !ok {
		t.Fatal("JobByName() not found after update")
	}
	if !info.ScheduleOverridden {
		t.Error("ScheduleOverridden = false, want true")
	}
	if want := "every 20ms"; info.Schedule != want {
		t.Errorf("Schedule = %q, want %q", info.Schedule, want)
	}
	if want := "daily 03:00 UTC"; info.DefaultSchedule != want {
		t.Errorf("DefaultSchedule = %q, want %q", info.DefaultSchedule, want)
	}
}

// TestResetScheduleRestoresDeclaredSchedule proves the override is
// reversible without a restart — the operator's escape hatch after
// speeding a job up to observe it.
func TestResetScheduleRestoresDeclaredSchedule(t *testing.T) {
	j := newFakeCron("test.reset", DailyAt(3, 0, time.UTC))
	rt := startRuntime(t, j)

	if err := rt.UpdateSchedule("test.reset", EveryDuration(20*time.Millisecond)); err != nil {
		t.Fatalf("UpdateSchedule() = %v, want nil", err)
	}
	awaitFire(t, j, 2*time.Second, "override active")

	if err := rt.ResetSchedule("test.reset"); err != nil {
		t.Fatalf("ResetSchedule() = %v, want nil", err)
	}

	info, ok := rt.JobByName("test.reset")
	if !ok {
		t.Fatal("JobByName() not found after reset")
	}
	if info.ScheduleOverridden {
		t.Error("ScheduleOverridden = true after reset, want false")
	}
	if want := "daily 03:00 UTC"; info.Schedule != want {
		t.Errorf("Schedule = %q after reset, want the declared %q", info.Schedule, want)
	}

	// Drain runs already in flight when reset landed, then confirm the
	// fast cadence has genuinely stopped.
	time.Sleep(100 * time.Millisecond)
	for len(j.fired) > 0 {
		<-j.fired
	}
	select {
	case <-j.fired:
		t.Fatal("job still firing on the overridden cadence after ResetSchedule")
	case <-time.After(300 * time.Millisecond):
	}

	// Reset is idempotent — a second call on a non-overridden job is a
	// no-op success, not an error.
	if err := rt.ResetSchedule("test.reset"); err != nil {
		t.Fatalf("second ResetSchedule() = %v, want nil (idempotent)", err)
	}
}

// TestUpdateScheduleRejectsInvalid guards the invariant that keeps a bad
// control-plane call from parking a scheduler loop: validation happens
// before any state is mutated, so a rejected update leaves the job on
// its previous schedule.
func TestUpdateScheduleRejectsInvalid(t *testing.T) {
	j := newFakeCron("test.invalid", DailyAt(3, 0, time.UTC))
	rt := startRuntime(t, j)

	for _, bad := range []Schedule{
		{},                          // zero value
		EveryDuration(0),            // non-positive interval
		DailyAt(25, 0, time.UTC),    // hour out of range
		WeeklyAt(9, 8, 0, time.UTC), // weekday out of range
	} {
		err := rt.UpdateSchedule("test.invalid", bad)
		if !errors.Is(err, ErrInvalidSchedule) {
			t.Fatalf("UpdateSchedule(%+v) = %v, want ErrInvalidSchedule", bad, err)
		}
	}

	info, _ := rt.JobByName("test.invalid")
	if info.ScheduleOverridden {
		t.Error("a rejected update mutated state; ScheduleOverridden = true, want false")
	}
	if want := "daily 03:00 UTC"; info.Schedule != want {
		t.Errorf("Schedule = %q after rejected updates, want the untouched %q", info.Schedule, want)
	}
}

func TestScheduleControlPlaneUnknownJob(t *testing.T) {
	rt := startRuntime(t, newFakeCron("test.known", DailyAt(3, 0, time.UTC)))

	if err := rt.UpdateSchedule("test.missing", EveryDuration(time.Minute)); !errors.Is(err, ErrJobNotFound) {
		t.Errorf("UpdateSchedule(unknown) = %v, want ErrJobNotFound", err)
	}
	if err := rt.ResetSchedule("test.missing"); !errors.Is(err, ErrJobNotFound) {
		t.Errorf("ResetSchedule(unknown) = %v, want ErrJobNotFound", err)
	}
	if _, ok := rt.JobByName("test.missing"); ok {
		t.Error("JobByName(unknown) = ok, want not found")
	}
}

// TestScheduleControlPlaneAfterStop covers the shutdown window: the
// control plane must fail fast rather than mutate a draining runtime.
func TestScheduleControlPlaneAfterStop(t *testing.T) {
	reg := NewRegistry()
	reg.RegisterCron(newFakeCron("test.stopped", DailyAt(3, 0, time.UTC)))
	rt := NewRuntime(Config{TaskWorkers: 1}, reg)
	rt.Start(context.Background())
	rt.Stop(context.Background())

	if err := rt.UpdateSchedule("test.stopped", EveryDuration(time.Minute)); !errors.Is(err, ErrRuntimeUnavailable) {
		t.Errorf("UpdateSchedule after Stop = %v, want ErrRuntimeUnavailable", err)
	}
	if err := rt.ResetSchedule("test.stopped"); !errors.Is(err, ErrRuntimeUnavailable) {
		t.Errorf("ResetSchedule after Stop = %v, want ErrRuntimeUnavailable", err)
	}
}

// TestCronLoopParksOnInvalidDeclaredSchedule locks in the change from
// "exit the loop" to "park on wake". A loop that exits is unrecoverable
// without a restart, which would make a later UpdateSchedule silently
// ineffective — the endpoint would return success and nothing would run.
func TestCronLoopParksOnInvalidDeclaredSchedule(t *testing.T) {
	// A code-declared malformed schedule: Next returns the zero Time.
	j := newFakeCron("test.parked", Schedule{})
	rt := startRuntime(t, j)

	time.Sleep(100 * time.Millisecond) // let the loop reach its parked state

	if err := rt.UpdateSchedule("test.parked", EveryDuration(20*time.Millisecond)); err != nil {
		t.Fatalf("UpdateSchedule() = %v, want nil", err)
	}
	awaitFire(t, j, 2*time.Second, "loop should have been revived from its parked state")
}
