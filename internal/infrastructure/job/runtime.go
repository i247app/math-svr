package job

import (
	"context"
	"fmt"
	"runtime/debug"
	"sort"
	"sync"
	"time"

	"math-ai.com/math-ai/internal/infrastructure/logger"
)

// Runtime owns the lifecycle of every registered CronJob and the task
// dispatcher. Lifecycle is one-shot: NewRuntime → Start → Stop. Re-start
// after Stop is not supported (the stop channel is single-use).
//
// Concurrency model:
//
//   - One scheduler goroutine per CronJob ("cron loop"). Sleeps until
//     Schedule.Next, fires the job, repeats. Cancelled via stopCh.
//   - A worker pool of cfg.TaskWorkers goroutines pulling from a bounded
//     chan. Each worker calls TaskHandler.Handle synchronously and
//     re-enqueues on retry.
//   - Each in-flight Run/Handle counts against runWg; each scheduler
//     and worker shell counts against workWg. Stop drains workWg first
//     (to stop new work being dispatched), then waits on runWg up to
//     cfg.DrainTimeout.
//
// All execution contexts derive from the runtime stop signal, so a job
// that honours ctx.Done() shuts down promptly when Stop is called.
type Runtime struct {
	cfg      Config
	registry *Registry
	queue    chan taskEnvelope

	mu       sync.Mutex
	started  bool
	stopping bool
	states   map[string]*cronState

	stopCh chan struct{}
	workWg sync.WaitGroup // schedulers + workers + delayed-enqueue timers
	runWg  sync.WaitGroup // in-flight CronJob.Run / TaskHandler.Handle calls
}

// cronState holds per-job mutable state. All fields are guarded by
// Runtime.mu — including the wake channel pointer, which is replaced
// on every pause/resume to wake the cron loop without race.
type cronState struct {
	job CronJob

	paused   bool
	inFlight bool

	nextRunAt  *time.Time
	lastRunAt  *time.Time
	lastStatus *RunStatus
	lastError  *string
	lastDur    *time.Duration

	// wake is closed by PauseJob / ResumeJob / TriggerOnce(?) to
	// interrupt the cron loop's current sleep. The loop reads it once
	// per iteration under Runtime.mu so the replace-after-close pattern
	// is safe.
	wake chan struct{}
}

type taskEnvelope struct {
	name    string
	payload []byte
	attempt int // 1-indexed: 1 = first try
	opts    TaskOptions
}

// NewRuntime builds an unstarted Runtime backed by reg. The registry
// must already be populated; the Runtime does not re-read it after
// Start.
func NewRuntime(cfg Config, reg *Registry) *Runtime {
	cfg = cfg.withDefaults()
	return &Runtime{
		cfg:      cfg,
		registry: reg,
		queue:    make(chan taskEnvelope, cfg.TaskQueueSize),
		states:   make(map[string]*cronState),
		stopCh:   make(chan struct{}),
	}
}

// Started reports whether Start has been called and Stop has not yet
// been initiated. The HTTP control plane uses this to fail fast on
// requests that arrive during a shutdown window.
func (r *Runtime) Started() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.started && !r.stopping
}

// Config returns the resolved (post-defaults) config. Read-only.
func (r *Runtime) Config() Config { return r.cfg }

// Start launches every registered CronJob's scheduler and the task
// worker pool. It is safe to call Start at most once per Runtime;
// subsequent calls are no-ops.
func (r *Runtime) Start(ctx context.Context) {
	r.mu.Lock()
	if r.started {
		r.mu.Unlock()
		return
	}
	r.started = true
	for _, j := range r.registry.Crons() {
		r.states[j.Name()] = &cronState{job: j, wake: make(chan struct{})}
	}
	r.mu.Unlock()

	log := logger.From(ctx)
	log.Infof("job.runtime.start crons=%d tasks=%d workers=%d queue_cap=%d drain=%s",
		len(r.registry.Crons()),
		len(r.registry.Tasks()),
		r.cfg.TaskWorkers,
		r.cfg.TaskQueueSize,
		r.cfg.DrainTimeout,
	)

	for _, j := range r.registry.Crons() {
		r.workWg.Add(1)
		go r.cronLoop(j)
	}
	for i := 0; i < r.cfg.TaskWorkers; i++ {
		r.workWg.Add(1)
		go r.taskWorker(i)
	}
}

