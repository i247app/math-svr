# IMPLEMENTATION-PLAN — Profile Learning Progress

> Spec: [`FEATURE-SPEC.md`](./FEATURE-SPEC.md).
> Bottom-up build order: migration/verify → status+enum → repo →
> application → module → wiring. Each phase ends with `make build` +
> `go vet ./...` (the Stop hook runs vet automatically). Stop for review
> at every **🛑 LAYER BOUNDARY** marker.

Legend: `[ ]` todo · `[~]` in progress · `[x]` done. Paths repo-relative.

> **Status (implemented 2026-06-26).** Phases 2–6 done: code + routes +
> index migration written; `go build ./...`, `go vet ./...`, and
> `go test -short ./...` clean across all feature packages. Remaining:
> Phase 1.1 (confirm `ma_exercise_submissions` table on each env) and
> Phase 7 (manual smoke against a populated DB) — both need a live MySQL.
> The `migrations/020_*.sql` index file is written but must be applied
> manually (boot-time Migrate is disabled).

Pure **read** feature: no UoW, no commands, no `seq`, no
`transaction.Repositories` change. The classroom `NewService` already
receives `submissionRepo` + `exerciseRepo` (currently unused args) and the
container already passes them — so bootstrap wiring is minimal.

---

## Phase 1 — Data dependency check + index

🛑 LAYER BOUNDARY (DB). `Edit(migrations/**)` is `ask`-gated. Boot-time
`database.Migrate` is commented out (known-issues §5) — any new migration
is applied manually.

- [ ] **1.1** Confirm `ma_exercise_submissions` exists on dev/prod with the
      columns the repo scans (`classroom_id`, `profile_id`,
      `classroom_exercise_id`, `score_percentage`, `correct_number`,
      `total_questions`, `submitted_dt`, `submission_status`, `status`,
      `deleted_dt`). The table has **no migration file** in `/migrations/`
      on this branch — find/author the canonical table migration if it is
      genuinely missing, OR document that it is applied out-of-band. Do not
      proceed to Phase 3 until the table is confirmed.
- [ ] **1.2** Add `migrations/020_ma_exercise_submissions_progress_indexes.sql`
      (next free number; adjust if a submissions-table migration is added
      first):
      ```sql
      -- migration up
      -- Backs /classrooms/progress/profile: the per-student range scan
      -- ordered by time. Safe to skip if an equivalent index already exists.
      ALTER TABLE ma_exercise_submissions
        ADD KEY ix_classroom_profile_submitted (classroom_id, profile_id, submitted_dt);
      ```
- [ ] **1.3** `make build` (sanity compile; no schema parser).

---

## Phase 2 — Status codes + enum

No layer crossing — pure constants.

- [ ] **2.1** `internal/domain/shared/status/code.go` — append after
      `CLASSROOM_PROGRAM_DUPLICATE StatusCode = 12224`:
      ```go
      CLASSROOM_PROGRESS_INVALID_DATE_RANGE StatusCode = 12225
      CLASSROOM_PROGRESS_INVALID_TZ         StatusCode = 12226
      CLASSROOM_PROGRESS_INVALID_PURPOSE    StatusCode = 12227
      ```
- [ ] **2.2** `internal/domain/shared/status/message_en.go` — add:
      `"Invalid date range."`, `"Invalid time zone."`,
      `"Invalid exercise type."` under the same keys.
- [ ] **2.3** `internal/domain/shared/status/message_vn.go` — add:
      `"Khoảng thời gian không hợp lệ."`, `"Múi giờ không hợp lệ."`,
      `"Loại bài không hợp lệ."`
- [ ] **2.4** Recreate `internal/shared/enum/classroom_progress.go`
      (recover from git `6144d38`, then trim) keeping only:
      - `ProgressComment` + its consts + `IsValid()` / `String()`
      - `DefaultProgressTz = "+07:00"`
      - `tzOffsetPattern` / `tzOffsetRegex` / `IsValidTzOffset(s)`
      Drop `ProgressBucket` and `TrendSource` (this endpoint uses neither).
- [ ] **2.5** `make build` + `go vet ./...`.

---

## Phase 3 — Repo layer

🛑 LAYER BOUNDARY (domain → infra).

