package jobs

import (
	"context"
	"time"

	userCommand "math-ai.com/math-ai/internal/application/command/user"
	"math-ai.com/math-ai/internal/infrastructure/job"
	"math-ai.com/math-ai/internal/infrastructure/logger"
)

// GuestCleanupJob retires guest accounts nobody came back for.
//
// A guest is opened the first time a visitor generates an exam without
// registering. Most never register, so without this the rows accumulate
// forever. Everything it does is a SOFT delete — `deleted_dt` plus a
// status flip — so a sweep run with the wrong cutoff is recoverable, and
// a guest who registered or was adopted is not a candidate at all
// (their identity_code left GUEST at that moment).
type GuestCleanupJob struct {
	cleanup *userCommand.CleanupGuestsCommandHandler
}

func NewGuestCleanupJob(cleanup *userCommand.CleanupGuestsCommandHandler) *GuestCleanupJob {
	return &GuestCleanupJob{cleanup: cleanup}
}

const guestCleanupName = "user.guest_cleanup"

const (
	// GuestIdleRetention is how long a guest may sit untouched before it
	// is retired. Measured from the last exam they were handed, not from
	// when the account was opened.
	GuestIdleRetention = 30 * 24 * time.Hour
	// guestCleanupBatch caps one sweep. It is a ceiling, not a target:
	// the daily run is expected to find a handful, and the cap is what
	// stops a mistaken cutoff from retiring everything at once.
	guestCleanupBatch = 500
)

func (j *GuestCleanupJob) Name() string { return guestCleanupName }

// Schedule: once a day, in the small hours. Nothing here is urgent —
// a guest idle for 30 days is no more urgent at 03:00 than at 03:30 —
// and off-peak keeps the sweep away from the generate traffic it shares
// a database with.
func (j *GuestCleanupJob) Schedule() job.Schedule {
	return job.DailyAt(3, 0, time.UTC)
}

func (j *GuestCleanupJob) Timeout() time.Duration { return 5 * time.Minute }

func (j *GuestCleanupJob) Run(ctx context.Context) error {
	log := logger.From(ctx)
	if j.cleanup == nil {
		log.Warn("guest_cleanup.skip reason=no_handler")
		return nil
	}

	res, err := j.cleanup.Handle(ctx, userCommand.CleanupGuestsCommand{
		IdleFor: GuestIdleRetention,
		Limit:   guestCleanupBatch,
	})
	if err != nil {
		return err
	}

	// Always log, including the zero case: "the sweep ran and found
	// nothing" and "the sweep did not run" look identical otherwise.
	log.Infof("guest_cleanup.swept retired=%d idle_days=%d",
		len(res.Retired), int(GuestIdleRetention.Hours()/24))
	return nil
}
