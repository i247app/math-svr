# math-svr

## Overview

`math-svr` (Go module `math-ai.com/math-ai`) is the HTTP backend for **Math-AI**, an
AI-powered math quiz platform for **Vietnamese primary-school students (Grades 1–5)**.
The audience is stated explicitly in the LLM prompt templates — e.g.
`internal/domain/bot/prompt_templates_en.go`: *"You are a math quiz generator for
Vietnamese primary-school students (Grades 1-5)."*

What the server does, as implemented in `internal/module/` and
`internal/bootstrap/routes/routes.go`:

- **Accounts & auth** — phone-first signup (`/users/create`) and an OTP login flow
  (`/auth/login` → `/auth/otp` → `/auth/logout`). A session starts *unsecure* and is
  promoted to *secure* only after OTP verification; device registration and a
  `ma_login_logs` audit row are written in the same step. OTP delivery is routed to
  SMS or email depending on the identifier shape (`internal/adapter/otp_delivery/`).
- **Learner profiles** — a user account holds one or more profiles (`ma_profiles`),
  each pinned to a curriculum program, grade, semester, and optionally a school.
- **Curriculum reference data** — programs, grades, semesters, chapters
  (`/programs/*`, `/grades/*`, `/semesters/*`, `/chapters/*`). Only `chapter` is
  bilingual (`ma_chapter_translations`).
- **AI quizzes** — `/quizzes/generate` builds a quiz through the bot adapter using
  curriculum context; `/quizzes/submit` grades deterministically, while
  `/quizzes/submit/cost-ai` grades through the LLM. Quizzes carry a *purpose*
  (`ASSESSMENT` / `PRACTICE` / `EXAM`) and a *type* (`GENERAL` / `REINFORCEMENT`,
  where reinforcement quizzes target previously-missed material) — see
  `internal/shared/enum/quiz.go` and `internal/domain/bot/prompts.go`.
- **Classrooms** — teacher-owned classrooms (`/classrooms/*`) with members, roles,
  archiving, invite codes, teacher-initiated invitations and student-initiated join
  requests (both are `ma_classroom_members` rows distinguished by `member_status`),
  and a per-profile learning-progress view.
- **Classroom exercises** — teacher-issued, AI-generated assignments
  (`/classroom-exercises/*`) with one submission per member, teacher roster views of
  submitted vs. non-submitted members, and the same deterministic/AI grading pair
  (`/classroom-exercise/submissions/submit` and `.../submit/cost-ai`).
- **Notifications** — a per-user in-app inbox (`/notifications/*`) with optional
  Firebase push delivery.
- **Home dashboard & banners** — `/home/layout` composes a role-aware dashboard for
  an acting profile; `/banners/*` manages promotional banner content (`ma_banners`).
- **Realtime** — an optional multiplexed WebSocket channel at `GET /ws/connect`,
  registered only when `SOCKET_ENABLED=true`.
- **Operations** — background job runtime (`/jobs/*`, `/tasks/enqueue`), session
  inspection (`/sessions/*`), liveness (`POST /ping`), and a graceful shutdown
  endpoint (`/server/shutdown`).

API conventions worth knowing before calling the server:

- **Almost every endpoint is `POST` with a JSON body.** The only `GET` is
  `/ws/connect`; even the health probe is `POST /ping`. Single-entity reads are
  `POST /<aggregate>/detail` with the external id in the body.
- **Responses always return HTTP 200.** The semantic outcome lives in the JSON body
  as `mstatus` / `mmessage` (`internal/shared/response/response.go`).
- **The session token is read only from the request body**, at
  `metadata.authorization`. `SessionTokenMiddleware` drops any client-supplied
  `Authorization` header before lifting the body value into it. The WebSocket
  handshake is the exception — it authenticates via a real `Authorization: Bearer`
  header.

## Setup

### Prerequisites

- **Go** — `go.mod` declares `go 1.25.0`.
- **MySQL 8** — collation `utf8mb4_0900_ai_ci`. The driver pins **TLS 1.2**
  (`internal/infrastructure/database/`), so the server must be reachable over TLS 1.2.
- **A HMAC key file** — any file whose bytes sign session JWTs; `hmac.key` is present
  at the repo root and is the conventional path.
- **A Gmail service-account JSON** — the email adapter is *mandatory at boot*
  (see the note below).
- **Docker** — only for the optional local observability stack. There is no
  Dockerfile for the application itself.

### 1. Configure `.env`

The binary loads `.env` from the working directory (`bootstrap.NewFromEnv(".env")` in
`cmd/mathsvr/main.go`).

```bash
cp .env.example .env
```

