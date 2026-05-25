# Math-AI — Jobs Design Explained

> Audience: developers and AI agents working on `math-svr`.
> Source of truth: Go files under `internal/infrastructure/job/`,
> `internal/jobs/`, and `internal/module/job/`. When this document
> disagrees with the code, the code wins — open a PR to update this doc.

This document explains *why* the job runtime is shaped the way it is,
not just *what* each file contains. Pair it with `/.claude/CLAUDE.md`
(architectural guide) and the source files themselves.

---

## 1. What the job runtime is for

`math-svr` needs to run work that does **not** belong inside an HTTP
request:

- Periodic housekeeping (expire stale sessions, force-delete
  soft-deleted quizzes).
- Periodic recomputation (leaderboards).
- Fan-out + per-row delivery (weekly digest mails).
- Ad-hoc deferred work an HTTP handler wants to "fire and forget"
  (queue an email, schedule a retry).

The runtime owns the scheduling, the timeouts, the panic recovery, the
retry behaviour, the drain on shutdown, and the observability for that
work. Application code only writes the *body* of each job.

The current implementation is single-process and in-memory ("Tier 1").
The architecture is set up so a future PR can swap the executor for a
MySQL-backed queue with `SELECT ... FOR UPDATE SKIP LOCKED` without
changing consumers ("Tier 2"); see §10.

---

## 2. Two concepts — CronJob vs Task

The runtime distinguishes two kinds of work. Don't conflate them.