// Stop signals every scheduler and worker to exit, cancels every
// in-flight execution context, and waits for the active goroutines to
// drain. If the drain exceeds cfg.DrainTimeout we log a warning and
// keep waiting — Go has no safe primitive for force-killing a
// goroutine, so jobs that ignore ctx.Done() can stall shutdown
// indefinitely. Honour ctx.Done().
//
// Stop is idempotent; the second call returns immediately.
func (r *Runtime) Stop(ctx context.Context) {
	r.mu.Lock()
	if !r.started || r.stopping {
		r.mu.Unlock()
		return
	}
	r.stopping = true
	r.mu.Unlock()

	log := logger.From(ctx)
	log.Infof("job.runtime.stopping drain_timeout=%s", r.cfg.DrainTimeout)

	close(r.stopCh)

	// Phase 1: scheduler shells + worker shells. They exit promptly on
	// stopCh; the wait here is sub-millisecond in the common case
	// (workers between iterations) or up to one in-flight call's
	// duration when a worker was mid-call.
	r.workWg.Wait()
	log.Info("job.runtime.shells_drained")

	// Phase 2: in-flight executions. They received ctx cancellation the
	// moment stopCh closed (see executionContext). Well-behaved jobs
	// exit quickly; misbehaving ones extend the drain.
	done := make(chan struct{})
	go func() {
		r.runWg.Wait()
		close(done)
	}()
	select {
	case <-done:
		log.Info("job.runtime.executions_drained")
	case <-time.After(r.cfg.DrainTimeout):
		log.Warnf("job.runtime.drain_timeout_exceeded budget=%s — some executions ignored ctx cancellation", r.cfg.DrainTimeout)
		<-done
		log.Info("job.runtime.executions_drained_late")
	}
	log.Info("job.runtime.stopped")
}

// ============================================================
// CronJob control plane
// ============================================================

// PauseJob marks a CronJob as paused; the next time its scheduler loop
// observes the wake signal it suspends until ResumeJob. An in-flight
// execution is NOT cancelled — pause only stops scheduling.
func (r *Runtime) PauseJob(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.started || r.stopping {
		return ErrRuntimeUnavailable
	}
	st, ok := r.states[name]
	if !ok {
		return ErrJobNotFound
	}
	if st.paused {
		return ErrJobAlreadyPaused
	}
	st.paused = true
	close(st.wake)
	st.wake = make(chan struct{})
	return nil
}

// ResumeJob clears the paused flag and wakes the scheduler loop.
func (r *Runtime) ResumeJob(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.started || r.stopping {
		return ErrRuntimeUnavailable
	}
	st, ok := r.states[name]
	if !ok {
		return ErrJobNotFound
	}
	if !st.paused {
		return ErrJobNotPaused
	}
	st.paused = false
	close(st.wake)
	st.wake = make(chan struct{})
	return nil
}

// TriggerOnce fires the named CronJob immediately, out of schedule. It
// respects the same skip-if-still-running guard as the scheduler:
// returns ErrJobInFlight if the job is already executing. The fire
// counts against the same runWg as scheduled fires, so a TriggerOnce
// call right before Stop will be drained.
func (r *Runtime) TriggerOnce(ctx context.Context, name string) error {
	r.mu.Lock()
	if !r.started || r.stopping {
		r.mu.Unlock()
		return ErrRuntimeUnavailable
	}
	st, ok := r.states[name]
	if !ok {
		r.mu.Unlock()
		return ErrJobNotFound
	}
	if st.inFlight {
		r.mu.Unlock()
		return ErrJobInFlight
	}
	st.inFlight = true
	j := st.job
	r.mu.Unlock()

	r.runWg.Add(1)
	go r.runCron(j, true)
	return nil
}

// ============================================================
// Task enqueue
// ============================================================

// Enqueue submits a task for execution. opts.Delay schedules the first
// attempt in the future; zero means run as soon as a worker is free.
// Returns ErrTaskHandlerNotFound if the name is not registered and
// ErrQueueFull if the queue is at capacity (non-blocking — Enqueue
// must not stall an HTTP handler).
//
// In Tier 2 this will accept the caller's UnitOfWork so the row insert
// is atomic with the trigger.
func (r *Runtime) Enqueue(ctx context.Context, name string, payload []byte, opts TaskOptions) error {
	r.mu.Lock()
	available := r.started && !r.stopping
	r.mu.Unlock()
	if !available {
		return ErrRuntimeUnavailable
	}
	if _, ok := r.registry.TaskByName(name); !ok {
		return ErrTaskHandlerNotFound
	}

	env := taskEnvelope{
		name:    name,
		payload: payload,
		attempt: 1,
		opts:    opts,
	}

	if opts.Delay > 0 {
		r.scheduleDelayed(env, opts.Delay)
		return nil
	}

	select {
	case r.queue <- env:
		return nil
	default:
		return ErrQueueFull
	}
}