> **`.env.example` has drifted from the loader** (`internal/infrastructure/config/config.go`).
> Keys in the example file that the code **does not read**: `ROOT_SHARED_KEY`,
> `ROOT_SESSION_DRIVER`, `PORT`, `LOG_FILE_PATH`, `DB_ROOT_PASSWORD`,
> `MAIL_*`, `TWILIO_FROM_PHONE_NUMBER`. Keys the code reads that the example file
> **does not contain**: `GEX_SHARED_KEY`, `GEX_SESSION_DRIVER`, `EMAIL_PROVIDER`,
> `GMAIL_CREDENTIALS_FILE`, `GMAIL_SENDER_EMAIL`, `SMS_PROVIDER`, `STORAGE_*`,
> `TWILIO_FROM`, `TWILIO_BASE_URL`, `ENABLE_OTP`, `DB_MAX_OPEN_CONNS` and the other
> pool keys. Add the required ones by hand.

**Required** — the loader calls `panic()` if any of these is missing or empty:

| Key | Notes |
|---|---|
| `SERVER_HOST`, `SERVER_PORT` | API listener |
| `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASS`, `DB_NAME` | note `DB_PASS`, not `DB_PASSWORD` |
| `GEX_SHARED_KEY` | **path to a file** (e.g. `hmac.key`); the process panics if it cannot be read |
| `EMAIL_PROVIDER` | must be `google` — it is the only case in `internal/adapter/email/factory.go`; any other value fails boot |

Because `EMAIL_PROVIDER` is required and only `google` is supported,
`GMAIL_CREDENTIALS_FILE` and `GMAIL_SENDER_EMAIL` must also point at a valid Gmail
service-account JSON — `internal/libs/email/google_email_client.go` returns an error
if the file is missing or unparsable, and that error aborts startup.

**Optional / disable-friendly** — these adapters return `nil` when set to `""` or
`"disabled"`, and consumers nil-guard them:

| Key | Values |
|---|---|
| `SMS_PROVIDER` | `twilio` \| `""` \| `disabled` (+ `TWILIO_ACCOUNT_SID`, `TWILIO_AUTH_TOKEN`, `TWILIO_FROM`, `TWILIO_MESSAGING_SERVICE_SID`) |
| `BOT_PROVIDER` | `langchain` \| `eino` \| `""` \| `disabled`; the backend vendor is picked by `BOT_LANGCHAIN_BACKEND` / `BOT_EINO_BACKEND` ∈ `googleai` \| `openai` \| `anthropic` \| `ollama` |
| `NOTIFICATION_PROVIDER` | `firebase` \| `""` \| `disabled` (+ `FIREBASE_CREDENTIALS_FILE`, `FIREBASE_PROJECT_ID`) |
| `STORAGE_PROVIDER` | `s3` or empty (empty is treated as S3) + `STORAGE_ACCESS_KEY`, `STORAGE_SECRET_KEY`, `STORAGE_REGION`, `STORAGE_BUCKET` |
| `SOCKET_ENABLED` | `false` → `/ws/connect` is not registered at all |
| `ENABLE_OTP` | default `false`; passed into the auth handler |
| `GEX_SESSION_DRIVER` | `xwt` → XWT session provider; anything else → JWT |
| `SERIALIZED_SESSION_FILE` | when set, sessions are dumped here on shutdown and reloaded on start |
| `OBS_*`, `LOG_*` | observability — see `.env.example` and `docker/README.md` |

### 2. Apply database migrations manually

**Migrations are not applied on boot.** The `database.Migrate(ctx, sqlDB, "migrations")`
call in `internal/bootstrap/app.go` is commented out, so the SQL files under
`migrations/` (`000_ma_seqs_table.sql` … `022_ma_banners.sql`, applied in
lexicographic order) must be run by hand:

```bash
for f in migrations/*.sql; do mysql -h "$DB_HOST" -P "$DB_PORT" -u "$DB_USER" -p"$DB_PASS" "$DB_NAME" < "$f"; done
```

Then **seed `ma_seqs`**: the `INSERT` statements in `migrations/000_ma_seqs_table.sql`
are commented out, and every create operation mints its external id from that table —
without a seed row the operation fails.

> Migrations are forward-only; there are no down-migrations.
> `bin/create_migration.sh` exists but writes into `migrations/up/` and
> `migrations/down/`, which does not match the flat `migrations/*.sql` layout the
> runner reads. **[CẦN XÁC NHẬN]** whether that script is still intended for use.

### 3. Run

```bash
make run
```

Other targets defined in the `Makefile` (`make help` lists them all):

| Command | Effect |
|---|---|
| `make run` | `go mod tidy` then `go run ./cmd/mathsvr` |
| `make run-dev` | same via `air` hot-reload (requires `air` on `PATH`; the build command is passed inline — there is no `.air.toml`) |
| `make build` | `go build -o dist/mathsvr ./cmd/mathsvr` |
| `make build-ec2-arm` / `make build-ec2-amd` | cross-compile for Linux arm64 / amd64 |
| `make tidy` | `go mod tidy` |
| `make linecount` | line count over `internal` and `pkg` |