- [ ] **3.1** `internal/domain/exercise/repository.go` — add ONE method to
      `ISubmissionRepository`:
      ```go
      // ListProfileSubmissionsInRange returns one student's graded,
      // active submissions in a classroom within [from, to], ordered by
      // submitted_dt ASC. score_percentage IS NOT NULL is enforced so
      // every row is chartable. No JOIN — exercise title/purpose are
      // hydrated by the caller via exercise.IRepository.
      ListProfileSubmissionsInRange(ctx context.Context, params ProfileSubmissionsRangeParams) ([]*Submission, error)
      ```
      Add the param type in the same file:
      ```go
      type ProfileSubmissionsRangeParams struct {
          ClassroomID int64
          ProfileID   int64
          From, To    mtime.MathTime
      }
      ```
      Leave the leftover `BucketedScoresParams` / `ProgressRow` etc. types
      as-is (or delete them if `go vet`/unused-lint complains — they are
      `001` residue and unused after this feature).
- [ ] **3.2** `internal/infrastructure/persistence/mysql/repositories/exercise_submission_repository.go`
      — implement, mirroring `ListRecentByProfileIds` (no pagination):
      ```sql
      SELECT <exerciseSubmissionColumns>
      FROM   ma_exercise_submissions s
      WHERE  <exerciseSubmissionActiveWhere>
        AND  s.classroom_id     = ?
        AND  s.profile_id       = ?
        AND  s.submitted_dt BETWEEN ? AND ?
        AND  s.submitted_dt     IS NOT NULL
        AND  s.score_percentage IS NOT NULL
      ORDER BY s.submitted_dt ASC
      ```
      Reuse `exerciseSubmissionActiveArgs()`; append `classroomID,
      profileID, from, to`. Scan via the existing `scanSubmission` helper.
      Wrap errors: `fmt.Errorf("exercise submission repo list-profile-range: %w", err)`.
      Return `(nil, nil)` on no rows (never `sql.ErrNoRows`).
- [ ] **3.3** `make build` + `go vet ./...`.

---

## Phase 4 — Application layer

🛑 LAYER BOUNDARY (infra → application). New DTO + new query package.

### 4a. DTO

- [ ] **4.1** Create `internal/application/dto/classroomprogress/profile_progress_dto.go`.
      Score fields nullable (`*float64`) so empty range serialises `null`.
      ```go
      type ProfileProgressReq struct {
          ProfileID       int64          `json:"profile_id"`
          ClassroomID     int64          `json:"classroom_id"`
          TargetProfileID *int64         `json:"target_profile_id"`
          FromDt, ToDt    mtime.MathTime `json:"from_dt" / "to_dt"`
          Purpose         *string        `json:"purpose"`
          Tz              string         `json:"tz"`
      }
      type ProfileProgressRes struct {
          TargetProfileID int64          `json:"target_profile_id"`
          FromDt, ToDt    mtime.MathTime `json:"from_dt"/"to_dt"`
          Tz              string         `json:"tz"`
          Purpose         *string        `json:"purpose"`
          Series          []ExercisePoint `json:"series"`
          Summary         ProgressSummary `json:"summary"`
      }
      type ExercisePoint struct {
          ExerciseID     int64          `json:"exercise_id"`
          Title          string         `json:"title"`
          SubmittedDt    mtime.MathTime `json:"submitted_dt"`
          Score          float64        `json:"score"`
          ScorePct       int64          `json:"score_pct"`
          CorrectNumber  *int64         `json:"correct_number"`
          TotalQuestions *int64         `json:"total_questions"`
          Passed         bool           `json:"passed"`
      }
      type ProgressSummary struct {
          TotalExercises       int64          `json:"total_exercises"`
          GradedCount          int64          `json:"graded_count"`
          AverageScore         *float64       `json:"average_score"`
          AverageScorePct      *float64       `json:"average_score_pct"`
          AverageDelta         *float64       `json:"average_delta"`
          HighestScore         *float64       `json:"highest_score"`
          HighestScorePct      *int64         `json:"highest_score_pct"`
          HighestExerciseID    *int64         `json:"highest_exercise_id"`
          HighestExerciseTitle *string        `json:"highest_exercise_title"`
          HighestSubmittedDt   mtime.MathTime `json:"highest_submitted_dt"`
          PassedCount          int64          `json:"passed_count"`
          CorrectRate          *float64       `json:"correct_rate"`
          Trend                string         `json:"trend"`
      }
      ```

### 4b. Query package — `classroomprogress`

- [ ] **4.2** Recover `internal/application/query/classroomprogress/trend.go`
      from git `6144d38` and trim to the three pure helpers this feature
      uses: `PctTo10Pt(pct float64) float64`,
      `LinearSlope(values []float64) float64`,
      `Classify(count int, avg10pt, slope float64) enum.ProgressComment`.
      Drop the bucket/day-window helpers (`BuildTrendSeries`,
      `BuildSlopeWindow`, `LastNExerciseRows`, `SlopeWindowDays`).
