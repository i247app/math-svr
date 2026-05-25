package jobs

import (
	"context"
	"time"

	"math-ai.com/math-ai/internal/infrastructure/job"
	"math-ai.com/math-ai/internal/infrastructure/logger"
	"math-ai.com/math-ai/internal/infrastructure/session"
)

type SessionCleanupJob struct {
	sm *session.SessionManager
}

func NewSessionCleanupJob(sm *session.SessionManager) *SessionCleanupJob {
	return &SessionCleanupJob{sm: sm}
}

const sessionCleanupName = "system.session_cleanup"

func (j *SessionCleanupJob) Name() string           { return sessionCleanupName }
func (j *SessionCleanupJob) Schedule() job.Schedule { return job.EveryDuration(15 * time.Minute) }
func (j *SessionCleanupJob) Timeout() time.Duration { return 30 * time.Second }

func (j *SessionCleanupJob) Run(ctx context.Context) error {
	log := logger.From(ctx)
	if j.sm == nil {
		log.Warn("session_cleanup.skip reason=no_session_manager")
		return nil
	}

	before := len(*j.sm.Sessions())
	j.sm.DeleteExpiredSessions()
	after := len(*j.sm.Sessions())
	log.Infof("session_cleanup.swept before=%d after=%d removed=%d", before, after, before-after)
	return nil
}