| | **CronJob** | **Task** |
|---|---|---|
| Purpose | Recurring work on a wall-clock schedule | One-shot work with a payload |
| Trigger | Scheduler fires it on `Schedule.Next` | Caller calls `Runtime.Enqueue(...)` |
| Payload | None | Opaque `[]byte` (typically JSON) |
| Retry | No (next tick is the next try) | Yes — `RetryPolicy` (constant or exponential) |
| Identity | Stable `Name()` | Stable `Name()` (the handler) |
| Concurrency | Skip-if-still-running (no overlap) | Worker pool, N in flight at once |
| Examples in repo | `system.session_cleanup`, `quiz.cleanup_soft_deleted`, `leaderboard.recalc`, `digest.weekly_fanout`, `system.noop` | `digest.send` (one user's digest) |

A common pattern is **cron-fans-out-tasks**: a weekly CronJob lists all
recipients and enqueues one Task per recipient. The cron stays small
and predictable, and per-user failure does not block the rest.

---

## 3. Package layout

```
internal/
├── infrastructure/job/      # the runtime — registry, scheduler, worker pool
│   ├── types.go             # Schedule, CronJob, TaskHandler, RetryPolicy, Snapshot DTOs
│   ├── registry.go          # name → CronJob / name → TaskHandler
│   ├── config.go            # tunables (workers, queue cap, drain timeout, defaults)
│   ├── runtime.go           # scheduler loops + worker pool + drain
│   └── errors.go            # sentinel errors (ErrJobNotFound, …)
│
├── jobs/                    # the *catalogue* — concrete job implementations
│   ├── jobs.go              # RegisterAll(reg, Deps) — single seam for adding a job
│   ├── timezone.go          # projectTimezone = Asia/Ho_Chi_Minh (UTC fallback)
│   ├── session_cleanup.go   # CronJob
│   ├── quiz_cleanup.go      # CronJob (stub body)
│   ├── leaderboard_recalc.go# CronJob (stub body)
│   ├── weekly_digest.go     # CronJob + Task pair
│   └── noop.go              # CronJob — hourly heartbeat
│
└── module/job/              # HTTP control plane
    ├── service.go           # wraps Runtime; sentinel errors → MathError
    ├── handler.go           # /jobs/list|trigger|pause|resume + /tasks/enqueue
    ├── dto.go               # request/response shapes
    └── validator.go
```

The package boundaries match the clean-architecture rule in
`.claude/CLAUDE.md`:

- `infrastructure/job/` is the runtime, depending only on stdlib and
  `infrastructure/logger`.
- `jobs/` is the catalogue. It imports `infrastructure/job` (interfaces)
  and may import application/adapter packages it needs to do work.
- `module/job/` is presentation. It depends on `infrastructure/job` for
  the runtime handle and on `domain/shared/{error,status}` for error
  translation.

`bootstrap/` is the only place that wires them together.

---

## 4. Schedule semantics

`Schedule` is a value-typed union with three constructors:

```go
job.EveryDuration(15 * time.Minute)              // "every 15 minutes from start"
job.DailyAt(3, 0, projectTimezone)               // "every day at 03:00 ICT"
job.WeeklyAt(time.Monday, 9, 0, projectTimezone) // "every Monday at 09:00 ICT"
```

`Next(now)` returns the next instant strictly after `now`:

- **EveryDuration** is *interval-anchored*: each next fire is the
  previous Next plus `d`. Across process restarts the cadence drifts by
  the restart time — fine for short intervals (15 min, 30 min).
- **DailyAt** / **WeeklyAt** are *wall-clock anchored* in the given
  timezone. Restart safe; the "nightly 03:00 cleanup" always fires at
  03:00 regardless of when the server booted.

The default project timezone is `Asia/Ho_Chi_Minh`, resolved once at
package init via `time.LoadLocation`. Containers without tzdata (Alpine
slim) fall back to UTC — explicit, no surprise — see `jobs/timezone.go`.

There is intentionally **no cron-string parser**. The three structured
constructors cover every scheduled workload `math-svr` has today; cron
strings are easy to typo and impossible to type-check. If you ever need
"every weekday at noon", add a fourth constructor; do not introduce
cron-string parsing without discussion.

---

## 5. Runtime internals

`Runtime` is the orchestrator. One `Runtime` per process.

### 5.1 Goroutine model

For N registered CronJobs and M workers, the runtime owns:

```
N  cron loops              — one per CronJob, sleeps until Schedule.Next, fires
M  task workers            — one per worker slot, reads from the bounded queue
k  delayed-enqueue timers  — transient: one per "enqueue with Delay" or one per task retry
```

All cron loops and task workers count against `workWg`. In-flight
`CronJob.Run` and `TaskHandler.Handle` calls count against `runWg`.
This separation is what lets `Stop()` distinguish "scheduler shells
exited" from "user code finished".

### 5.2 Skip-if-still-running

A slow cron run must not be double-fired by the next tick. Before each
fire the runtime atomically checks the `inFlight` flag; if set, the
tick logs `job.cron.skip_overlap` and waits for the next schedule. The
in-flight execution is *never* cancelled by the next tick — pause and
shutdown are the only ways to cancel.

### 5.3 Per-attempt timeout

Every CronJob declares `Timeout() time.Duration`. Zero means "no
timeout"; otherwise the runtime wraps the call in
`context.WithTimeout`. A job that returns *after* its timeout fires
has `last_status = "timed_out"`.

For tasks the timeout is per-attempt and comes from
`TaskOptions.Timeout`, falling back to
`Config.DefaultTaskTimeout` (60s).

### 5.4 Panic recovery

Every `Run` and `Handle` call is wrapped in `recover()`. A panic is
recorded as `last_status = "panicked"`, the stack trace is logged at
`Error`, and the goroutine exits cleanly. The runtime keeps running.

### 5.5 Retry

Tasks (not crons) carry a `RetryPolicy`:

```go
type RetryPolicy struct {
    Attempts int            // total attempts incl. first; 1 = no retry
    Backoff  BackoffKind    // BackoffConstant or BackoffExponential
    Base     time.Duration  // first delay
    MaxDelay time.Duration  // cap (only applies to exponential)
}
```

A failed/panicked task is re-enqueued via `scheduleDelayed(env, delay)`
which is a transient goroutine bound to `workWg` so `Stop` waits for
pending retries to either fire or abort on `stopCh`. If the queue is
full at re-enqueue time, the retry is **dropped and logged**
(`job.task.retry_dropped reason=queue_full`). Without persistence this
is the best we can do — Tier 2 will requeue from MySQL.

The runtime applies `Config.DefaultRetryPolicy`
(`Attempts=3, exponential, base=5s, cap=5m`) when `TaskOptions.Retry`
is the zero value.

### 5.6 Bounded queue

The task queue is a `chan` with capacity `Config.TaskQueueSize` (default
1024). `Enqueue` is non-blocking: a full queue returns
`ErrQueueFull` (`TASK_QUEUE_FULL` to clients). HTTP handlers calling
`Enqueue` therefore never stall on a backed-up worker pool.

### 5.7 Drain on shutdown

`Stop(ctx)` runs in two phases:

1. Close `stopCh`. This:
   - tells every cron loop to exit (no new fires),
   - tells every worker to stop pulling new tasks,
   - cancels every in-flight execution context (see the bridge in
     `executionContext`).
2. Wait for `workWg` (scheduler/worker shells), then wait for `runWg`
   (in-flight executions) up to `Config.DrainTimeout` (default 30s).

A well-behaved job sees `ctx.Done()` and returns promptly. A
misbehaving job that ignores cancellation extends the wait — Go has no
safe primitive for force-killing a goroutine, so we log
`job.runtime.drain_timeout_exceeded` and keep waiting.

The shutdown hook in `bootstrap/app.go` drains the runtime **before**
session serialization, so any job that touches a session has finished
before we snapshot the session store.

---

## 6. Concrete jobs that ship today

| Name | Kind | Schedule | Body | Source |
|---|---|---|---|---|
| `system.session_cleanup`     | Cron | every 15m         | Real — calls `SessionManager.DeleteExpiredSessions` | `internal/jobs/session_cleanup.go` |
| `quiz.cleanup_soft_deleted`  | Cron | daily 03:00 ICT   | Stub — needs `QuizRepository.ForceDeleteSoftDeletedBefore` | `internal/jobs/quiz_cleanup.go` |
| `leaderboard.recalc`         | Cron | every 30m         | Stub — leaderboard aggregate not yet built | `internal/jobs/leaderboard_recalc.go` |
| `digest.weekly_fanout`       | Cron | weekly Mon 09:00 ICT | Stub — enqueues one `digest.send` per user (user list pending) | `internal/jobs/weekly_digest.go` |
| `system.noop`                | Cron | hourly            | Hourly heartbeat — `system.noop.heartbeat` | `internal/jobs/noop.go` |
| `digest.send`                | Task | n/a               | Stub — one user's digest email (handler registration currently commented out in `RegisterAll`) | `internal/jobs/weekly_digest.go` |

> Stubs intentionally log and return nil. The schedule, timeout, retry
> policy, and observability surface are already correct; only the body
> is pending. When you fill in a stub, do not change the `Name()` — it
> is the addressing key and will become the distributed-claim key in
> Tier 2.

---

## 7. Adding a new job — recipe

### Adding a CronJob

1. Create `internal/jobs/<name>.go`:

   ```go
   package jobs

   type MyJob struct{ /* deps */ }

   func NewMyJob(/* deps */) *MyJob { return &MyJob{} }

   func (j *MyJob) Name() string           { return "my.job" }
   func (j *MyJob) Schedule() job.Schedule { return job.EveryDuration(10 * time.Minute) }
   func (j *MyJob) Timeout() time.Duration { return 2 * time.Minute }

   func (j *MyJob) Run(ctx context.Context) error {
       logger.From(ctx).Info("my.job.start")
       // ... ctx-aware work ...
       return nil
   }
   ```

2. Add deps to `jobs.Deps` (`internal/jobs/jobs.go`) if you need
   anything new. Keep them nil-tolerant if the underlying adapter is
   disable-friendly.

3. Add one line to `RegisterAll`:

   ```go
   reg.RegisterCron(NewMyJob(deps.Whatever))
   ```

4. `go build ./...` and run the server. The startup log line
   `job.runtime.start crons=N tasks=M ...` should show the new count.

### Adding a Task

1. Create the handler in `internal/jobs/<name>.go`:

   ```go
   type MyTask struct{ /* deps */ }
   func (t *MyTask) Name() string { return "my.task" }
   func (t *MyTask) Handle(ctx context.Context, payload []byte) error {
       var p MyPayload
       if err := json.Unmarshal(payload, &p); err != nil { return err }
       // ... do work ...
       return nil
   }
   ```

2. Register it in `RegisterAll`:

   ```go
   reg.RegisterTask(NewMyTask(deps.Whatever))
   ```

3. Enqueue from application code:

   ```go
   payload, _ := json.Marshal(MyPayload{ID: "..."})
   err := res.JobRuntime.Enqueue(ctx, "my.task", payload, job.TaskOptions{
       Delay:   30 * time.Second,
       Timeout: 2 * time.Minute,
       Retry:   job.RetryPolicy{Attempts: 5, Backoff: job.BackoffExponential, Base: 2*time.Second, MaxDelay: 5*time.Minute},
   })
   ```

   Or via the HTTP control plane (`POST /tasks/enqueue`) for ops.

### Naming convention

Use `<aggregate-or-domain>.<verb>` for the `Name()`:

- `system.session_cleanup`, `system.noop`
- `quiz.cleanup_soft_deleted`
- `digest.weekly_fanout`, `digest.send`

The dot-separated prefix lets you grep all jobs from one domain in
logs and (Tier 2) filter the runs table.

---

## 8. HTTP control plane

All routes are gated by `authMiddleware` (any authenticated user can
call them — see `.claude/rules/known-issues.md` §11 on pre-RBAC state).
Registered in `internal/bootstrap/routes/routes.go`:

| Method | Path | Purpose |
|---|---|---|
| `POST` | `/jobs/list`     | Full snapshot: every CronJob with status, next/last run, last error, in-flight flag; every registered task by name; queue gauges |
| `POST` | `/jobs/trigger`  | Body `{"name":"<job>"}` — fire a CronJob immediately, out of schedule. Returns `JOB_IN_FLIGHT` if already running |
| `POST` | `/jobs/pause`    | Body `{"name":"<job>"}` — suspend scheduling. In-flight executions are NOT cancelled |
| `POST` | `/jobs/resume`   | Body `{"name":"<job>"}` — reverse pause |
| `POST` | `/tasks/enqueue` | Body `{"name","payload","delay_seconds","attempts","timeout_seconds"}` — enqueue a one-shot task |

Responses follow the project envelope (`response.WriteJson`):
HTTP 200 with `mstatus` carrying the semantic outcome. Job-related
status codes live in the **12000 block** (`code.go`,
`message_en.go`, `message_vn.go` — kept in lockstep):

```
JOB_NOT_FOUND            12001
JOB_ALREADY_PAUSED       12002
JOB_NOT_PAUSED           12003
JOB_IN_FLIGHT            12004
JOB_RUNTIME_UNAVAILABLE  12005
JOB_MISSING_NAME         12006
JOB_TRIGGER_FAILED       12007
TASK_HANDLER_NOT_FOUND   12008
TASK_QUEUE_FULL          12009
TASK_MISSING_NAME        12010
TASK_ENQUEUE_FAILED      12011
```

The service layer (`module/job/service.go`) maps the runtime's
sentinel errors (`ErrJobNotFound`, `ErrJobInFlight`, …) to these codes;
the runtime itself stays free of presentation concerns.

---

## 9. Configuration

`job.Config` is constructed in `bootstrap/resource.go` and tunes the
runtime. A zero `Config` picks sensible defaults; all fields are
optional from the caller's perspective:

| Field | Default | Meaning |
|---|---|---|
| `TaskWorkers`        | 4               | Worker pool size |
| `TaskQueueSize`      | 1024            | Bounded channel capacity |
| `DrainTimeout`       | 30s             | Wait budget for in-flight executions on `Stop()` |
| `DefaultTaskTimeout` | 60s             | Per-attempt cap when caller doesn't set `TaskOptions.Timeout` |
| `DefaultRetryPolicy` | 3 / exp / 5s / 5m | Applied when `TaskOptions.Retry` is the zero value |
| `DefaultTimezone`    | not used yet    | Reserved — jobs read `projectTimezone` from `jobs/timezone.go` |

If you need to override these per environment, plumb them through
`config.Env` (see how `BotConfig` does it) — but don't bother until
real workloads tell you which knob matters.

---

## 10. Tier 1 limitations — what is NOT here yet

The current runtime is single-process and in-memory. Things it
**does not** do:

- **Persistence**: `last_run_at`, retry queue depth, in-flight
  executions all evaporate on restart. The schedule for `DailyAt` /
  `WeeklyAt` is recomputed from wall-clock so the cadence survives;
  the *history* does not.
- **Horizontal scale**: every replica runs every CronJob. Running two
  replicas in production today means every cron fires twice. The
  system is **deployed single-replica today**.
- **Separate process**: jobs run inside the HTTP binary. An OOM in the
  HTTP server kills the worker pool; a wedged job (e.g. an LLM hang)
  delays HTTP shutdown.
- **At-least-once delivery for tasks**: a process crash mid-execution
  loses the in-flight task. Idempotency is the application's
  responsibility — write task bodies that can safely run twice.

These are all addressed in Tier 2 (see §11).

---

## 11. Tier 2 — what the seams are designed for

The next PR introduces persistence and a distributed-safe claim. The
shape of the change:

1. **Migrations**:
   - `ma_jobs` — schedule definitions (name, schedule_kind, schedule_args, enabled, next_fire_at)
   - `ma_job_runs` — executions (name, status, attempt, started_at, finished_at, locked_by, lock_until, error, payload)
2. **Repos** under `infrastructure/persistence/mysql/repositories/`
   following the column-constants + `findOneBy` + active-where pattern.
3. **Claim**: `SELECT … FOR UPDATE SKIP LOCKED LIMIT N` on `ma_job_runs`.
   MySQL 8 supports this natively.
4. **Reaper**: a goroutine that requeues runs whose `lock_until < now()`,
   so a crashed worker's task is picked up by another within seconds.
5. **Atomic enqueue**: `Enqueue` accepts a `transaction.UnitOfWork`
   handle so the row insert is in the same transaction as its trigger
   (e.g. user signs up → digest row enqueued — both commit or neither).
6. **Out-of-process worker** (Tier 3): `cmd/worker/main.go` runs the
   same `Runtime` against the MySQL queue, off the HTTP binary's fate.

Why the current shape supports this with **no breaking change to
consumers**:

- The `CronJob` / `TaskHandler` interfaces stay identical.
- `Runtime.Enqueue` already accepts `ctx` — Tier 2 adds an overload
  taking the UoW.
- `JobInfo` is already shaped to mirror `ma_job_runs`, so
  `/jobs/list` responses do not change.
- Job addressing is by stable `Name()` — the distributed-claim key
  is already in the right place.
- The runtime's executor is internal; swapping `queue chan` for a
  `mysqlQueueClient` is a Tier 2 commit that touches only files under
  `infrastructure/job/`.

---

## 12. Observability

The runtime emits structured logs via `logger.From(ctx)`. Key event
names (greppable):

- `job.runtime.start crons=N tasks=M workers=K queue_cap=Q drain=30s`
- `job.runtime.stopping` / `shells_drained` / `executions_drained` /
  `executions_drained_late` / `drain_timeout_exceeded`
- `job.cron.completed name=… triggered=… dur_ms=…`
- `job.cron.timed_out` / `job.cron.failed` / `job.cron.panicked`
- `job.cron.skip_overlap name=…` — fired tick was skipped because a
  previous run is still in flight; investigate if you see it often
- `job.task.completed name=… attempt=… worker=… dur_ms=…`
- `job.task.failed_terminal name=… attempt=… err=…` — retries
  exhausted
- `job.task.retry_scheduled name=… attempt=X->Y delay=… err=…`
- `job.task.retry_dropped name=… reason=queue_full` — should be rare;
  tune `TaskQueueSize` if persistent
- `system.noop.heartbeat` — confirms the runtime is alive (logged
  hourly)

Metrics (counter / histogram) are intentionally not exported yet — the
project has no metrics infra wired (`OTEL_*` env vars exist in
`.env.example` but no provider is initialized; see
`.claude/rules/known-issues.md` §7). When that lands, the natural
metric names are `job.run.total{name,status}` and
`job.run.duration_ms{name}`.

---

## 13. Operational guidance

### "I just deployed and a cron didn't fire"

1. `POST /jobs/list` — confirm the job exists, is not paused, and
   `next_run_at` is in the future.
2. Check the start-time log line: `job.runtime.start crons=N ...` —
   does N match the registry?
3. For `DailyAt` / `WeeklyAt` schedules, check the server timezone.
   The runtime uses `projectTimezone` (Asia/Ho_Chi_Minh) **inside the
   job's Schedule** but `time.Now()` uses the process timezone for the
   "is now after next" comparison — they have to agree. The fallback
   is UTC; tzdata in the container matters.

### "Tasks are accumulating in the queue"

1. `POST /jobs/list` → `queue.depth` / `queue.capacity` — how close to
   full?
2. Check worker logs for stuck tasks (no `job.task.completed` in
   minutes). Bump `Config.TaskWorkers` or fix the slow handler.
3. If you see `job.task.retry_dropped reason=queue_full`, the queue
   is filling faster than workers drain it — task body needs to be
   faster, the workload needs more workers, or the upstream needs
   throttling.

### "Pause a runaway job"

`POST /jobs/pause` with the job name stops future fires; the
**currently running execution is not cancelled** (pause is "stop
scheduling", not "abort"). Wait for the in-flight run to finish (watch
`in_flight=false` in `/jobs/list`) or accept the next graceful drain
on deploy.

### "Force shutdown is taking forever"

The runtime waits up to `DrainTimeout` (30s) for in-flight executions
to honour ctx cancellation. If you see
`job.runtime.drain_timeout_exceeded`, find the misbehaving job — it is
not respecting `ctx.Done()`. Cron bodies that do long-running work
must check `ctx.Err()` between phases.

---

## 14. References

- `.claude/CLAUDE.md` — top-level architectural index
- `.claude/rules/architecture.md` — layer rules the job runtime obeys
- `.claude/rules/conventions.md` §Errors, §Status codes, §Logging —
  conventions the runtime uses
- `.claude/rules/known-issues.md` §11 (pre-RBAC), §7 (no OTEL), §13
  (job runtime status) — context for the operational gaps called out
  above
- `docs/DatabaseDesignExplain.md` — where Tier 2's `ma_jobs` /
  `ma_job_runs` will fit
