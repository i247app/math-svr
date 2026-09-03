package job

import (
	"context"
	"fmt"
	"strings"
	"time"

	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	jobruntime "math-ai.com/math-ai/internal/infrastructure/job"
)

// Bounds on kind=every, enforced here (the API trust boundary) rather
// than in the runtime, so code-declared jobs keep full freedom while an
// operator cannot set a cadence that harms the process.
//
// The floor exists because the runtime skips a tick whose previous
// execution is still in flight (job.cron.skip_overlap): a sub-minute
// interval on a job that takes seconds degrades into a hot loop of
// skipped ticks that only shows up as log noise. The ceiling keeps
// EverySeconds * time.Second from overflowing time.Duration's int64
// nanoseconds into a negative value.
const (
	minEveryInterval = time.Minute
	maxEveryInterval = 30 * 24 * time.Hour
)

func validateJobName(ctx context.Context, name string) error {
	if strings.TrimSpace(name) == "" {
		return errs.NewError(ctx, status.JOB_MISSING_NAME, nil, ErrJobNameRequired)
	}
	return nil
}

func validateTaskName(ctx context.Context, name string) error {
	if strings.TrimSpace(name) == "" {
		return errs.NewError(ctx, status.TASK_MISSING_NAME, nil, ErrTaskNameRequired)
	}
	return nil
}

// parseSchedule converts the wire form into a runtime Schedule,
// rejecting anything malformed with JOB_INVALID_SCHEDULE. Every failure
// carries a field-level reason in the debug channel — an operator
// retargeting a job needs to know WHICH field was wrong.
//
// Timezone resolution deliberately fails loudly on an unknown name
// rather than falling back to UTC the way jobs.loadProjectTimezone
// does. A silent fallback would ack a schedule the operator did not ask
// for and only surface hours later, when the job fires at the wrong
// local time.
func parseSchedule(ctx context.Context, req *ScheduleRequest) (jobruntime.Schedule, error) {
	if req == nil {
		return jobruntime.Schedule{}, errs.NewError(ctx, status.JOB_INVALID_SCHEDULE, nil, ErrScheduleRequired)
	}

	switch strings.ToLower(strings.TrimSpace(req.Kind)) {
	case "every":
		d := time.Duration(req.EverySeconds) * time.Second
		if req.EverySeconds < int(minEveryInterval.Seconds()) || req.EverySeconds > int(maxEveryInterval.Seconds()) {
			return jobruntime.Schedule{}, errs.NewError(ctx, status.JOB_INVALID_SCHEDULE, nil,
				fmt.Errorf("every_seconds must be within [%d,%d], got %d",
					int(minEveryInterval.Seconds()), int(maxEveryInterval.Seconds()), req.EverySeconds))
		}
		return jobruntime.EveryDuration(d), nil

	case "daily":
		loc, err := parseTimezone(ctx, req.Timezone)
		if err != nil {
			return jobruntime.Schedule{}, err
		}
		s := jobruntime.DailyAt(req.Hour, req.Minute, loc)
		if err := s.Validate(); err != nil {
			return jobruntime.Schedule{}, errs.NewError(ctx, status.JOB_INVALID_SCHEDULE, nil, err)
		}
		return s, nil

	case "weekly":
		if req.Weekday == nil {
			return jobruntime.Schedule{}, errs.NewError(ctx, status.JOB_INVALID_SCHEDULE, nil, ErrWeekdayRequired)
		}
		loc, err := parseTimezone(ctx, req.Timezone)
		if err != nil {
			return jobruntime.Schedule{}, err
		}
		s := jobruntime.WeeklyAt(time.Weekday(*req.Weekday), req.Hour, req.Minute, loc)
		if err := s.Validate(); err != nil {
			return jobruntime.Schedule{}, errs.NewError(ctx, status.JOB_INVALID_SCHEDULE, nil, err)
		}
		return s, nil

	default:
		return jobruntime.Schedule{}, errs.NewError(ctx, status.JOB_INVALID_SCHEDULE, nil,
			fmt.Errorf("%w: got %q", ErrScheduleKindUnknown, req.Kind))
	}
}

// parseTimezone resolves an IANA name against the tzdata embedded in
// the binary (cmd/mathsvr imports time/tzdata), so the result does not
// depend on whether the deploy host ships /usr/share/zoneinfo. Empty
// means UTC — an explicit choice the response echoes back.
func parseTimezone(ctx context.Context, name string) (*time.Location, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return time.UTC, nil
	}
	loc, err := time.LoadLocation(name)
	if err != nil {
		return nil, errs.NewError(ctx, status.JOB_INVALID_SCHEDULE, nil,
			fmt.Errorf("unknown timezone %q: %w", name, err))
	}
	return loc, nil
}
