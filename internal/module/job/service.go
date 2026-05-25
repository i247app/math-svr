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
		stderrors.New("job runtime is not initialised"))
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
	default:
		return errs.NewError(ctx, fallback, nil, err)
	}
}
