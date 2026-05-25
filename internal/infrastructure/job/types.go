// Package job is the in-process runtime for math-svr's background work.
//
// Two concepts live here:
//
//   - CronJob — a recurring unit of work driven by a wall-clock Schedule
//     (every-N / daily / weekly). No payload, no per-fire arguments. Used
//     for things like nightly quiz cleanup or every-15-minute session
//     reaping.
//
//   - Task    — a one-shot unit of work with an opaque payload and a
//     RetryPolicy. Used for ad-hoc work enqueued by request handlers or
//     by a cron job that fans out per-row work (e.g. weekly digest mails).
//
// Both are addressed by a stable Name string that survives restarts and
// will become the distributed-claim key in Tier 2 (MySQL-backed runs
// with SELECT … FOR UPDATE SKIP LOCKED). Keep names stable.
//
// The Runtime is single-process and in-memory in Tier 1. Persistence,
// horizontal-safe claim, and an out-of-process worker binary are Tier 2/3.
package job

import (
	"context"
	"fmt"
	"time"
)

// ============================================================
// Schedule
// ============================================================

// ScheduleKind tags which variant of Schedule a value represents.
// Construct schedules through EveryDuration / DailyAt / WeeklyAt rather
// than the zero value — Next returns the zero Time for an unrecognised
// kind, which the scheduler treats as "do not run".
type ScheduleKind int

const (
	ScheduleEvery ScheduleKind = iota + 1
	ScheduleDaily
	ScheduleWeekly
)

// Schedule is the (value-typed) source of truth for when a CronJob
// should fire next. The runtime calls Next(time.Now()) each iteration;
// the result is anchored on wall-clock time, not on a per-process tick
// counter, so the cadence survives restarts.
type Schedule struct {
	Kind ScheduleKind

	// Every is set when Kind == ScheduleEvery. Must be > 0.
	Every time.Duration

	// Hour, Minute are set when Kind is Daily or Weekly.
	Hour, Minute int

	// Weekday is set when Kind == ScheduleWeekly.
	Weekday time.Weekday

	// Location is the timezone for Hour/Minute/Weekday. The constructor
	// helpers always set it; the Next fallback is UTC.
	Location *time.Location
}

// EveryDuration fires every d, starting d after the runtime computes
// its first Next call (i.e. not anchored on a wall-clock boundary).
// Use DailyAt/WeeklyAt when you need an anchored schedule.
func EveryDuration(d time.Duration) Schedule {
	return Schedule{Kind: ScheduleEvery, Every: d}
}

// DailyAt fires once per day at hour:minute in loc. If loc is nil, the
// schedule resolves in UTC.
func DailyAt(hour, minute int, loc *time.Location) Schedule {
	return Schedule{Kind: ScheduleDaily, Hour: hour, Minute: minute, Location: loc}
}

// WeeklyAt fires once per week on weekday at hour:minute in loc.
func WeeklyAt(weekday time.Weekday, hour, minute int, loc *time.Location) Schedule {
	return Schedule{Kind: ScheduleWeekly, Weekday: weekday, Hour: hour, Minute: minute, Location: loc}
}

// Next returns the next instant strictly after now at which the
// schedule fires. Zero Time means "malformed — do not schedule".
func (s Schedule) Next(now time.Time) time.Time {
	loc := s.Location
	if loc == nil {
		loc = time.UTC
	}
	switch s.Kind {
	case ScheduleEvery:
		if s.Every <= 0 {
			return time.Time{}
		}
		return now.Add(s.Every)
	case ScheduleDaily:
		t := time.Date(now.Year(), now.Month(), now.Day(), s.Hour, s.Minute, 0, 0, loc)
		if !t.After(now) {
			t = t.AddDate(0, 0, 1)
		}
		return t
	case ScheduleWeekly:
		t := time.Date(now.Year(), now.Month(), now.Day(), s.Hour, s.Minute, 0, 0, loc)
		diff := (int(s.Weekday) - int(t.Weekday()) + 7) % 7
		t = t.AddDate(0, 0, diff)
		if !t.After(now) {
			t = t.AddDate(0, 0, 7)
		}
		return t
	}
	return time.Time{}
}

// String returns a stable, human-readable form. Used in logs and in the
// JobInfo response — keep it terse and grep-friendly.
func (s Schedule) String() string {
	loc := "UTC"
	if s.Location != nil {
		loc = s.Location.String()
	}
	switch s.Kind {
	case ScheduleEvery:
		return "every " + s.Every.String()
	case ScheduleDaily:
		return fmt.Sprintf("daily %02d:%02d %s", s.Hour, s.Minute, loc)
	case ScheduleWeekly:
		return fmt.Sprintf("weekly %s %02d:%02d %s", s.Weekday, s.Hour, s.Minute, loc)
	}
	return "invalid"
}