// ============================================================
// Snapshot
// ============================================================

// Snapshot returns the current runtime state. Output is sorted by name
// so consecutive calls produce stable diffs.
func (r *Runtime) Snapshot() Snapshot {
	r.mu.Lock()
	jobs := make([]JobInfo, 0, len(r.states))
	for name, st := range r.states {
		info := JobInfo{
			Name:     name,
			Schedule: st.job.Schedule().String(),
			Status:   JobStatusRunning,
			InFlight: st.inFlight,
		}
		if st.paused {
			info.Status = JobStatusPaused
		}
		if st.nextRunAt != nil {
			t := *st.nextRunAt
			info.NextRunAt = &t
		}
		if st.lastRunAt != nil {
			t := *st.lastRunAt
			info.LastRunAt = &t
		}
		if st.lastStatus != nil {
			s := *st.lastStatus
			info.LastStatus = &s
		}
		if st.lastError != nil {
			s := *st.lastError
			info.LastError = &s
		}
		if st.lastDur != nil {
			s := fmt.Sprintf("%d", st.lastDur.Milliseconds())
			info.LastDuration = &s
		}
		jobs = append(jobs, info)
	}
	started := r.started && !r.stopping
	r.mu.Unlock()

	tasks := make([]TaskInfo, 0)
	for _, h := range r.registry.Tasks() {
		tasks = append(tasks, TaskInfo{Name: h.Name()})
	}

	sort.Slice(jobs, func(i, j int) bool { return jobs[i].Name < jobs[j].Name })
	sort.Slice(tasks, func(i, j int) bool { return tasks[i].Name < tasks[j].Name })

	return Snapshot{
		Started: started,
		Jobs:    jobs,
		Tasks:   tasks,
		Queue: QueueInfo{
			Capacity: cap(r.queue),
			Depth:    len(r.queue),
			Workers:  r.cfg.TaskWorkers,
		},
	}
}

// ============================================================
// Cron scheduler loop
// ============================================================

func (r *Runtime) cronLoop(j CronJob) {
	defer r.workWg.Done()

	for {
		r.mu.Lock()
		st := r.states[j.Name()]
		paused := st.paused
		wake := st.wake
		r.mu.Unlock()

		if paused {
			select {
			case <-r.stopCh:
				return
			case <-wake:
				continue
			}
		}

		now := time.Now()
		next := j.Schedule().Next(now)
		if next.IsZero() {
			logger.From(context.Background()).Warnf("job.cron.invalid_schedule name=%s — loop exiting", j.Name())
			return
		}
		r.mu.Lock()
		if st := r.states[j.Name()]; st != nil {
			tmp := next
			st.nextRunAt = &tmp
		}
		r.mu.Unlock()

		delay := time.Until(next)
		if delay < 0 {
			delay = 0
		}
		timer := time.NewTimer(delay)
		select {
		case <-r.stopCh:
			timer.Stop()
			return
		case <-wake:
			timer.Stop()
			continue
		case <-timer.C:
		}

		// Skip-if-still-running: a slow execution should not be
		// double-fired by the next tick. Pause-during-sleep is also
		// handled here (the wake select above only catches sleep, not
		// the moment of fire).
		r.mu.Lock()
		st = r.states[j.Name()]
		if st == nil || st.paused {
			r.mu.Unlock()
			continue
		}
		if st.inFlight {
			r.mu.Unlock()
			logger.From(context.Background()).Warnf("job.cron.skip_overlap name=%s", j.Name())
			continue
		}
		st.inFlight = true
		r.mu.Unlock()

		r.runWg.Add(1)
		go r.runCron(j, false)
	}
}

func (r *Runtime) runCron(j CronJob, triggered bool) {
	defer r.runWg.Done()

	ctx, cancel := r.executionContext(j.Timeout())
	defer cancel()
	log := logger.From(ctx)

	start := time.Now()
	var (
		execErr  error
		panicked bool
	)
	func() {
		defer func() {
			if rec := recover(); rec != nil {
				panicked = true
				execErr = fmt.Errorf("panic: %v\n%s", rec, debug.Stack())
			}
		}()
		execErr = j.Run(ctx)
	}()

	dur := time.Since(start)
	final := RunStatusSucceeded
	var errStrPtr *string
	if panicked {
		final = RunStatusPanicked
	} else if execErr != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			final = RunStatusTimedOut
		} else {
			final = RunStatusFailed
		}
		s := execErr.Error()
		errStrPtr = &s
	}

	r.mu.Lock()
	if st := r.states[j.Name()]; st != nil {
		st.inFlight = false
		now := time.Now()
		st.lastRunAt = &now
		st.lastStatus = &final
		st.lastError = errStrPtr
		d := dur
		st.lastDur = &d
	}
	r.mu.Unlock()

	switch final {
	case RunStatusSucceeded:
		log.Infof("job.cron.completed name=%s triggered=%t dur_ms=%d", j.Name(), triggered, dur.Milliseconds())
	case RunStatusTimedOut:
		log.Warnf("job.cron.timed_out name=%s dur=%v err=%v", j.Name(), dur, execErr)
	case RunStatusPanicked:
		log.Errorf("job.cron.panicked name=%s dur=%v err=%v", j.Name(), dur, execErr)
	default:
		log.Warnf("job.cron.failed name=%s dur=%v err=%v", j.Name(), dur, execErr)
	}
}