- [ ] **4.3** Create `internal/application/query/classroomprogress/profile_progress_query.go`:
      - `const PassScorePctThreshold int64 = 50`
      - Constructor `NewProfileProgressQueryHandler(submissionRepo exercise.ISubmissionRepository, exerciseRepo exercise.IRepository)`.
      - `Handle(ctx, q ProfileProgressQuery) (*ProfileProgressResult, error)`:
        1. `cur := submissionRepo.ListProfileSubmissionsInRange(classroom, target, from, to)`.
        2. Hydrate titles + purpose: collect `classroom_exercise_id`s →
           `exerciseRepo.ListByClassroomExerciseIds(ids)` → map by id.
        3. If `q.Purpose != nil`: drop submissions whose exercise purpose
           ≠ requested (exercise missing → treat as excluded only when a
           purpose filter is active; otherwise keep with `#id` title).
        4. Build `series` (sorted ASC already): `score = PctTo10Pt(pct)`,
           `passed = pct >= PassScorePctThreshold`, title fallback `#<id>`.
        5. Aggregate summary over the filtered set:
           - `graded_count = len(series)`, `total_exercises = graded_count`
           - `average_score_pct = mean(pct)`, `average_score = PctTo10Pt(avg_pct)`
           - highest by `pct` (tie-break: latest `submitted_dt`) → score/title/id/dt
           - `passed_count = #(passed)`, `correct_rate = passed/graded` (nil if graded==0)
           - `slope = LinearSlope([]float64 of series scores in 10pt)`
           - `trend = Classify(graded_count, average_score, slope)`
        6. Prior period: `dur = to - from`;
           `priorTo = from - 1µs`; `priorFrom = from - dur`.
           Second `ListProfileSubmissionsInRange(...)` for the prior window
           (apply the same purpose filter). `average_delta =`
           `average_score(cur) - average_score(prior)`, nil when either
           period has no graded submission.
      - Return a result struct the module maps to the DTO. No avatar / no
        I/O here (pure read; nothing to hydrate via storage for this view).

- [ ] **4.4** `make build` + `go vet ./...`.

---

## Phase 5 — Module layer

🛑 LAYER BOUNDARY (application → module/presentation).

- [ ] **5.1** `internal/module/classroom/errors.go` — add
      `ErrProgressInvalidPurpose = errors.New("purpose must be HOMEWORK, QUIZ, or EXAM")`.
      Reuse existing `ErrProgressInvalidDateRange` / `ErrProgressInvalidTz`.
- [ ] **5.2** Create `internal/module/classroom/progress_validator.go`:
      `ValidateProfileProgress(ctx, req) *MathError`:
      - `classroom_id != 0` → else `CLASSROOM_MISSING_ID`
      - `profile_id != 0` → else `PROFILE_NOT_FOUND`
      - `from_dt < to_dt` AND `to_dt - from_dt ≤ 2y` → else `CLASSROOM_PROGRESS_INVALID_DATE_RANGE`
      - `tz` empty → default `enum.DefaultProgressTz`; else must match
        `enum.IsValidTzOffset` → else `CLASSROOM_PROGRESS_INVALID_TZ`
      - `purpose` nil OK; else ∈ {HOMEWORK, QUIZ, EXAM} → else
        `CLASSROOM_PROGRESS_INVALID_PURPOSE` (use the exercise purpose enum
        / validator already used by `module/exercise` — reuse, don't
        hand-roll the whitelist).
- [ ] **5.3** Create `internal/module/classroom/progress_service.go`:
      - `requireProfileProgressAccess(ctx, classroomID int64, caller *profile.Profile, target int64) error`
        per `FEATURE-SPEC §5` (reuse `requireMember` / `requireManager`).
      - `GetProfileProgress(ctx, req, sessionUserID) (*dto.ProfileProgressRes, error)`:
        1. `ValidateProfileProgress`.
        2. `caller := s.resolveActingProfile(ctx, req.ProfileID, sessionUserID)`.
        3. `target := req.TargetProfileID or caller.ProfileId()`.
        4. `s.requireProfileProgressAccess(ctx, req.ClassroomID, caller, target)`.
        5. Load classroom (`classroomRepo.FindByClassroomId`) → nil →
           `CLASSROOM_NOT_FOUND`.
        6. Apply tz default; delegate to `profileProgressQuery.Handle(...)`.
        7. Map result → DTO; return.
