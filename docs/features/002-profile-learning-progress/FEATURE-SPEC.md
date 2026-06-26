# FEATURE-SPEC — Profile Learning Progress (single-student detail)

> Parent/student-facing endpoint showing ONE student's learning-progress
> detail inside ONE classroom: a per-exercise score chart, four summary
> cards, and a trend banner. Read-only, single endpoint, no commands, no
> UoW.

Aggregate: `classroom` (lives inside `internal/module/classroom/`).
Data source: `ma_exercise_submissions` (+ `ma_exercises` for titles/purpose).

UI reference: the "Tiến độ học tập" screen attached to the request — a
line chart "Điểm số theo bài đánh giá" with one point per assessment
(Bài 1…Bài 6), four cards (Điểm trung bình / Điểm cao nhất / Tỷ lệ làm
đúng / Xu hướng), and a banner ("Con đang tiến bộ tốt!").

This is distinct from the (removed) `001-class-learning-progress`
teacher feature, whose chart bucketed the whole class by DAY/WEEK/MONTH.
Here the x-axis is **per exercise** for a **single** profile.

---

## 1. Story

> "As a parent (or the student themself), I want to open one child's
> classroom progress and see, over a date range I pick:
> 1. A **line chart** with one point per graded assessment, so I can see
>    how each bài went over time.
> 2. Four **summary cards**: average score (with delta vs the previous
>    period), highest score (and which bài), pass rate (đạt/total), and a
>    trend label.
> 3. A **banner** that congratulates or nudges based on the trend."

A teacher (OWNER / CO_TEACHER) may call the endpoint for any student in
their classroom; a member may only call it for their own profile.

---

## 2. Decisions (signed off)

