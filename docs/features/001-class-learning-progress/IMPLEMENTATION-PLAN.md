# IMPLEMENTATION-PLAN — Class Learning Progress

> Spec: [`FEATURE-SPEC.md`](./FEATURE-SPEC.md).
> Phases are bottom-up: migration → status/enum → repo → application →
> module → wiring. Each phase ends with `make build` + `go vet ./...`
> (the project Stop hook runs vet automatically). Stop and let the user
> review at every **🛑 LAYER BOUNDARY** marker.

Legend:
- `[ ]` = todo, `[~]` = in progress, `[x]` = done.
- File paths are repo-relative.

---

## Phase 1 — Migration + indexes

🛑 LAYER BOUNDARY (DB). `Edit(migrations/**)` is `ask`-gated; the user
must approve. Also: per `.claude/rules/known-issues.md` §5, the boot-time
`database.Migrate(...)` call is currently commented out — this file
**must be applied manually** on every environment after merge.

- [x] **1.1** Create `migrations/021_ma_exercise_submissions_progress_indexes.sql`:
  ```sql
  -- migration up
  -- Indexes that back the class-learning-progress endpoints.
  -- ix_classroom_submitted        — covers the GROUP BY bucket aggregation
  --                                 in /classrooms/progress/scores-over-time.
  -- ix_classroom_profile_submitted — covers the per-student fetch ordered
  --                                 by time in /classrooms/progress/students,
  --                                 including the 5-day trend-window slice.
  ALTER TABLE ma_exercise_submissions
    ADD KEY ix_classroom_submitted (classroom_id, submitted_dt),
    ADD KEY ix_classroom_profile_submitted (classroom_id, profile_id, submitted_dt);
  ```
- [ ] **1.2** Run `EXPLAIN` on the two canonical queries (from §Phase 4
      below) against a non-empty DB; confirm the planner picks the new
      indexes (`Using index condition` on `ix_classroom_submitted` for
      the chart, `ix_classroom_profile_submitted` for the student query).
      **Deferred — needs a live MySQL with data; run manually after
      applying the migration to dev.**
- [x] **1.3** Verify `make build` still passes (no schema parser, but
      sanity check the repo compiles). — `go mod tidy && go build ./cmd/mathsvr` clean.

---

## Phase 2 — Status codes + enum

No layer crossing — pure constants. Single PR-able commit.

- [x] **2.1** `internal/domain/shared/status/code.go` — append after
      `CLASSROOM_PROGRAM_DUPLICATE StatusCode = 12224`:
  ```go
  CLASSROOM_PROGRESS_INVALID_BUCKET      StatusCode = 12225
  CLASSROOM_PROGRESS_INVALID_DATE_RANGE  StatusCode = 12226
  CLASSROOM_PROGRESS_INVALID_TZ          StatusCode = 12227
  ```
- [x] **2.2** `internal/domain/shared/status/message_en.go` — add three
      lines under the same key constants:
      `"Invalid bucket."`, `"Invalid date range."`, `"Invalid time zone."`
- [x] **2.3** `internal/domain/shared/status/message_vn.go` —
      `"Đơn vị thời gian không hợp lệ."`, `"Khoảng thời gian không hợp lệ."`,
      `"Múi giờ không hợp lệ."`
- [x] **2.4** Create `internal/shared/enum/classroom_progress.go`:
  ```go
  package enum

  // ProgressBucket selects the chart's x-axis granularity.
  type ProgressBucket string

  const (
      ProgressBucketDay   ProgressBucket = "DAY"
      ProgressBucketWeek  ProgressBucket = "WEEK"
      ProgressBucketMonth ProgressBucket = "MONTH"
  )

  func IsValidProgressBucket(s string) bool { /* DAY|WEEK|MONTH */ }

  // ProgressComment classifies a single student's recent trend.
  // See FEATURE-SPEC §4 for the rule table and evaluation order.
  type ProgressComment string

  const (
      ProgressCommentNoData       ProgressComment = "NO_DATA"
      ProgressCommentInsufficient ProgressComment = "INSUFFICIENT"
      ProgressCommentNeedToTry    ProgressComment = "NEED_TO_TRY"
      ProgressCommentProgress     ProgressComment = "PROGRESS"
      ProgressCommentGoodProgress ProgressComment = "GOOD_PROGRESS"
  )

  // TzOffsetRegex validates request tz strings ("+07:00", "-05:30").
  // Validator-only; the value is passed verbatim into MySQL CONVERT_TZ.
  const TzOffsetPattern = `^[+-]\d{2}:\d{2}$`

  const DefaultProgressTz = "+07:00"
  ```
