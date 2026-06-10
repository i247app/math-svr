# FEATURE-SPEC — Class Learning Progress

> Teacher-facing dashboard endpoint for tracking classroom members' score
> trends over time. Read-only, two endpoints, no commands, no UoW.

Aggregate: `classroom` (lives inside `internal/module/classroom/`).
Data source: `ma_exercise_submissions`.

---

## 1. Story

> "As a teacher (classroom OWNER or CO_TEACHER), I want to see how my
> class is doing over time, in two views:
> 1. A **line chart** of avg / highest / lowest scores per day, week, or
>    month, so I can spot good and bad stretches.
> 2. A **per-student list** with each student's average, a sparkline,
>    and a coloured chip (`Tiến bộ tốt` / `Tiến bộ` / `Cần cố gắng`),
>    so I can see at a glance who needs help."

A student may call the same endpoints for their own profile only
(`filter_profile_id == self`).

UI reference: screenshot in conversation. The two endpoints between them
fill every section of that screen — the chart, the three summary cards,
and the per-student table.

---

## 2. Decisions (signed off)

| # | Decision |
|---|---|
| 1 | Trend window: **always last 5 calendar days**, independent of bucket |
| 2 | Comment thresholds: `NEED_TO_TRY` (avg<5 OR slope≤-0.05), `GOOD_PROGRESS` (avg≥7.5 AND slope≥0), `PROGRESS` (the rest, with ≥2 subs) |
| 3 | Trend = linear regression on daily-avg scores, slope in 10-point pts per day |
| 4 | Summary deltas = vs immediately-prior same-length period |
| 5 | `improving_count` includes both `PROGRESS` and `GOOD_PROGRESS` |
| 6 | `need_support_count` is only `NEED_TO_TRY` (NO_DATA / INSUFFICIENT excluded) |
| 7 | `total_students` denominator = ACTIVE STUDENT-role members only |
| 8 | Chart zero-fills empty buckets with `submission_count:0, avg_score:null` |
| 9 | Bucket labels = ISO strings (`2026-05-06`, `2026-05`, Monday-of-week as date) |
| 10 | Week = ISO week, Monday start |
| 11 | `tz` accepted in request body (default `+07:00`) |
| 12 | Validator rejects `to_dt - from_dt > 2 years` |
| 13 | Each score returned as both raw `score_percentage` (0–100) and 10-point (1 dp) |
| 14 | `trend_series` density = per-bucket matching the chart's bucket (last 5 buckets) |
| 15 | Students may call either endpoint when `filter_profile_id == callerProfileID` |
| 16 | OWNER + CO_TEACHER access via `requireManager` (status quo) |
| 17 | Chart endpoint is unpaginated |
| 18 | Student list always includes all members, even `NO_DATA` / `INSUFFICIENT` |

> **Slope vs sparkline divergence (intentional).** `slope` and `comment`
> are always over the last 5 calendar days. `trend_series` is the last 5
> buckets at the request's bucket size. For `bucket=DAY` they coincide;
> for `WEEK`/`MONTH` the sparkline shows wider history while the chip
> still tracks last-5-days behaviour.

---

## 3. Endpoints

### 3a. `POST /classrooms/progress/scores-over-time`

Powers the line chart only.

**Request**
```json
{
  "profile_id": 1001,
  "classroom_id": 5001,
  "bucket": "DAY",
  "from_dt": "2026-05-06 00:00:00",
  "to_dt":   "2026-05-10 23:59:59",
  "tz": "+07:00",
  "filter_profile_id": null
}
```

