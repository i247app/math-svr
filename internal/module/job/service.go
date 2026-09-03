package job

import (
	"context"
	"encoding/json"
	stderrors "errors"
	"fmt"
	"time"

	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	jobruntime "math-ai.com/math-ai/internal/infrastructure/job"
	"math-ai.com/math-ai/internal/infrastructure/logger"
)

// Service wraps *jobruntime.Runtime with the project's error contract:
// the runtime returns plain sentinel errors; the service translates
// them into MathError values for the HTTP layer to render.
//
// All methods are safe to call before Runtime.Start (they return
// JOB_RUNTIME_UNAVAILABLE) and during shutdown (same).
type Service struct {
	runtime *jobruntime.Runtime
}

func NewService(runtime *jobruntime.Runtime) *Service {
	return &Service{runtime: runtime}
}

// Snapshot returns the runtime view. Caller should treat the returned
// pointer as read-only.
func (s *Service) Snapshot(ctx context.Context) (*SnapshotResponse, error) {
	if s.runtime == nil {
		return nil, runtimeUnavailable(ctx)
	}
	snap := s.runtime.Snapshot()
	return &snap, nil
}

func (s *Service) TriggerJob(ctx context.Context, req *JobNameRequest) (*ActionResponse, error) {
	if err := validateJobName(ctx, req.Name); err != nil {
		return nil, err
	}
	if s.runtime == nil {
		return nil, runtimeUnavailable(ctx)
	}
	if err := s.runtime.TriggerOnce(ctx, req.Name); err != nil {
		return nil, mapRuntimeError(ctx, err, status.JOB_TRIGGER_FAILED)
	}
	return &ActionResponse{Name: req.Name, Result: "triggered"}, nil
}

func (s *Service) PauseJob(ctx context.Context, req *JobNameRequest) (*ActionResponse, error) {
	if err := validateJobName(ctx, req.Name); err != nil {
		return nil, err
	}
	if s.runtime == nil {
		return nil, runtimeUnavailable(ctx)
	}
	if err := s.runtime.PauseJob(req.Name); err != nil {
		return nil, mapRuntimeError(ctx, err, status.JOB_TRIGGER_FAILED)
	}
	return &ActionResponse{Name: req.Name, Result: "paused"}, nil
}

func (s *Service) ResumeJob(ctx context.Context, req *JobNameRequest) (*ActionResponse, error) {
	if err := validateJobName(ctx, req.Name); err != nil {
		return nil, err
	}
	if s.runtime == nil {
		return nil, runtimeUnavailable(ctx)
	}
	if err := s.runtime.ResumeJob(req.Name); err != nil {
		return nil, mapRuntimeError(ctx, err, status.JOB_TRIGGER_FAILED)
	}
	return &ActionResponse{Name: req.Name, Result: "resumed"}, nil
}

// UpdateSchedule retargets a CronJob's cadence at runtime. The new
// schedule takes effect immediately — the runtime wakes the scheduler
// loop rather than letting the pending sleep run out.
//
// The override lives in memory only: a restart reverts the job to the
// schedule declared in internal/jobs/. Treat this as an operations
// lever (slow a noisy job down, speed one up to observe it), not as
// configuration.
func (s *Service) UpdateSchedule(ctx context.Context, req *UpdateScheduleRequest) (*ScheduleResponse, error) {
	if err := validateJobName(ctx, req.Name); err != nil {
		return nil, err
	}
	sched, err := parseSchedule(ctx, req.Schedule)
	if err != nil {
		return nil, err
	}
	if s.runtime == nil {
		return nil, runtimeUnavailable(ctx)
	}
	if err := s.runtime.UpdateSchedule(req.Name, sched); err != nil {
		return nil, mapRuntimeError(ctx, err, status.JOB_UPDATE_SCHEDULE_FAILED)
	}
	return s.scheduleResult(ctx, req.Name, "schedule_updated")
}

// ResetSchedule drops the runtime override so the job returns to the
// schedule declared in code. Idempotent.
func (s *Service) ResetSchedule(ctx context.Context, req *JobNameRequest) (*ScheduleResponse, error) {
	if err := validateJobName(ctx, req.Name); err != nil {
		return nil, err
	}
	if s.runtime == nil {
		return nil, runtimeUnavailable(ctx)
	}
	if err := s.runtime.ResetSchedule(req.Name); err != nil {
		return nil, mapRuntimeError(ctx, err, status.JOB_UPDATE_SCHEDULE_FAILED)
	}
	return s.scheduleResult(ctx, req.Name, "schedule_reset")
}