- [x] **2.5** `make build` + `go vet ./...`. — build clean; vet clean on the
      files Phase 2 touched (pre-existing format-arg warnings in
      `classroom/service_helper.go`, `grade/service.go`, `profile/service.go`
      are unrelated — flag and fix later or spawn a cleanup task).

---

## Phase 3 — Repo layer (domain interface + MySQL impl)

🛑 LAYER BOUNDARY (domain → infra). Add the read methods to
`ISubmissionRepository` first, then implement.

- [x] **3.1** `internal/domain/exercise/repository.go` — extend
      `ISubmissionRepository` with two methods (with doc comments):

  ```go
  // ListBucketedScores returns the chart-shaped aggregation: one row
  // per (bucket_label) inside [from, to], with count + sum/min/max of
  // score_percentage. tzOffset is a MySQL CONVERT_TZ-friendly string
  // like "+07:00". bucket selects DAY|WEEK|MONTH. filterProfileID is
  // optional. Caller is responsible for zero-filling missing buckets.
  ListBucketedScores(ctx context.Context, params BucketedScoresParams) ([]*BucketedScoreRow, error)

  // ListForProgress streams the (profile_id, score_percentage,
  // submitted_dt) tuples for a classroom inside [from, to], ordered by
  // (profile_id, submitted_dt ASC). One round trip; in-memory grouping
  // by the caller. filterProfileID is optional (one student or all).
  ListForProgress(ctx context.Context, params ProgressRangeParams) ([]*ProgressRow, error)
  ```

  Add the three small types in the same file:
  ```go
  type BucketedScoresParams struct {
      ClassroomID     int64
      Bucket          string // DAY|WEEK|MONTH (validator-normalized)
      TzOffset        string // "+07:00"
      From, To        mtime.MathTime
      FilterProfileID *int64
  }
  type BucketedScoreRow struct {
      BucketLabel     string // YYYY-MM-DD or YYYY-MM
      SubmissionCount int64
      AvgPct          float64
      MinPct          int64
      MaxPct          int64
  }
  type ProgressRangeParams struct {
      ClassroomID     int64
      From, To        mtime.MathTime
      FilterProfileID *int64
  }
  type ProgressRow struct {
      ProfileID       int64
      ScorePercentage int64
      SubmittedDt     mtime.MathTime
  }
  ```

- [x] **3.2** `internal/infrastructure/persistence/mysql/repositories/exercise_submission_repository.go`
      — append both methods. SQL templates:

  Bucket-label expression dispatch (build per `params.Bucket`):
  ```go
  switch params.Bucket {
  case "DAY":
      labelExpr = `DATE_FORMAT(CONVERT_TZ(s.submitted_dt,'+00:00',?),'%Y-%m-%d')`
  case "WEEK":
      labelExpr = `DATE_FORMAT(DATE_SUB(CONVERT_TZ(s.submitted_dt,'+00:00',?),
                                INTERVAL WEEKDAY(CONVERT_TZ(s.submitted_dt,'+00:00',?)) DAY),
                                '%Y-%m-%d')`
  case "MONTH":
      labelExpr = `DATE_FORMAT(CONVERT_TZ(s.submitted_dt,'+00:00',?),'%Y-%m')`
  }
  ```
  The number of `?` placeholders for tz varies by bucket (1 for DAY/MONTH,
  2 for WEEK) — bind `params.TzOffset` once per placeholder.

  `ListBucketedScores` SQL skeleton:
  ```sql
  SELECT <labelExpr> AS bucket_label,
         COUNT(*)                  AS n,
         AVG(s.score_percentage)   AS avg_pct,
         MIN(s.score_percentage)   AS min_pct,
         MAX(s.score_percentage)   AS max_pct
  FROM ma_exercise_submissions s
  WHERE <activeWhere>
    AND s.classroom_id = ?
    AND s.submitted_dt BETWEEN ? AND ?
    AND s.score_percentage IS NOT NULL
    [AND s.profile_id = ?]   -- when filterProfileID set
  GROUP BY bucket_label
  ORDER BY bucket_label ASC;
  ```

  `ListForProgress` SQL skeleton:
  ```sql
  SELECT s.profile_id, s.score_percentage, s.submitted_dt
  FROM ma_exercise_submissions s
  WHERE <activeWhere>
    AND s.classroom_id = ?
    AND s.submitted_dt BETWEEN ? AND ?
    AND s.score_percentage IS NOT NULL
    [AND s.profile_id = ?]
  ORDER BY s.profile_id ASC, s.submitted_dt ASC;
  ```

  Both reuse `exerciseSubmissionActiveArgs()` for the active predicate.
  Wrap errors with `fmt.Errorf("exercise submission repo list-X: %w", err)`.
  Return `(nil, nil)`-style empties (no `sql.ErrNoRows`).