Startup order is visible in the logs: config → logger → tracing → DB connect →
adapters → routes.

### 4. Optional — local observability stack

```bash
make obs-up
```

Starts Prometheus, Tempo, Loki, Alloy, and Grafana from `docker/docker-compose.yml`.
Grafana → `http://localhost:3000` (`admin` / `admin`), Prometheus → `http://localhost:9090`.
The app runs on the **host**; Prometheus scrapes it at `host.docker.internal:9091`
(the dedicated metrics listener, `OBS_METRICS_ADDR`). Pair it with
`OBS_TRACING_ENABLED=true`, `LOG_FILE=./logs/app.log`, `LOG_FILE_FORMAT=json`.
`make obs-logs`, `make obs-down`, `make obs-reset` manage the stack.

### 5. Optional — deploy and database helpers

Deployment targets an EC2 host over SSH and requires an `.env.ec2-credentials` file
(`SSH_KEY`, `USER`, `HOST`, or `HOST1`…`HOST4` for the `t1`–`t4` targets); see
`.env.ec2-credentials.example`.

```bash
make deploy RHOST=t1
```

`bin/deploy.sh` runs validate → build → prepare → deliver → activate and prompts for
confirmation. Related targets: `make deploy-quick` (skip build), `make deploy-amd`,
`make deploy-rollback`, `make watch-logs RHOST=…`, `make login RHOST=…`,
`make connect-mysql`, and `make clear-data-local` / `make clear-data-ec2` (which run
`sql/clear_data.sql` — destructive).

## Architecture

### Layering

Clean architecture with a CQRS-lite split. Dependencies point inward; only
`bootstrap/` is allowed to import everything.

```
cmd/mathsvr                        entry point — bootstrap.NewFromEnv(".env") → app.Start()
        ↓
internal/bootstrap/                composition root: app lifecycle, DI container, middleware, routes
        ↓
internal/module/<agg>/             HTTP layer: handler → service → validator
        ↓
internal/application/              command/ (writes, via UnitOfWork) · query/ (reads, via repo)
  {command,query,dto}/<agg>/       transaction/ (UnitOfWork + Repositories) · resource/ (Resource)
        ↓
internal/domain/<agg>/             entities, repository interfaces, status codes, MathError, MathTime
        ↑
internal/infrastructure/           MySQL repositories implement the domain interfaces;
                                   database, logger, config, session, metrics, tracing, httproute, job, socket
internal/adapter/<kind>/           the only layer that talks to a third-party SaaS
internal/libs/<vendor>/            raw vendor clients (s3, email, twilio, firebase, langchain, eino)
```

Supporting packages: `internal/jobs/` (concrete background jobs) and
`internal/shared/` (`enum`, `pagination`, `response`, `metadata` helpers, `utils`).

### Request pipeline

Middleware is registered in `App.setupMiddleware` (`internal/bootstrap/app.go`),
outermost first:

```
MetricsMiddleware        → RED metrics over the full request lifetime (no-op when metrics disabled)
TracingMiddleware        → opens the request span before the logger, so log lines carry trace_id
SessionTokenMiddleware   → drops any client Authorization header, lifts metadata.authorization from the body
GexSessionMiddleware     → deserializes the JWT/XWT token into a session on the context
LoggerMiddleware         → per-request logger, reached everywhere via logger.From(ctx)
LogRequestMiddleware     → snapshots request/response bodies with secret redaction
MetadataMiddleware       → binds IP, user agent, device metadata, trace id into the context
RecoveryMiddleware       → converts panics into a JSON envelope
GzipMiddleware           → gzip when the client asks for it
```

Routes are registered in `internal/bootstrap/routes/routes.go` through a local
`reg(spec, handler, mw...)` wrapper that calls `gexSvr.AddRoute` **and** mirrors the
spec into `res.RouteClassifier`, so metrics labels and span names use the route
template rather than the concrete path. Per-route auth is
`middleware.AuthRequiredMiddleware(res.SessionManager)`, which requires a *secure*
(OTP-verified) session.

Routes intentionally registered **without** auth today: `POST /ping`,
`/misc/logs-time-format`, `/users/create`, `/auth/login`, `/auth/login-resume`,
`/auth/otp`, `/otps/send`, `/otps/verify`, `/programs/list`, `/grades/list`,
`/semesters/list`, `/ai/shake`, all of `/devices/*`, `/sessions/dump`, and
`/sessions/delete-all`.

### Persistence and the Unit of Work

All SQL lives in `internal/infrastructure/persistence/mysql/repositories/`
(one repository per aggregate; modules never see `*sql.DB`). Writes go through
`transaction.UnitOfWork.Do`, implemented by `SqlUnitOfWork`, which opens a `*sql.Tx`
and hands the callback a `transaction.Repositories` bundle bound to it:

> User, Alias, Profile, LoginLog, Device, Otp, Quiz, Chapter, ChapterTranslation,
> Grade, Semester, Program, School, Seq, Classroom, ClassroomMember,
> ClassroomProgram, Exercise, ExerciseSubmission, Notification, Banner.

`Seq` is in the bundle by design: **all ids are `int64`**, and each aggregate's
external id (`<entity>_id`) is minted from the central `ma_seqs` registry inside the
same transaction as the insert. Tables use the `ma_` prefix, an internal
`BIGINT UNSIGNED` primary key plus a unique external id, dual status columns
(`status` for system visibility, `<entity>_status` for business lifecycle), and
`DATETIME(6)` audit timestamps. There are no foreign keys — integrity is enforced in
the application layer inside `UnitOfWork` blocks.

### Adapters

Each external-I/O kind follows the same Adapter + Provider + factory shape, where the
factory is the only place that knows the vendor:

| Adapter | Vendors | Boot behavior |
|---|---|---|
| `email` | Gmail API (`google`) | **required** — boot fails without valid credentials |
| `sms` | Twilio | `nil` when disabled |
| `storage` | AWS S3 | used for avatar/static-file uploads |
| `bot` | LangChain (`tmc/langchaingo`) and eino (`cloudwego/eino`), each over googleai / openai / anthropic / ollama | `nil` when disabled; every framework whose `BOT_<X>_BACKEND` is set registers, `BOT_PROVIDER` picks the default |
| `notification` | Firebase | `nil` when disabled |
| `otp_delivery` | composite over `sms` + `email` | routes by identifier: contains `@` → email, else → SMS |

The quiz and exercise modules reach the LLM only through their own
`bot_service.go`, which owns prompt construction (`internal/domain/bot/`), JSON-mode
enforcement, and response parsing.

### Runtime components inside `cmd/mathsvr`

One process hosts all of the following (`cmd/woker/main.go` is a non-buildable stub —
`package woker; // comming soon` — and there is no separate worker binary):

- **API server** — `gex.Server` on `SERVER_HOST:SERVER_PORT`, HTTPS when
  `HTTPS_CERT_FILE` / `HTTPS_KEY_FILE` are set.
- **Metrics listener** — a separate middleware-free `http.Server` on
  `OBS_METRICS_ADDR` (default `:9091`) serving only `GET /metrics` (OpenMetrics, for
  trace-id exemplars) and `GET /ping`. `/metrics` is deliberately not an app route.
- **Job runtime** — `internal/infrastructure/job/` (registry + cron scheduler + task
  worker pool), fed by `internal/jobs/RegisterAll` and exposed over `/jobs/*` and
  `/tasks/enqueue`. In-memory, no persistent queue. **Only `NewNoopJob()` is
  registered today**; the session-cleanup, quiz-cleanup, and weekly-digest
  registrations are commented out, so triggering them returns a not-found status.
- **WebSocket hub** — `internal/infrastructure/socket/` (Hub + per-connection read/write
  pumps) behind the `application/socket` `Publisher` port and the `module/socket`
  handler/authorizer. Built only when `SOCKET_ENABLED=true`.
- **Sessions** — in-process `SessionManager`, JWT or XWT provider signed with the
  `GEX_SHARED_KEY` bytes, 14-day TTL, optionally serialized to disk across restarts.

Shutdown (`OnShutdown` + `App.Close`) closes WebSocket connections with
`StatusGoingAway`, drains the job runtime, serializes sessions, then stops the metrics
server, flushes traces, closes the DB pool, and closes the log file.

### Observability

- **Metrics** — Prometheus via `client_golang`, on the dedicated listener above.
- **Traces** — OpenTelemetry over OTLP/HTTP to Tempo. Initialization is best-effort:
  a failure logs a warning and the server continues with tracing off.
- **Logs** — structured logs through `logger.From(ctx)` only, with per-destination
  encodings (`LOG_CONSOLE_FORMAT` / `LOG_FILE_FORMAT`); the JSON file sink is tailed
  by Grafana Alloy into Loki.
- Label cardinality is bounded by `internal/infrastructure/httproute` `Classifier`,
  which maps a concrete path back to its registered route template.

### Testing

`go test ./...`. Tests are stdlib-only with hand-rolled fakes, and exist across
several layers: infrastructure (`metrics`, `httproute`, `logger`, `tracing`,
`metadata`, the socket hub, one repository test), bootstrap middleware (metrics,
session token, WebSocket upgrade), application (`command/otp`, `command/user`,
`command/shared/seqgen`, `query/quiz`, `query/progress`), modules (`quiz` validator,
`socket` integration over a real `websocket.Dial`), the bot adapter, and
`libs/eino`. Most module services and repositories remain untested. There is no CI
workflow in the repository.