// ============================================================
// CronJob: recurring, no payload
// ============================================================

// CronJob is the contract for a recurring job. Name must be stable across
// restarts (it is the lookup key, the control-plane handle, and the
// distributed-claim key in Tier 2). Schedule is read once per loop
// iteration; jobs may return a different value on each call but the
// runtime does not memoise. Timeout is the per-execution wall-clock cap
// — zero disables it (the job is bounded only by Runtime drain).
//
// Run MUST honour ctx.Done(): the runtime cancels ctx when the timeout
// fires or when Stop initiates a drain.
type CronJob interface {
	Name() string
	Schedule() Schedule
	Timeout() time.Duration
	Run(ctx context.Context) error
}

// ============================================================
// Task: one-shot, payload + retry
// ============================================================

// TaskHandler processes one task. payload is opaque bytes — typically a
// JSON-encoded request the caller produced at Enqueue time. The handler
// decodes; the runtime never inspects payload contents.
//
// Returning a non-nil error triggers retry per the envelope's RetryPolicy.
type TaskHandler interface {
	Name() string
	Handle(ctx context.Context, payload []byte) error
}

// BackoffKind chooses how retry delay grows with attempt number.
type BackoffKind int

const (
	BackoffConstant    BackoffKind = iota + 1 // delay = Base
	BackoffExponential                        // delay = Base * 2^(attempt-1), capped at MaxDelay
)

// RetryPolicy controls re-enqueue behaviour. Attempts is the total
// number of tries (so Attempts=1 means "no retry"). The runtime treats
// a zero RetryPolicy as "use Runtime.Config.DefaultRetryPolicy".
type RetryPolicy struct {
	Attempts int
	Backoff  BackoffKind
	Base     time.Duration
	MaxDelay time.Duration
}

// DelayFor returns the wait before the (attempt+1)-th run, given that
// attempt failed. attempt is 1-indexed.
func (p RetryPolicy) DelayFor(attempt int) time.Duration {
	if p.Base <= 0 {
		return 0
	}
	max := p.MaxDelay
	if max <= 0 {
		max = 30 * time.Minute
	}
	switch p.Backoff {
	case BackoffExponential:
		d := p.Base
		for i := 1; i < attempt && d < max; i++ {
			d *= 2
		}
		if d > max {
			d = max
		}
		return d
	default:
		return p.Base
	}
}

// TaskOptions are per-enqueue overrides. Zero fields fall back to the
// runtime defaults.
type TaskOptions struct {
	Delay   time.Duration // initial delay; 0 = run as soon as a worker is free
	Retry   RetryPolicy
	Timeout time.Duration // per-attempt timeout; 0 = use runtime default
}

// ============================================================
// Status DTOs (returned by Runtime.Snapshot / handler responses)
// ============================================================

type JobStatus string

const (
	JobStatusRunning JobStatus = "running"
	JobStatusPaused  JobStatus = "paused"
)

type RunStatus string

const (
	RunStatusSucceeded RunStatus = "succeeded"
	RunStatusFailed    RunStatus = "failed"
	RunStatusPanicked  RunStatus = "panicked"
	RunStatusTimedOut  RunStatus = "timed_out"
)

// JobInfo is the read-only view of a CronJob's runtime state. The shape
// is intentionally close to what `ma_job_runs` will store in Tier 2 so
// the HTTP response does not change when persistence lands.
type JobInfo struct {
	Name         string     `json:"name"`
	Schedule     string     `json:"schedule"`
	Status       JobStatus  `json:"status"`
	InFlight     bool       `json:"in_flight"`
	NextRunAt    *time.Time `json:"next_run_at,omitempty"`
	LastRunAt    *time.Time `json:"last_run_at,omitempty"`
	LastStatus   *RunStatus `json:"last_status,omitempty"`
	LastError    *string    `json:"last_error,omitempty"`
	LastDuration *string    `json:"last_duration_ms,omitempty"`
}

// TaskInfo describes a registered task handler. No per-task statistics
// are tracked in Tier 1 — restarts would lose them anyway.
type TaskInfo struct {
	Name string `json:"name"`
}

// QueueInfo summarises the in-memory task queue.
type QueueInfo struct {
	Capacity int `json:"capacity"`
	Depth    int `json:"depth"`
	Workers  int `json:"workers"`
}

// Snapshot is the consolidated read-only view of the runtime.
type Snapshot struct {
	Started bool       `json:"started"`
	Jobs    []JobInfo  `json:"jobs"`
	Tasks   []TaskInfo `json:"tasks"`
	Queue   QueueInfo  `json:"queue"`
}