- [x] **3.3** `make build` + `go vet ./...`. — build clean; vet clean on
      `internal/domain/exercise/...` and
      `internal/infrastructure/persistence/mysql/repositories/...`.
      A `rangeint` modernization suggestion on the new code was applied.

---

## Phase 4 — Application layer

🛑 LAYER BOUNDARY (infra → application). New query package; new DTO
package. No UoW, no commands.

### 4a. DTO package

- [x] **4.1** Create `internal/application/dto/classroomprogress/progress_dto.go`
      with the request/response shapes from `FEATURE-SPEC §3`. Tag every
      JSON field. Score fields use `*float64` / `*int64` so empty
      buckets can serialise as `null` per decision #8.

  Key types (names only — implement per spec):
  ```go
  type ScoresOverTimeReq struct { ProfileID, ClassroomID int64; Bucket string;
      FromDt, ToDt mtime.MathTime; Tz string; FilterProfileID *int64 }
  type ScoresOverTimeRes struct { Bucket, Tz string; FromDt, ToDt mtime.MathTime;
      Points []ChartPoint }
  type ChartPoint struct { BucketLabel string; BucketStart mtime.MathTime;
      SubmissionCount int64;
      AvgScore, HighestScore, LowestScore *float64;
      AvgScorePct, HighestScorePct, LowestScorePct *float64 }

  type StudentsProgressReq struct { ProfileID, ClassroomID int64; Bucket string;
      FromDt, ToDt mtime.MathTime; Tz string;
      FilterProfileID *int64; Page, Size int64 }
  type StudentsProgressRes struct { Bucket, Tz string; FromDt, ToDt mtime.MathTime;
      Summary ProgressSummary; Students []StudentProgressRow;
      Pagination *pagination.Pagination }
  type ProgressSummary struct {
      TotalStudents, ParticipatingCount int64
      ParticipationRate                 float64
      ImprovingCount                    int64
      ImprovingDeltaPct                 *float64
      NeedSupportCount                  int64
      NeedSupportDeltaPct               *float64
  }
  type StudentProgressRow struct {
      ProfileID int64; ProfileName string; AvatarUrl *string
      SubmissionCount int64
      AvgScore, HighestScore, LowestScore         *float64
      AvgScorePct, HighestScorePct, LowestScorePct *float64
      FirstSubmittedDt, LastSubmittedDt mtime.MathTime
      TrendSeries []TrendPoint
      Slope       *float64
      Comment     string // enum.ProgressComment as string
  }
  type TrendPoint struct { BucketLabel string; AvgScore, AvgScorePct *float64 }
  ```

### 4b. Query package — `classroomprogress`

- [x] **4.2** Create `internal/application/query/classroomprogress/bucket.go`:
  - `BucketLabel(t time.Time, bucket, tz string) string` — format any
    timestamp into its bucket label, mirroring the SQL expressions in
    §3.2. Used for the trend-series grouping in Go.
  - `BucketSequence(from, to time.Time, bucket, tz string) []BucketKey` —
    walks the bucket boundaries between `from` and `to` for zero-fill.
    Each `BucketKey` carries `{Label, Start}` so the chart point can
    populate `bucket_start`.