**Response** — every bucket in `[from_dt, to_dt]` present; empty buckets carry nulls (decision #8):
```json
{
  "mstatus": 200,
  "bucket": "DAY", "from_dt": "...", "to_dt": "...", "tz": "+07:00",
  "points": [
    {
      "bucket_label": "2026-05-06",
      "bucket_start": "2026-05-06 00:00:00",
      "submission_count": 12,
      "avg_score": 6.2,     "avg_score_pct":     62.4,
      "highest_score": 8.6, "highest_score_pct": 86.0,
      "lowest_score": 3.8,  "lowest_score_pct":  38.0
    },
    {
      "bucket_label": "2026-05-07", "bucket_start": "2026-05-07 00:00:00",
      "submission_count": 0,
      "avg_score": null,     "avg_score_pct": null,
      "highest_score": null, "highest_score_pct": null,
      "lowest_score": null,  "lowest_score_pct": null
    }
  ]
}
```

### 3b. `POST /classrooms/progress/students`

Powers the three summary cards + the per-student table.

**Request**
```json
{
  "profile_id": 1001,
  "classroom_id": 5001,
  "bucket": "DAY",
  "from_dt": "2026-05-01 00:00:00",
  "to_dt":   "2026-05-10 23:59:59",
  "tz": "+07:00",
  "filter_profile_id": null,
  "page": 1,
  "size": 20
}
```

**Response**
```json
{
  "mstatus": 200,
  "from_dt": "...", "to_dt": "...", "tz": "+07:00", "bucket": "DAY",
  "summary": {
    "total_students":           35,
    "participating_count":      35,
    "participation_rate":     1.00,
    "improving_count":          18,
    "improving_delta_pct":    0.51,
    "need_support_count":       5,
    "need_support_delta_pct": -0.02
  },
  "students": [
    {
      "profile_id":  7001,
      "profile_name":"Nguyễn Minh Anh",
      "avatar_url":  "https://...",
      "submission_count": 8,
      "avg_score": 8.2,     "avg_score_pct": 82.0,
      "highest_score": 9.4, "highest_score_pct": 94.0,
      "lowest_score": 6.8,  "lowest_score_pct": 68.0,
      "first_submitted_dt":"2026-05-01 09:00:00",
      "last_submitted_dt": "2026-05-10 14:00:00",
      "trend_series": [
        { "bucket_label":"2026-05-06","avg_score":7.8,"avg_score_pct":78.0 },
        { "bucket_label":"2026-05-07","avg_score":8.0,"avg_score_pct":80.0 },
        { "bucket_label":"2026-05-08","avg_score":8.4,"avg_score_pct":84.0 },
        { "bucket_label":"2026-05-09","avg_score":8.1,"avg_score_pct":81.0 },
        { "bucket_label":"2026-05-10","avg_score":8.6,"avg_score_pct":86.0 }
      ],
      "slope":   0.18,
      "comment": "GOOD_PROGRESS"
    },
    {
      "profile_id": 7099, "profile_name": "...", "avatar_url": "...",
      "submission_count": 0,
      "avg_score": null,    "avg_score_pct": null,
      "highest_score": null,"highest_score_pct": null,
      "lowest_score": null, "lowest_score_pct": null,
      "first_submitted_dt": null, "last_submitted_dt": null,
      "trend_series": [], "slope": null,
      "comment": "NO_DATA"
    }
  ],
  "pagination": { "page":1, "size":20, "total_count":35, "total_pages":2, "...": "..." }
}
```

Sort order for `students`: by comment priority
(`NEED_TO_TRY` → `NO_DATA` → `INSUFFICIENT` → `PROGRESS` → `GOOD_PROGRESS`),
tie-break by `profile_name` ASC. Surfaces students needing attention first.

---

## 4. Comment enum

Lives in `internal/shared/enum/classroom_progress.go`.

| Symbol | UI label (VN) | Chip color | Rule |
|---|---|---|---|
| `NO_DATA` | (no chip) | — | `submission_count == 0` |
| `INSUFFICIENT` | (no chip) | — | `submission_count == 1` |
| `NEED_TO_TRY` | Cần cố gắng | orange | `avg_score < 5.0` OR `slope ≤ -0.05` |
| `GOOD_PROGRESS` | Tiến bộ tốt | green | `avg_score ≥ 7.5` AND `slope ≥ 0` |
| `PROGRESS` | Tiến bộ | green | everything else with `submission_count ≥ 2` |

First-match-wins evaluation order: `NO_DATA` → `INSUFFICIENT` → `NEED_TO_TRY` → `GOOD_PROGRESS` → `PROGRESS`.

---

## 5. Permission rule

Both endpoints, single helper `requireProgressAccess`:

```
uid     := sessionUID(r)
caller  := resolveActingProfile(ctx, req.ProfileID, uid)
if req.FilterProfileID != nil && *req.FilterProfileID == caller.ProfileId():
    requireMember(ctx, req.ClassroomID, caller.ProfileId())   // ACTIVE member, any role
else:
    requireManager(ctx, req.ClassroomID, caller.ProfileId())  // OWNER or CO_TEACHER
```

Rejection codes: `UNAUTHORIZED`, `PROFILE_NOT_FOUND`,
`CLASSROOM_NOT_FOUND`, `CLASSROOM_PERMISSION_DENIED`.

---

## 6. Status codes (new — 12200 block)

| Code | Symbol | EN | VN |
|---|---|---|---|
| 12225 | `CLASSROOM_PROGRESS_INVALID_BUCKET` | "Invalid bucket." | "Đơn vị thời gian không hợp lệ." |
| 12226 | `CLASSROOM_PROGRESS_INVALID_DATE_RANGE` | "Invalid date range." | "Khoảng thời gian không hợp lệ." |
| 12227 | `CLASSROOM_PROGRESS_INVALID_TZ` | "Invalid time zone." | "Múi giờ không hợp lệ." |

Lockstep across `code.go`, `message_en.go`, `message_vn.go`.

Re-used: `UNAUTHORIZED`, `PROFILE_NOT_FOUND`, `CLASSROOM_NOT_FOUND`,
`CLASSROOM_PERMISSION_DENIED`, `FAIL`.

---

## 7. Data source

Table: `ma_exercise_submissions`
([`migrations/020_ma_exercise_submissions_table.sql`](../../../migrations/020_ma_exercise_submissions_table.sql)).

Active predicate (both endpoints):
```
status = 'ACTIVE'
AND deleted_dt IS NULL
AND (submission_status IS NULL OR submission_status != 'DELETED')
AND submitted_dt IS NOT NULL
AND score_percentage IS NOT NULL
AND classroom_id = ?
AND submitted_dt BETWEEN ? AND ?
```

New indexes (migration 021):
```sql
ALTER TABLE ma_exercise_submissions
  ADD KEY ix_classroom_submitted (classroom_id, submitted_dt),
  ADD KEY ix_classroom_profile_submitted (classroom_id, profile_id, submitted_dt);
```

---

## 8. Out of scope (v1)

- No pre-aggregation tables; queries hit the raw submission table directly.
- No background job to keep counts warm.
- No translation table for the comment enum — the client owns label localization.
- No CSV/export endpoint (screenshot's top-right ↓ icon is presumably
  a client-side export of the rendered data).
- No student self-view endpoint at `/profiles/me/progress` —
  students reach the teacher endpoints with `filter_profile_id == self`.

---

## 9. Refinements — 2026-06-10

Applied after the v1 implementation shipped. These supersede the
matching v1 decisions where they conflict.

### R1 — `avg_score` is now 2-dp

`avg_score = round(avg_score_pct / 10, 2)` everywhere — chart buckets
and per-student rows. `pct = 87.5 → 8.75`. Single source of truth:
`PctTo10Pt` in `internal/application/query/classroomprogress/trend.go`.
No conflict with the quiz module (which doesn't use this helper).

Supersedes the previous 1-dp rounding implied by decision #13.

### R2 — Per-student row carries **two parallel sparkline series**

The old `trend_series` field is gone; in its place:

- **`trend_series_by_day`** — last 5 calendar days, daily averages.
  Always daily density (the chart's `bucket` request field no longer
  affects this — supersedes decision #14).
- **`trend_series_by_exercise`** — last 5 graded submissions, one
  point per submission. Each point is
  `{ bucket_label, exercise_id, avg_score, avg_score_pct }` where
  `bucket_label` is the exercise title (fallback `#<exercise_id>`
  when the title row can't be resolved). `ExerciseID` is the stable
  join key.

Both series are returned unconditionally. No new request toggle.

### R3 — `comment_source` flagged on the response

Top-level `comment_source` string on `StudentsProgressRes`. Values
match `enum.TrendSource`:
- `"LAST_5_DAYS"` — comment chip + slope derived from the day window
  (decision #1, screenshot footer "Xu hướng hiển thị dựa trên 5 ngày
  gần nhất"). **Current default and only value emitted today.**
- `"LAST_5_EXERCISES"` — reserved for a future flip; would require
  re-calibrating the Classify thresholds (currently in pts-per-day).

Open question: should the boss sign off on switching to
per-attempt classification? The screenshot footer locks day-based
today; flagged but not changed.

### Touch list for the refinement

```
internal/shared/enum/classroom_progress.go                                     [+TrendSource enum]
internal/domain/exercise/repository.go                                         [+ClassroomExerciseID on ProgressRow]
internal/infrastructure/persistence/mysql/repositories/exercise_submission_repository.go
                                                                               [SELECT +classroom_exercise_id]
internal/application/dto/classroomprogress/progress_dto.go                     [TrendSeries → TrendSeriesByDay,
                                                                                +TrendSeriesByExercise,
                                                                                +ExerciseTrendPoint,
                                                                                +CommentSource on Res]
internal/application/query/classroomprogress/trend.go                          [PctTo10Pt → 2 dp,
                                                                                +LastNExerciseRows]
internal/application/query/classroomprogress/students_progress_query.go        [+exerciseRepo,
                                                                                +loadExerciseTitlesForLast5,
                                                                                buildStudentRow rewritten]
internal/module/classroom/service.go                                           [+exerciseRepo constructor arg]
internal/bootstrap/container/services.go                                       [+repos.ExerciseRepository thread-through]
```
