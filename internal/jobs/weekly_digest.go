package jobs

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"math-ai.com/math-ai/internal/adapter/email"
	"math-ai.com/math-ai/internal/infrastructure/job"
	"math-ai.com/math-ai/internal/infrastructure/logger"
)

// Weekly digest is a two-piece workflow: a CronJob fans out per-user
// work as Tasks, the Task does one user's mail. Splitting it this way
// keeps the cron tick small (no risk of blowing through its timeout on
// a slow recipient) and gives us retry on a per-user basis.

// ============================================================
// CronJob: fan-out
// ============================================================

// WeeklyDigestCronJob fires Monday 09:00 ICT and enqueues one
// digest.send task per user. The user list is currently a stub — wiring
// the real query (probably UserRepository.ListUsers with email
// non-null) is a follow-up; the topology and retry behaviour are
// already correct.
type WeeklyDigestCronJob struct {
	enqueuer *job.Runtime
}

func NewWeeklyDigestCronJob(enqueuer *job.Runtime) *WeeklyDigestCronJob {
	return &WeeklyDigestCronJob{enqueuer: enqueuer}
}

const weeklyDigestCronName = "digest.weekly_fanout"

func (j *WeeklyDigestCronJob) Name() string { return weeklyDigestCronName }
func (j *WeeklyDigestCronJob) Schedule() job.Schedule {
	return job.WeeklyAt(time.Monday, 9, 0, projectTimezone)
}
func (j *WeeklyDigestCronJob) Timeout() time.Duration { return 5 * time.Minute }

func (j *WeeklyDigestCronJob) Run(ctx context.Context) error {
	log := logger.From(ctx)
	if j.enqueuer == nil {
		return errors.New("weekly_digest: no enqueuer wired")
	}

	// TODO(tier2): swap this for a UserRepository.ListUsers query that
	// streams (or paginates) user_id + email. Until that exists, the
	// fan-out is a no-op so we don't spam an empty task queue.
	userIDs := []string{}

	enqueued, dropped := 0, 0
	for _, uid := range userIDs {
		payload, err := json.Marshal(WeeklyDigestPayload{UserID: uid})
		if err != nil {
			log.Warnf("weekly_digest.marshal_failed uid=%s err=%v", uid, err)
			dropped++
			continue
		}
		if err := j.enqueuer.Enqueue(ctx, weeklyDigestTaskName, payload, job.TaskOptions{}); err != nil {
			log.Warnf("weekly_digest.enqueue_failed uid=%s err=%v", uid, err)
			dropped++
			continue
		}
		enqueued++
	}
	log.Infof("weekly_digest.fanout enqueued=%d dropped=%d total=%d", enqueued, dropped, len(userIDs))
	return nil
}

// ============================================================
// Task: per-user send
// ============================================================

// WeeklyDigestPayload is the JSON shape exchanged between the cron
// fan-out and the task. UserID is the external uuid; the task resolves
// it to a profile + email at run time.
type WeeklyDigestPayload struct {
	UserID string `json:"user_id"`
}

// WeeklyDigestTask renders and sends one user's weekly digest email.
// Body is a stub until the digest template + recipient lookup land;
// the retry policy and observability are already in place.
type WeeklyDigestTask struct {
	email *email.Adapter
}

func NewWeeklyDigestTask(emailAdapter *email.Adapter) *WeeklyDigestTask {
	return &WeeklyDigestTask{email: emailAdapter}
}

const weeklyDigestTaskName = "digest.send"

func (t *WeeklyDigestTask) Name() string { return weeklyDigestTaskName }

func (t *WeeklyDigestTask) Handle(ctx context.Context, payload []byte) error {
	log := logger.From(ctx)
	var msg WeeklyDigestPayload
	if err := json.Unmarshal(payload, &msg); err != nil {
		// Bad payload is a terminal failure — retrying won't fix it.
		// Wrap as a non-retryable surface once we add an error kind;
		// for now log and exit, the retry policy attempt cap caps the
		// blast radius.
		log.Errorf("weekly_digest.bad_payload err=%v", err)
		return err
	}
	if msg.UserID == "" {
		log.Errorf("weekly_digest.empty_user_id")
		return errors.New("weekly_digest: empty user_id")
	}
	if t.email == nil {
		log.Warnf("weekly_digest.skip uid=%s reason=email_adapter_disabled", msg.UserID)
		return nil
	}

	// TODO(tier2): resolve user/profile, build template, send via t.email.Send.
	log.Infof("weekly_digest.stub uid=%s — template + lookup pending", msg.UserID)
	return nil
}