- [x] **4.3** Create `internal/application/query/classroomprogress/trend.go`:
  - `LinearSlope(points []float64) float64` — simple linear regression
    of y on integer x = 0..N-1. Returns 0 when N < 2.
  - `Classify(submissionCount int, avgScore10pt, slope float64) enum.ProgressComment`
    — applies the first-match-wins table from `FEATURE-SPEC §4`.
  - `BuildTrendSeries(rows []exercise.ProgressRow, bucket, tz string, to time.Time, n int) []dto.TrendPoint`
    — buckets the rows per `bucket` in `tz`, returns the **last `n`
    buckets** with submissions (n=5 in §3b's spec). The boundary is
    "buckets ≤ `to`."
  - `BuildSlopeWindow(rows []exercise.ProgressRow, tz string, to time.Time) []float64`
    — buckets the rows into **daily** averages for the **last 5
    calendar days** in `tz`, regardless of the chart's bucket. Returns
    the 0..5 daily means (omitting days with no submissions).

- [x] **4.4** Create `internal/application/query/classroomprogress/scores_over_time_query.go`:
  - `ScoresOverTimeQuery` struct = the inputs (mirrors the DTO request,
    minus `ProfileID` which is consumed in the module layer).
  - `Handle(ctx, q) (*ScoresOverTimeResult, error)`:
    1. Call `submissionRepo.ListBucketedScores(...)`.
    2. Walk `BucketSequence(from, to, bucket, tz)`; index DB rows by
       `bucket_label`; emit a point per sequence entry (zero-filled if
       missing).
    3. Convert `score_percentage` → 10-point via `pct / 10.0` rounded
       to 1 dp (helper `score.ToTenPoint`).

- [x] **4.5** Create `internal/application/query/classroomprogress/students_progress_query.go`:
  - `StudentsProgressQuery` struct = inputs from the DTO request.
  - `Handle(ctx, q) (*StudentsProgressResult, error)`:
    1. Load STUDENT members via
       `memberRepo.ListMembers(ClassroomId=q.ClassroomID, Role=STUDENT,
       Status=ACTIVE, take_all)`. Drives `total_students` (decision #7)
       and `students` denominator (decision #18 — include everyone).
    2. Compute prior period:
       `priorTo = q.FromDt - 1µs; priorFrom = q.FromDt - (q.ToDt - q.FromDt)`.
    3. Call `submissionRepo.ListForProgress` for **current** range.
    4. Call `submissionRepo.ListForProgress` for **prior** range.
    5. Bucket both result sets by `profile_id` in Go.
    6. For each STUDENT member, compute:
       - aggregates (count/avg/min/max/first/last) over current slice
       - `trend_series` via `BuildTrendSeries(rows, q.Bucket, q.Tz, q.ToDt, 5)`
       - `slopeWindow := BuildSlopeWindow(rows, q.Tz, q.ToDt)`;
         `slope := LinearSlope(slopeWindow_10pt)` where the window
         values are converted to 10-point first
       - `comment := Classify(count, avg10pt, slope)`
       - **Prior-period comment** for delta accounting: same algorithm
         using the prior slice + `priorTo` as the window end.
    7. Build `summary`:
       - `TotalStudents = len(students)`
       - `ParticipatingCount = #(current count ≥ 1)`
       - `ParticipationRate = participating / total` (or 0 when total = 0)
       - `ImprovingCount = #(current comment ∈ {PROGRESS, GOOD_PROGRESS})`
       - `NeedSupportCount = #(current comment == NEED_TO_TRY)`
       - Prior counts the same way.
       - `ImprovingDeltaPct = (cur - prior) / max(prior, 1)` if `prior > 0`,
         else `nil`. Same for need-support.
    8. Apply `filter_profile_id` to the `students` list.
    9. Sort `students` by comment priority (per `FEATURE-SPEC §3b`),
       then by name.
    10. Paginate via `pagination.NewPagination(page, size, len(students))`
        + manual slicing (the slice is already in memory; do not re-query).
  - **Avatar hydration** is intentionally NOT done here — it's I/O
    against the storage adapter, which lives in the module layer. The
    query returns `avatar_key`; the module turns it into `avatar_url`.

- [x] **4.6** `make build` + `go vet ./...`. — `go build ./...` clean;
      targeted vet clean on both new packages. Applied `min`/`max`
      builtin modernizations and hoisted `studentResult` to package
      scope (named slice ≠ anonymous-struct slice for assignability).

---

## Phase 5 — Module layer

🛑 LAYER BOUNDARY (application → module/presentation).

- [x] **5.1** Create `internal/module/classroom/progress_validator.go`:
  - `ValidateScoresOverTime(ctx, req)` — checks:
    - `bucket ∈ {DAY, WEEK, MONTH}` → `CLASSROOM_PROGRESS_INVALID_BUCKET`
    - `from_dt < to_dt`, `to_dt - from_dt ≤ 2 years` →
      `CLASSROOM_PROGRESS_INVALID_DATE_RANGE`
    - `tz` defaults to `enum.DefaultProgressTz` when empty; otherwise
      must match `enum.TzOffsetPattern` → `CLASSROOM_PROGRESS_INVALID_TZ`
    - `profile_id != 0` → `PROFILE_NOT_FOUND`
    - `classroom_id != 0` → `CLASSROOM_MISSING_ID`
  - `ValidateStudentsProgress(ctx, req)` — same checks + `size` clamp
    via `pagination.DefaultPageSize` (no error, just a clamp), `page ≥ 1`.

- [x] **5.2** Create `internal/module/classroom/progress_service.go`:
  - New method on `*Service`:
    ```go
    func (s *Service) requireProgressAccess(
        ctx context.Context, classroomID int64,
        caller *profileDomain.Profile, filterProfileID *int64,
    ) error
    ```
    Logic per `FEATURE-SPEC §5`. Reuses existing `requireMember` and
    `requireManager` helpers from `permission.go`.
  - `GetScoresOverTime(ctx, req, sessionUserID) (*dto.ScoresOverTimeRes, error)`:
    1. `ValidateScoresOverTime`.
    2. `caller := s.resolveActingProfile(...)`.
    3. `s.requireProgressAccess(...)`.
    4. Load classroom via `classroomRepo.FindByClassroomId` → if nil →
       `CLASSROOM_NOT_FOUND`. (Same shape as `GetClassroom`.)
    5. Apply the tz default (`enum.DefaultProgressTz` when empty).
    6. Delegate to `scoresOverTimeQuery.Handle(...)`.
    7. Map to DTO response. Return.
  - `GetStudentsProgress(ctx, req, sessionUserID) (*dto.StudentsProgressRes, error)`:
    same flow + **avatar URL hydration** after the query returns
    (use `s.storageProvider` like `populateCoverUrl` does for
    classrooms — extract a small helper `hydrateProgressAvatars`).
    **Hydration must use a TTL of `coverUrlTTL` (1h)** — match the
    classroom convention.

- [x] **5.3** Extend `*Service` constructor in `service.go` with two new
      fields and inject them in `NewService(...)`:
  ```go
  scoresOverTimeQuery    *progressQuery.ScoresOverTimeQueryHandler
  studentsProgressQuery  *progressQuery.StudentsProgressQueryHandler
  ```
  Constructor calls
  `progressQuery.NewScoresOverTimeQueryHandler(submissionRepo)` and
  `progressQuery.NewStudentsProgressQueryHandler(memberRepo, profileRepo, submissionRepo)`.
  This **adds two new repos to the constructor signature**
  (`submissionRepo exerciseDomain.ISubmissionRepository`). Update
  `bootstrap/container/services.go` accordingly (Phase 6).

- [x] **5.4** Create `internal/module/classroom/progress_handler.go`:
  - Add two methods to `*ClassroomHandler`. Same shape as the existing
    handlers:
    ```go
    // POST /classrooms/progress/scores-over-time
    func (h *ClassroomHandler) HandleProgressScoresOverTime(w, r) { ... }

    // POST /classrooms/progress/students
    func (h *ClassroomHandler) HandleProgressStudents(w, r) { ... }
    ```
    JSON-only decoding (no multipart for these). `sessionUID` →
    `classroomSvc.GetScoresOverTime` / `GetStudentsProgress` →
    `response.WriteJson`.

- [x] **5.5** `make build` + `go vet ./...`. — `go build ./...` clean.
      Package-mode vet on `module/classroom/`,
      `application/query/classroomprogress/`,
      `application/dto/classroomprogress/` clean apart from the
      pre-existing `service_helper.go:241` `%s` warning (chip
      `task_46e2425e`).

---

## Phase 6 — Bootstrap wiring + routes

🛑 LAYER BOUNDARY (module → bootstrap). Last layer.

- [x] **6.1** `internal/bootstrap/container/services.go` — the
      `classroom.NewService(...)` call needs the new `submissionRepo`
      argument. The container already exposes
      `repos.ExerciseSubmissionRepository`
      (`bootstrap/container/repositories.go:30`), so just thread it
      through. — **Pulled forward into Phase 5** so the build stays
      green at the layer boundary; Phase 6 then only registers the
      two new routes.

- [x] **6.2** `internal/bootstrap/routes/routes.go` — append two routes
      inside the existing classroom block (after the join-request lines):
  ```go
  gexSvr.AddRoute("POST /classrooms/progress/scores-over-time",
      classroomHandler.HandleProgressScoresOverTime, authMiddleware)
  gexSvr.AddRoute("POST /classrooms/progress/students",
      classroomHandler.HandleProgressStudents, authMiddleware)
  ```

- [x] **6.3** Final `make build` + `go vet ./...`. — `go build ./...`
      clean across the whole tree. Vet clean on every Phase 6 file;
      pre-existing format-arg warnings in grade/profile/classroom-
      service_helper modules are covered by chip `task_46e2425e`.

---

## Phase 7 — Smoke verification (manual)

No automated tests in the repo yet (`.claude/rules/testing.md`).
Spot-check via a local server + curl/Postman.

- [ ] **7.1** With a populated DB: hit
      `POST /classrooms/progress/scores-over-time` for a known
      classroom across a 7-day range, bucket=DAY. Verify the response
      contains 7 buckets (zero-filled where appropriate) and that
      `avg_score` matches `avg_score_pct / 10` to 1 dp.
- [ ] **7.2** Hit `POST /classrooms/progress/students` for the same
      classroom and range. Verify:
      - `summary.total_students` matches `SELECT COUNT(*) FROM
        ma_classroom_members WHERE classroom_id=? AND role='STUDENT'
        AND member_status='ACTIVE'`.
      - At least one student carries each comment value seen in the
        screenshot examples.
      - Sort order surfaces `NEED_TO_TRY` first.
- [ ] **7.3** Auth checks:
      - Call as a STUDENT with `filter_profile_id == self` → 200.
      - Call as a STUDENT with `filter_profile_id == other` →
        `CLASSROOM_PERMISSION_DENIED`.
      - Call without a session → `UNAUTHORIZED`.
- [ ] **7.4** Validator checks:
      - `bucket=YEAR` → `CLASSROOM_PROGRESS_INVALID_BUCKET`.
      - `to_dt < from_dt` → `CLASSROOM_PROGRESS_INVALID_DATE_RANGE`.
      - `to_dt - from_dt = 3 years` → `CLASSROOM_PROGRESS_INVALID_DATE_RANGE`.
      - `tz="Asia/Ho_Chi_Minh"` (IANA, not offset) → `CLASSROOM_PROGRESS_INVALID_TZ`.
- [ ] **7.5** `EXPLAIN` both queries one more time on the populated DB;
      confirm new indexes are used.

---

## Touch list — recap

```
migrations/021_ma_exercise_submissions_progress_indexes.sql                    [new]
internal/domain/shared/status/code.go                                          [+3 codes]
internal/domain/shared/status/message_en.go                                    [+3 messages]
internal/domain/shared/status/message_vn.go                                    [+3 messages]
internal/shared/enum/classroom_progress.go                                     [new]
internal/domain/exercise/repository.go                                         [+2 methods, +4 types on ISubmissionRepository]
internal/infrastructure/persistence/mysql/repositories/exercise_submission_repository.go
                                                                               [+2 method implementations]
internal/application/dto/classroomprogress/progress_dto.go                     [new]
internal/application/query/classroomprogress/bucket.go                         [new]
internal/application/query/classroomprogress/trend.go                          [new]
internal/application/query/classroomprogress/scores_over_time_query.go         [new]
internal/application/query/classroomprogress/students_progress_query.go        [new]
internal/module/classroom/progress_validator.go                                [new]
internal/module/classroom/progress_service.go                                  [new]
internal/module/classroom/progress_handler.go                                  [new]
internal/module/classroom/service.go                                           [+2 query handlers in *Service + constructor arg]
internal/bootstrap/container/services.go                                       [thread submissionRepo into classroom.NewService]
internal/bootstrap/routes/routes.go                                            [+2 routes]
```

Untouched: `application/transaction/Repositories`, `unit_of_work.go`,
`domain/seq/names.go`, `application/command/*`. Pure read feature.