- [ ] **5.4** `internal/module/classroom/service.go`:
      - Add struct fields `profileProgressQuery *progressQuery.ProfileProgressQueryHandler`,
        and store the already-passed `submissionRepo` / `exerciseRepo` if a
        field is needed (query handler is constructed in `NewService` from
        the two existing args — no signature change).
      - In `NewService`, construct
        `progressQuery.NewProfileProgressQueryHandler(submissionRepo, exerciseRepo)`
        and assign it. (These two params already exist on the constructor
        and were previously dropped.)
- [ ] **5.5** Create `internal/module/classroom/progress_handler.go`:
      ```go
      // POST /classrooms/progress/profile
      func (h *ClassroomHandler) HandleProfileProgress(w http.ResponseWriter, r *http.Request) { ... }
      ```
      JSON decode → `sessionUID` → `classroomSvc.GetProfileProgress` →
      `response.WriteJson(w, res, err)`. Mirror an existing classroom
      handler exactly.
- [ ] **5.6** `make build` + `go vet ./...`.

---

## Phase 6 — Routes

🛑 LAYER BOUNDARY (module → bootstrap). Container needs **no change** —
`NewService` signature is unchanged; the two repos were already wired.

- [ ] **6.1** `internal/bootstrap/routes/routes.go` — append inside the
      classroom block (after the join-requests lines, ~line 206):
      ```go
      gexSvr.AddRoute("POST /classrooms/progress/profile",
          classroomHandler.HandleProfileProgress, authMiddleware)
      ```
- [ ] **6.2** Final `make build` + `go vet ./...`.

---

## Phase 7 — Smoke verification (manual)

No automated tests in repo (`.claude/rules/testing.md`). Spot-check via
local server + curl/Postman against a populated DB.

- [ ] **7.1** `POST /classrooms/progress/profile` for a known
      (classroom, target) over a range that spans several graded
      submissions → `series` length = graded count, `submitted_dt` ASC,
      `score == score_pct/10` (2 dp), `passed` correct vs the 50% threshold.
- [ ] **7.2** Summary: `average_score == mean(score)`; `highest_*` matches
      the max-score row; `passed_count`/`correct_rate == passed/graded`;
      `average_delta` matches current-minus-prior; `trend` reasonable.
- [ ] **7.3** Purpose filter: pass `purpose:"EXAM"` → only EXAM exercises
      counted; invalid `purpose:"FOO"` → `CLASSROOM_PROGRESS_INVALID_PURPOSE`.
- [ ] **7.4** Auth: member with `target == self` → 200; member with
      `target == other` → `CLASSROOM_PERMISSION_DENIED`; OWNER/CO_TEACHER
      with `target == any student` → 200; no session → `UNAUTHORIZED`.
- [ ] **7.5** Validator: `to_dt < from_dt` and `to_dt - from_dt = 3y` →
      `CLASSROOM_PROGRESS_INVALID_DATE_RANGE`; `tz:"Asia/Ho_Chi_Minh"` →
      `CLASSROOM_PROGRESS_INVALID_TZ`.
- [ ] **7.6** `EXPLAIN` the range query; confirm
      `ix_classroom_profile_submitted` is used.

---

## Touch list — recap

```
migrations/020_ma_exercise_submissions_progress_indexes.sql                    [new, ask-gated]
internal/domain/shared/status/code.go                                          [+3 codes]
internal/domain/shared/status/message_en.go                                    [+3 messages]
internal/domain/shared/status/message_vn.go                                    [+3 messages]
internal/shared/enum/classroom_progress.go                                     [new — recovered subset]
internal/domain/exercise/repository.go                                         [+1 method, +1 param type]
internal/infrastructure/persistence/mysql/repositories/exercise_submission_repository.go  [+1 impl]
internal/application/dto/classroomprogress/profile_progress_dto.go             [new]
internal/application/query/classroomprogress/trend.go                          [new — recovered subset]
internal/application/query/classroomprogress/profile_progress_query.go         [new]
internal/module/classroom/errors.go                                            [+1 error var]
internal/module/classroom/progress_validator.go                                [new]
internal/module/classroom/progress_service.go                                  [new]
internal/module/classroom/progress_handler.go                                  [new]
internal/module/classroom/service.go                                           [+1 field, +1 constructor body line]
internal/bootstrap/routes/routes.go                                            [+1 route]
```

Untouched: `application/transaction/Repositories`, `unit_of_work.go`,
`domain/seq/names.go`, `application/command/*`, `bootstrap/container/*`.
Pure read feature; constructor signature unchanged.
```