| # | Decision |
|---|---|
| 1 | One endpoint returns the whole screen (chart series + summary + banner inputs). |
| 2 | Chart x-axis is **per exercise** — one point per graded submission, ordered `submitted_dt ASC`. No DAY/WEEK/MONTH bucket. |
| 3 | "Tỷ lệ làm đúng" = **số bài đạt / số bài đã chấm** = `passed_count / graded_count`. |
| 4 | "Đạt" threshold = `score_percentage >= 50` (i.e. ≥ 5.0/10). Single const `PassScorePctThreshold = 50`. |
| 5 | All exercise purposes counted by default; optional `purpose` request filter (`HOMEWORK` / `QUIZ` / `EXAM`) narrows it. |
| 6 | "Giai đoạn trước" delta = current avg − prior avg, where prior = the immediately-preceding window of the **same length** (`[from−(to−from), from−1µs]`). 10-point scale. |
| 7 | Trend label reuses the `001` algorithm: `LinearSlope` (least-squares) over the per-exercise 10-point score series → `Classify(count, avg10pt, slope)` → `enum.ProgressComment`. |
| 8 | Score wire format: both raw `score_percentage` (0–100, int) and 10-point (`PctTo10Pt`, 2 dp). Client rounds for display. |
| 9 | Only graded submissions count: `submitted_dt IS NOT NULL AND score_percentage IS NOT NULL`, active rows only. |
| 10 | `tz` accepted in body (default `+07:00`); validated as numeric offset, passed verbatim to `CONVERT_TZ` if ever needed. Date-range filter is on `submitted_dt`. |
| 11 | Validator rejects `from_dt >= to_dt` and `to_dt − from_dt > 2 years`. |
| 12 | Exercise title fallback: `#<classroom_exercise_id>` when the title row can't be resolved. |
| 13 | Endpoint is unpaginated (a single student's submissions in a range is small). |

---

## 3. Endpoint

### `POST /classrooms/progress/profile`

**Request**
```json
{
  "profile_id": 1001,            // acting profile (logged-in parent/student)
  "classroom_id": 5001,
  "target_profile_id": 7001,     // student to view; null/0 => self
  "from_dt": "2026-05-01 00:00:00",
  "to_dt":   "2026-06-24 23:59:59",
  "purpose": null,               // null=all; HOMEWORK|QUIZ|EXAM
  "tz": "+07:00"
}
```

**Response**
```json
{
  "mstatus": 200,
  "target_profile_id": 7001,
  "from_dt": "2026-05-01 00:00:00",
  "to_dt":   "2026-06-24 23:59:59",
  "tz": "+07:00",
  "purpose": null,
  "series": [
    {
      "exercise_id": 9001,
      "title": "Bài 1",
      "submitted_dt": "2026-05-05 10:00:00",
      "score": 6.0,  "score_pct": 60,
      "correct_number": 6, "total_questions": 10,
      "passed": true
    }
    // … one entry per graded submission, submitted_dt ASC
  ],
  "summary": {
    "total_exercises":      6,
    "graded_count":         6,
    "average_score":        7.7,  "average_score_pct": 77.0,
    "average_delta":        1.2,                 // null when no prior-period data
    "highest_score":        10.0, "highest_score_pct": 100,
    "highest_exercise_id":  9006, "highest_exercise_title": "Bài 6",
    "highest_submitted_dt": "2026-06-24 09:00:00",
    "passed_count":         5,
    "correct_rate":         0.83,                // passed_count / graded_count
    "trend":                "PROGRESS"           // enum.ProgressComment
  }
}
```

Empty range → `series: []`, `summary` with zero counts, `average_score`,
`average_delta`, `highest_*` all `null`, `trend: "NO_DATA"`.

### Field → UI mapping

| Screen element | Response field |
|---|---|
| Line chart points (Bài N, ngày, điểm) | `series[].title` / `submitted_dt` / `score` |
| "6 bài đánh giá" | `summary.total_exercises` |
| Điểm trung bình `7.7/10` | `summary.average_score` |
| `↑ 1.2 so với giai đoạn trước` | `summary.average_delta` |
| Điểm cao nhất `10/10`, `Bài 6 - 24/06` | `summary.highest_score` / `highest_exercise_title` / `highest_submitted_dt` |
| Tỷ lệ làm đúng `83%`, `5/6 bài` | `summary.correct_rate` / `passed_count` / `graded_count` |
| Xu hướng `Tiến bộ` + banner | `summary.trend` (+ `average_delta`) |

---

## 4. Trend enum (reused)

`enum.ProgressComment` (re-add `internal/shared/enum/classroom_progress.go`,
the subset this feature needs):

| Symbol | UI label (VN) | Rule |
|---|---|---|
| `NO_DATA` | (no chip) | `graded_count == 0` |
| `INSUFFICIENT` | (no chip) | `graded_count == 1` |
| `NEED_TO_TRY` | Cần cố gắng | `avg10pt < 5.0` OR `slope ≤ -0.05` |
| `GOOD_PROGRESS` | Tiến bộ tốt | `avg10pt ≥ 7.5` AND `slope ≥ 0` |
| `PROGRESS` | Tiến bộ | everything else with `graded_count ≥ 2` |

First-match-wins; identical to `001` §4. `slope` here is pts-per-exercise
(x = exercise index 0..n-1), not pts-per-day — the only deviation from
`001`, and acceptable because this view is per-exercise by design.

Banner copy is client-side: derive from `trend` + sign of `average_delta`.

---

## 5. Permission rule

Single helper `requireProfileProgressAccess`:

```
uid    := sessionUID(r)
caller := resolveActingProfile(ctx, req.ProfileID, uid)
target := req.TargetProfileID (default = caller.ProfileId() when null/0)

if target == caller.ProfileId():
    requireMember(ctx, req.ClassroomID, caller.ProfileId())   // ACTIVE member, any role
else:
    requireManager(ctx, req.ClassroomID, caller.ProfileId())  // OWNER or CO_TEACHER
```

Reuses `resolveActingProfile` / `requireMember` / `requireManager` from
`internal/module/classroom/permission.go`.

Rejection codes: `UNAUTHORIZED`, `PROFILE_NOT_FOUND`,
`CLASSROOM_NOT_FOUND`, `CLASSROOM_PERMISSION_DENIED`.

---

## 6. Status codes

Re-add the 12200-block progress codes removed with `001` (last live
classroom code is `CLASSROOM_PROGRAM_DUPLICATE = 12224`):

| Code | Symbol | EN | VN |
|---|---|---|---|
| 12225 | `CLASSROOM_PROGRESS_INVALID_DATE_RANGE` | "Invalid date range." | "Khoảng thời gian không hợp lệ." |
| 12226 | `CLASSROOM_PROGRESS_INVALID_TZ` | "Invalid time zone." | "Múi giờ không hợp lệ." |
| 12227 | `CLASSROOM_PROGRESS_INVALID_PURPOSE` | "Invalid exercise type." | "Loại bài không hợp lệ." |

Lockstep across `code.go`, `message_en.go`, `message_vn.go`.

Re-used: `UNAUTHORIZED`, `PROFILE_NOT_FOUND`, `CLASSROOM_NOT_FOUND`,
`CLASSROOM_PERMISSION_DENIED`, `CLASSROOM_MISSING_ID`, `FAIL`.

> Note: this endpoint has no `bucket` param, so the old
> `CLASSROOM_PROGRESS_INVALID_BUCKET` is NOT reintroduced.

---

## 7. Data source

Table: `ma_exercise_submissions` (domain `exercise.Submission`,
repo `exercise.ISubmissionRepository`). Titles + purpose come from
`ma_exercises` via `exercise.IRepository.ListByClassroomExerciseIds`.

> ⚠️ Dependency: the `ma_exercise_submissions` **table migration is not
> checked into `/migrations/` on this branch** (highest file is
> `019_ma_exercises_table.sql`). The domain/repo assume the table exists
> (applied manually). Confirm the table + its base columns exist on every
> target DB before shipping. See `IMPLEMENTATION-PLAN.md` Phase 1.

Active predicate (reuses `exerciseSubmissionActiveWhere` +
`exerciseSubmissionActiveArgs()`):
```
status = 'ACTIVE'
AND deleted_dt IS NULL
AND (submission_status IS NULL OR submission_status != 'DELETED')
AND classroom_id = ?
AND profile_id   = ?
AND submitted_dt BETWEEN ? AND ?
AND submitted_dt IS NOT NULL
AND score_percentage IS NOT NULL
ORDER BY submitted_dt ASC
```

Recommended index (migration, optional but advised):
```sql
ALTER TABLE ma_exercise_submissions
  ADD KEY ix_classroom_profile_submitted (classroom_id, profile_id, submitted_dt);
```

Purpose filtering is applied **in Go** after hydrating exercise rows
(a single student's submission set is small), so no JOIN is needed in SQL.

---

## 8. Out of scope (v1)

- No pre-aggregation; the query hits the raw submission table directly.
- No background job.
- No CSV/export (the screen's top-right ↓ icon is a client concern).
- No revival of the teacher-facing `001` two-endpoint dashboard.
- No per-question breakdown — `correct_number`/`total_questions` are used
  only for the pass test and surfaced raw per point.

---

## 9. Reuse map (recover from git `6144d38`)

| Recovered file | Use here |
|---|---|
| `internal/shared/enum/classroom_progress.go` | `ProgressComment`, `DefaultProgressTz`, `IsValidTzOffset`. Drop `ProgressBucket` / `TrendSource` (unused). |
| `internal/application/query/classroomprogress/trend.go` | `PctTo10Pt`, `LinearSlope`, `Classify`. Drop the bucket/day-window helpers. |

`internal/module/classroom/errors.go` already carries leftover
`ErrProgressInvalid*` vars from `001`; reuse `ErrProgressInvalidDateRange`
/ `ErrProgressInvalidTz`, add `ErrProgressInvalidPurpose`, drop
`ErrProgressInvalidBucket` (or leave it — harmless).
