package jobs

import (
	"context"
	"time"

	"math-ai.com/math-ai/internal/infrastructure/job"
	"math-ai.com/math-ai/internal/infrastructure/logger"
	"math-ai.com/math-ai/internal/infrastructure/session"
)

// ExpiredSessionSweeper deletes sessions no request can reach any more and
// reports how many it removed. gex's JwtSessionProvider implements it; the XWT
// provider does not (an XWT token carries no trustworthy expiry), so under
// that driver the sweeper is nil and the job does nothing.
type ExpiredSessionSweeper interface {
	SweepExpired() int
}

// SessionCleanupJob frees the memory held by sessions whose JWT has expired.
// Every request without a token opens a session, and an expired token is
// answered with a new one, so without this sweep the in-memory store only
// grows. Removing an unreachable session changes no request's outcome.
type SessionCleanupJob struct {
	sweeper ExpiredSessionSweeper
	sm      *session.SessionManager
}

func NewSessionCleanupJob(sweeper ExpiredSessionSweeper, sm *session.SessionManager) *SessionCleanupJob {
	return &SessionCleanupJob{sweeper: sweeper, sm: sm}
}

const sessionCleanupName = "system.session_cleanup"

func (j *SessionCleanupJob) Name() string           { return sessionCleanupName }
func (j *SessionCleanupJob) Schedule() job.Schedule { return job.EveryDuration(15 * time.Minute) }
func (j *SessionCleanupJob) Timeout() time.Duration { return 30 * time.Second }

func (j *SessionCleanupJob) Run(ctx context.Context) error {
	log := logger.From(ctx)
	if j.sweeper == nil {
		log.Warn("session_cleanup.skip reason=no_expiry_sweeper")
		return nil
	}

	removed := j.sweeper.SweepExpired()
	// A swept session may have been persisted (it had a uid); rewrite the
	// file so a restart does not load it back.
	if removed > 0 && j.sm != nil {
		j.sm.MarkDirty()
	}
	log.Infof("session_cleanup.swept removed=%d", removed)
	return nil
}