// ============================================================
// Task worker pool
// ============================================================

func (r *Runtime) taskWorker(id int) {
	defer r.workWg.Done()
	for {
		select {
		case <-r.stopCh:
			return
		case env := <-r.queue:
			r.runWg.Add(1)
			r.runTask(id, env)
			r.runWg.Done()
		}
	}
}

func (r *Runtime) runTask(workerID int, env taskEnvelope) {
	handler, ok := r.registry.TaskByName(env.name)
	if !ok {
		// Should not happen: handler is checked at Enqueue. Defensive.
		logger.From(context.Background()).Errorf("job.task.handler_missing name=%s — dropping", env.name)
		return
	}

	timeout := env.opts.Timeout
	if timeout == 0 {
		timeout = r.cfg.DefaultTaskTimeout
	}
	ctx, cancel := r.executionContext(timeout)
	defer cancel()
	log := logger.From(ctx)

	start := time.Now()
	var (
		execErr  error
		panicked bool
	)
	func() {
		defer func() {
			if rec := recover(); rec != nil {
				panicked = true
				execErr = fmt.Errorf("panic: %v\n%s", rec, debug.Stack())
			}
		}()
		execErr = handler.Handle(ctx, env.payload)
	}()
	dur := time.Since(start)

	if execErr == nil && !panicked {
		log.Infof("job.task.completed name=%s attempt=%d worker=%d dur_ms=%d",
			env.name, env.attempt, workerID, dur.Milliseconds())
		return
	}

	policy := env.opts.Retry
	if policy == (RetryPolicy{}) {
		policy = r.cfg.DefaultRetryPolicy
	}
	if env.attempt >= policy.Attempts {
		log.Errorf("job.task.failed_terminal name=%s attempt=%d dur=%v err=%v",
			env.name, env.attempt, dur, execErr)
		return
	}

	delay := policy.DelayFor(env.attempt)
	next := env
	next.attempt++
	next.opts.Retry = policy
	log.Warnf("job.task.retry_scheduled name=%s attempt=%d->%d delay=%s err=%v",
		env.name, env.attempt, next.attempt, delay, execErr)

	r.scheduleDelayed(next, delay)
}

// scheduleDelayed sleeps for delay then enqueues env. Tracked under
// workWg (it's a runtime-managed timer goroutine, not an execution).
// If the queue is full at re-enqueue time, the retry is dropped and
// logged — without persistence we cannot do better in Tier 1.
func (r *Runtime) scheduleDelayed(env taskEnvelope, delay time.Duration) {
	r.workWg.Add(1)
	go func() {
		defer r.workWg.Done()
		if delay > 0 {
			t := time.NewTimer(delay)
			defer t.Stop()
			select {
			case <-r.stopCh:
				return
			case <-t.C:
			}
		}
		select {
		case <-r.stopCh:
			return
		case r.queue <- env:
			return
		default:
			logger.From(context.Background()).Errorf(
				"job.task.retry_dropped name=%s attempt=%d reason=queue_full",
				env.name, env.attempt)
		}
	}()
}

// ============================================================
// Execution context helper
// ============================================================

// executionContext returns a context that is cancelled when EITHER the
// per-attempt timeout fires OR the runtime stops. The bridge goroutine
// exits cleanly under both paths so it does not leak.
func (r *Runtime) executionContext(timeout time.Duration) (context.Context, context.CancelFunc) {
	var (
		ctx    context.Context
		cancel context.CancelFunc
	)
	if timeout > 0 {
		ctx, cancel = context.WithTimeout(context.Background(), timeout)
	} else {
		ctx, cancel = context.WithCancel(context.Background())
	}
	go func() {
		select {
		case <-r.stopCh:
			cancel()
		case <-ctx.Done():
		}
	}()
	return ctx, cancel
}