// scheduleResult reads the job's post-mutation state back out of the
// runtime so the caller sees the resolved schedule and the recomputed
// next_run_at rather than an echo of its own request.
func (s *Service) scheduleResult(ctx context.Context, name, result string) (*ScheduleResponse, error) {
	info, ok := s.runtime.JobByName(name)
	if !ok {
		// The job existed a moment ago inside the runtime's lock; states
		// are never deleted, so this is unreachable in practice.
		return nil, errs.NewError(ctx, status.JOB_NOT_FOUND, nil, jobruntime.ErrJobNotFound)
	}
	logger.From(ctx).Infof("job.schedule.%s name=%s schedule=%q next_run_at=%v",
		result, info.Name, info.Schedule, info.NextRunAt)
	return &ScheduleResponse{Result: result, Job: info}, nil
}

func (s *Service) EnqueueTask(ctx context.Context, req *EnqueueTaskRequest) (*ActionResponse, error) {
	if err := validateTaskName(ctx, req.Name); err != nil {
		return nil, err
	}
	if s.runtime == nil {
		return nil, runtimeUnavailable(ctx)
	}

	var payload []byte
	if req.Payload != nil {
		b, err := json.Marshal(req.Payload)
		if err != nil {
			return nil, errs.NewError(ctx, status.TASK_ENQUEUE_FAILED, nil,
				fmt.Errorf("payload marshal: %w", err))
		}
		payload = b
	}

	opts := jobruntime.TaskOptions{}
	if req.DelaySeconds > 0 {
		opts.Delay = time.Duration(req.DelaySeconds) * time.Second
	}
	if req.TimeoutSeconds > 0 {
		opts.Timeout = time.Duration(req.TimeoutSeconds) * time.Second
	}
	if req.Attempts > 0 {
		opts.Retry = jobruntime.RetryPolicy{
			Attempts: req.Attempts,
			Backoff:  jobruntime.BackoffExponential,
			Base:     5 * time.Second,
			MaxDelay: 5 * time.Minute,
		}
	}

	if err := s.runtime.Enqueue(ctx, req.Name, payload, opts); err != nil {
		return nil, mapRuntimeError(ctx, err, status.TASK_ENQUEUE_FAILED)
	}
	return &ActionResponse{Name: req.Name, Result: "enqueued"}, nil
}

func runtimeUnavailable(ctx context.Context) error {
	return errs.NewError(ctx, status.JOB_RUNTIME_UNAVAILABLE, nil,
		ErrJobRuntimeNotInitialised)
}

// mapRuntimeError translates the runtime's sentinel errors into the
// matching JOB_*/TASK_* status code. fallback covers errors we did not
// explicitly enumerate — the runtime should not produce them, but we
// degrade gracefully to a generic failure status rather than 500.
func mapRuntimeError(ctx context.Context, err error, fallback status.StatusCode) error {
	switch {
	case stderrors.Is(err, jobruntime.ErrJobNotFound):
		return errs.NewError(ctx, status.JOB_NOT_FOUND, nil, err)
	case stderrors.Is(err, jobruntime.ErrJobAlreadyPaused):
		return errs.NewError(ctx, status.JOB_ALREADY_PAUSED, nil, err)
	case stderrors.Is(err, jobruntime.ErrJobNotPaused):
		return errs.NewError(ctx, status.JOB_NOT_PAUSED, nil, err)
	case stderrors.Is(err, jobruntime.ErrJobInFlight):
		return errs.NewError(ctx, status.JOB_IN_FLIGHT, nil, err)
	case stderrors.Is(err, jobruntime.ErrRuntimeUnavailable):
		return errs.NewError(ctx, status.JOB_RUNTIME_UNAVAILABLE, nil, err)
	case stderrors.Is(err, jobruntime.ErrTaskHandlerNotFound):
		return errs.NewError(ctx, status.TASK_HANDLER_NOT_FOUND, nil, err)
	case stderrors.Is(err, jobruntime.ErrQueueFull):
		return errs.NewError(ctx, status.TASK_QUEUE_FULL, nil, err)
	case stderrors.Is(err, jobruntime.ErrInvalidSchedule):
		// parseSchedule rejects these first; this covers the runtime's own
		// re-validation so the caller still gets the precise status code.
		return errs.NewError(ctx, status.JOB_INVALID_SCHEDULE, nil, err)
	default:
		return errs.NewError(ctx, fallback, nil, err)
	}
}
