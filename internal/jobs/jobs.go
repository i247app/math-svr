// Package jobs is the concrete-job catalogue.
//
// One file per job kind. `RegisterAll` is the single seam bootstrap
// calls — it gets a populated Registry plus a Deps bundle of services
// and decides which jobs to enable. Adding a new job is: write a new
// file in this package, append one `reg.RegisterCron` or
// `reg.RegisterTask` line to RegisterAll, done.
//
// Jobs live here (not under module/job) because they are not handlers
// — they are the runtime's consumers of domain/application services.
// Concretely, they import application/command/* like any module
// service would, and may reach side-effect adapters (email, storage,
// bot) directly. They never see *sql.DB.
package jobs

import (
	"math-ai.com/math-ai/internal/adapter/email"
	"math-ai.com/math-ai/internal/infrastructure/job"
	"math-ai.com/math-ai/internal/infrastructure/session"
)

// Deps bundles the cross-cutting services concrete jobs need. Add a
// field here whenever a new job requires a new dependency; the
// alternative — fishing services out of a Resource pointer at run
// time — hides the wiring. Keep nil-tolerant where the underlying
// adapter is disable-friendly (email/SMS/bot/storage), so a dev that
// turns one off in .env can still boot.
type Deps struct {
	// Runtime is the same *job.Runtime that owns the Registry. Cron
	// jobs that fan out per-row work (e.g. weekly_digest) use it as a
	// task enqueuer; pure cron jobs ignore it.
	Runtime *job.Runtime

	// SessionManager is used by session_cleanup. Never nil — sessions
	// are core infrastructure.
	SessionManager *session.SessionManager

	// EmailProvider may be nil when the email adapter is disabled in
	// .env. Jobs that depend on it must nil-guard and degrade.
	EmailProvider *email.Adapter
}

// RegisterAll wires every concrete job into reg. Called once at boot,
// before Runtime.Start. Order does not affect runtime behaviour —
// each job has its own scheduler loop — but the source order here is
// a useful "menu" of what runs in production.
func RegisterAll(reg *job.Registry, deps Deps) {
	// Cron jobs (recurring, no payload).
	// reg.RegisterCron(NewSessionCleanupJob(deps.SessionManager))
	// reg.RegisterCron(NewQuizCleanupJob())
	// reg.RegisterCron(NewWeeklyDigestCronJob(deps.Runtime))
	reg.RegisterCron(NewNoopJob())

	// Tasks (one-shot, payload).
	// reg.RegisterTask(NewWeeklyDigestTask(deps.EmailProvider))
}
