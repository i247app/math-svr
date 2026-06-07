package jobs

import (
	"context"
	"time"

	"math-ai.com/math-ai/internal/infrastructure/job"
	"math-ai.com/math-ai/internal/infrastructure/logger"
)

type QuizCleanupJob struct{}

func NewQuizCleanupJob() *QuizCleanupJob { return &QuizCleanupJob{} }

const quizCleanupName = "quiz.cleanup_soft_deleted"

// quizRetention is the grace window between soft-delete and physical
// delete. 30 days lets ops restore an accidentally-deleted quiz.
const quizRetention = 30 * 24 * time.Hour

func (j *QuizCleanupJob) Name() string { return quizCleanupName }
func (j *QuizCleanupJob) Schedule() job.Schedule {
	return job.DailyAt(3, 0, loadProjectTimezone("Asia/Ho_Chi_Minh"))
}
func (j *QuizCleanupJob) Timeout() time.Duration { return 10 * time.Minute }

func (j *QuizCleanupJob) Run(ctx context.Context) error {
	log := logger.From(ctx)
	cutoff := time.Now().Add(-quizRetention)
	log.Infof("quiz_cleanup.stub cutoff=%s — repository method pending", cutoff.Format(time.RFC3339))
	return nil
}
